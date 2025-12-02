package server

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"project/internal/repositories/sqlite"
	"project/internal/services"
)

// NewServer crea e inicializa una nueva instancia del servidor.
//
// Este constructor realiza los siguientes pasos:
// 1. Inicializa el repositorio SQLite, encargado de la persistencia.
// 2. Ejecuta la siembra (Seed) para cargar datos iniciales en la base de datos.
// 3. Crea el servicio de negocio (ItemService), aplicando el patrón de inyección de dependencias.
// 4. Configura el router con todas las rutas HTTP y middleware.
// 5. Construye el servidor HTTP con configuraciones de timeout apropiadas.
func NewServer(cfg Config) (*Server, error) {
	repo, err := sqlite.NewSQLiteItemRepository(cfg.DBPath)
	if err != nil {
		return nil, fmt.Errorf("error al inicializar el repositorio: %w", err)
	}

	ctx := context.Background()
	if err := repo.Seed(ctx); err != nil {
		return nil, fmt.Errorf("error al poblar la base de datos: %w", err)
	}

	service := services.NewItemService(repo)

	router := SetupRouter(service)

	httpServer := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	return &Server{
		router:     router,
		repo:       repo,
		service:    service,
		httpServer: httpServer,
	}, nil
}

// Start inicia el servidor HTTP y maneja el apagado seguro (graceful shutdown).
//
// Este método:
// 1. Escucha señales del sistema operativo (SIGINT, SIGTERM).
// 2. Inicia el servidor en una goroutine para no bloquear el flujo principal.
// 3. Cuando llega una señal de finalización, inicia un apagado controlado:
//   - Detiene nuevas conexiones.
//   - Espera hasta 10 segundos para que las conexiones activas finalicen.
//   - Cierra la base de datos de manera segura.
func (s *Server) Start() error {
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	go func() {
		log.Printf("Servidor iniciándose en %s", s.httpServer.Addr)
		if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Error al iniciar el servidor: %v", err)
		}
	}()

	<-stop
	log.Println("Deteniendo el servidor...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := s.httpServer.Shutdown(ctx); err != nil {
		return fmt.Errorf("error al detener el servidor: %w", err)
	}

	if err := s.repo.Close(); err != nil {
		return fmt.Errorf("Error al cerrar la base de datos: %w", err)
	}

	log.Println("Servidor detenido correctamente")
	return nil
}
