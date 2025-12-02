package server

import (
	"time"

	"project/internal/handlers"
	customMiddleware "project/internal/middleware"
	"project/internal/services"

	"github.com/go-chi/chi/v5"
	chiMiddleware "github.com/go-chi/chi/v5/middleware"
)

// SetupRouter configura y retorna el router de Chi con todas las rutas y middlewares.
// Este diseño respeta principios de **Inyección de Dependencias**, **Responsabilidad Única (SRP)**
// y conceptos de **Arquitectura Limpia**, permitiendo que el router no dependa directamente
// de la capa de datos, sino únicamente de los servicios.
func SetupRouter(itemService services.ItemService) *chi.Mux {
	r := chi.NewRouter()

	// ----------------------------
	// Middlewares globales
	// ----------------------------
	// SecurityHeaders: añade cabeceras de seguridad como Content-Security-Policy, X-Frame-Options, etc.
	// Debe ir primero para asegurar que todas las respuestas incluyan estas cabeceras.
	r.Use(customMiddleware.SecurityHeaders)

	// CORS: habilita el intercambio de recursos entre dominios
	r.Use(customMiddleware.CORS)

	// RequestID: asigna un ID único por petición, útil para trazabilidad y debug.
	r.Use(chiMiddleware.RequestID)

	// Logger: registra cada petición HTTP con detalles como ruta, método, latencia y código de respuesta.
	r.Use(chiMiddleware.Logger)

	// Recoverer: captura cualquier panic en la ejecución de handlers y previene que el servidor colapse.
	r.Use(chiMiddleware.Recoverer)

	// RateLimiter: límite de 100 solicitudes por minuto por IP.
	// Esto protege la API contra abuso o ataques de denegación de servicio (DoS).
	rateLimiter := customMiddleware.NewRateLimiter(100, 1*time.Minute)
	r.Use(rateLimiter.RateLimit)

	// ----------------------------
	// Inicialización de handlers
	// ----------------------------
	// ItemHandler maneja las rutas relacionadas con elementos.
	// Se inyecta itemService.
	itemHandler := handlers.NewItemHandler(itemService)

	// ----------------------------
	// Definición de rutas
	// ----------------------------
	r.Route("/api/v1", func(r chi.Router) {
		// Items endpoints
		r.Route("/items", func(r chi.Router) {
			r.Get("/", itemHandler.GetAllItems)
			r.Get("/{id}", itemHandler.GetItemByID)
			r.Post("/compare", itemHandler.CompareItems)
		})
	})

	return r
}
