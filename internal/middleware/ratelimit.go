package middleware

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"project/internal/errors"

	"golang.org/x/time/rate"
)

// RateLimiter gestiona el límite de tasa por IP de cliente
type RateLimiter struct {
	// clients almacena los limitadores de tasa por IP de cliente
	clients            map[string]*clientLimiter
	mu                 sync.RWMutex
	rateLimit          int
	timeWindow         time.Duration
	cleanupInterval    time.Duration
	limiterEvictionAge time.Duration
	stopCleanup        context.CancelFunc
}

// clientLimiter envuelve un rate.Limiter junto con la hora del último acceso
type clientLimiter struct {
	limiter    *rate.Limiter
	lastAccess time.Time
}

// NewRateLimiter crea una nueva instancia de RateLimiter.
// rateLimit: número de peticiones permitidas
// timeWindow: ventana de tiempo para el límite de tasa (ej. 1*time.Minute)
func NewRateLimiter(rateLimit int, timeWindow time.Duration) *RateLimiter {
	ctx, cancel := context.WithCancel(context.Background())

	rl := &RateLimiter{
		clients:            make(map[string]*clientLimiter),
		rateLimit:          rateLimit,
		timeWindow:         timeWindow,
		cleanupInterval:    5 * time.Minute,
		limiterEvictionAge: 10 * time.Minute,
		stopCleanup:        cancel,
	}

	go rl.cleanup(ctx)

	return rl
}

// RateLimit es el middleware HTTP que aplica el límite de tasa
func (rl *RateLimiter) RateLimit(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		clientIP := rl.getClientIP(r)

		limiter := rl.getLimiter(clientIP)

		if !limiter.Allow() {
			rl.writeRateLimitError(w)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// getClientIP extrae la IP del cliente desde la petición.
// Comprueba primero el header X-Forwarded-For (para reverse proxies), luego X-Real-IP, y finalmente RemoteAddr.
func (rl *RateLimiter) getClientIP(r *http.Request) string {
	forwarded := r.Header.Get("X-Forwarded-For")
	if forwarded != "" {
		ips := strings.Split(forwarded, ",")
		if len(ips) > 0 {
			ip := strings.TrimSpace(ips[0])
			if rl.isValidIP(ip) {
				return ip
			}
		}
	}

	realIP := r.Header.Get("X-Real-IP")
	if realIP != "" && rl.isValidIP(realIP) {
		return realIP
	}

	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}

	return ip
}

// isValidIP comprueba si la dirección IP es válida
func (rl *RateLimiter) isValidIP(ip string) bool {
	parsed := net.ParseIP(strings.TrimSpace(ip))
	return parsed != nil
}

// getLimiter obtiene o crea un limitador de tasa para la IP de cliente dada.
// Esta función adquiere un bloqueo de escritura (write lock) porque puede crear una
// nueva entrada y siempre actualiza la hora de último acceso.
func (rl *RateLimiter) getLimiter(clientIP string) *rate.Limiter {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()

	if cl, exists := rl.clients[clientIP]; exists {
		cl.lastAccess = now
		return cl.limiter
	}

	ratePerSecond := float64(rl.rateLimit) / rl.timeWindow.Seconds()
	burst := rl.rateLimit

	limiter := rate.NewLimiter(rate.Limit(ratePerSecond), burst)

	rl.clients[clientIP] = &clientLimiter{
		limiter:    limiter,
		lastAccess: now,
	}

	return limiter
}

// cleanup elimina periódicamente los limitadores antiguos que no se han usado recientemente.
// Esto previene fugas de memoria por acumular limitadores de clientes inactivos.
func (rl *RateLimiter) cleanup(ctx context.Context) {
	ticker := time.NewTicker(rl.cleanupInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			rl.performCleanup()
		}
	}
}

// performCleanup elimina los limitadores que no han sido accedidos recientemente
func (rl *RateLimiter) performCleanup() {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	cutoff := time.Now().Add(-rl.limiterEvictionAge)

	for ip, cl := range rl.clients {
		if cl.lastAccess.Before(cutoff) {
			delete(rl.clients, ip)
		}
	}
}

// writeRateLimitError escribe una respuesta de error de límite de tasa estandarizada
func (rl *RateLimiter) writeRateLimitError(w http.ResponseWriter) {
	domainErr := errors.NewDomainError(
		errors.ErrorCodeTooManyRequests,
		"Límite de tasa excedido. Por favor, inténtelo de nuevo más tarde",
		nil,
	)

	statusCode := domainErr.HTTPStatus()
	errorResponse := domainErr.ToErrorResponse()

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Retry-After", rl.calculateRetryAfter())
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(errorResponse)
}

// calculateRetryAfter calcula el valor del header Retry-After.
// Devuelve la duración de la ventana de tiempo en segundos como string.
func (rl *RateLimiter) calculateRetryAfter() string {
	seconds := int(rl.timeWindow.Seconds())
	if seconds < 1 {
		seconds = 1
	}
	return fmt.Sprintf("%d", seconds)
}

// Stop detiene la gorutina de limpieza y debe llamarse durante el apagado (shutdown)
func (rl *RateLimiter) Stop() {
	rl.stopCleanup()
}
