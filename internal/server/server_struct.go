package server

import (
	"net/http"

	"project/internal/repositories"
	"project/internal/services"

	"github.com/go-chi/chi/v5"
)

// Server representa el servidor HTTP y sus dependencias.
type Server struct {
	router     *chi.Mux
	repo       repositories.ItemRepository
	service    services.ItemService
	httpServer *http.Server
}
