package handlers

import (
	"encoding/json"
	"net/http"
	"project/internal/errors"
	"project/internal/models"
	"project/internal/services"
	"strconv"

	"github.com/go-chi/chi/v5"
)

// ItemHandler maneja las peticiones HTTP para los endpoints relacionados con items.
type ItemHandler struct {
	service services.ItemService
}

// NewItemHandler crea una nueva instancia del handler de items.
func NewItemHandler(service services.ItemService) *ItemHandler {
	return &ItemHandler{
		service: service,
	}
}

// GetAllItems maneja GET /api/v1/items
// Devuelve todos los items en el sistema.
func (h *ItemHandler) GetAllItems(w http.ResponseWriter, r *http.Request) {
	items, err := h.service.GetAllItems(r.Context())
	if err != nil {
		h.handleError(w, err)
		return
	}

	h.writeJSON(w, http.StatusOK, items)
}

// GetItemByID maneja GET /api/v1/items/{id}
// Devuelve un único item por su ID.
func (h *ItemHandler) GetItemByID(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		domainErr := errors.NewBadRequestError(
			"formato de id de item inválido",
			err,
		)
		h.handleError(w, domainErr)
		return
	}

	item, err := h.service.GetItemByID(r.Context(), id)
	if err != nil {
		h.handleError(w, err)
		return
	}

	h.writeJSON(w, http.StatusOK, item)
}

// CompareItems maneja POST /api/v1/items/compare
// Recibe IDs de items y devuelve detalles de comparación.
func (h *ItemHandler) CompareItems(w http.ResponseWriter, r *http.Request) {
	const maxBodySize = 1024 * 1024
	r.Body = http.MaxBytesReader(w, r.Body, maxBodySize)

	var req models.CompareRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		domainErr := errors.NewBadRequestError(
			"cuerpo de la petición (body) inválido",
			err,
		)
		h.handleError(w, domainErr)
		return
	}

	response, err := h.service.CompareItems(r.Context(), req.ItemIDs)
	if err != nil {
		h.handleError(w, err)
		return
	}

	h.writeJSON(w, http.StatusOK, response)
}

// handleError procesa errores de dominio y escribe la respuesta HTTP apropiada.
func (h *ItemHandler) handleError(w http.ResponseWriter, err error) {
	domainErr, ok := err.(*errors.DomainError)
	if !ok {
		domainErr = errors.NewInternalServerError(
			"un error inesperado ha ocurrido",
			err,
		)
	}

	statusCode := domainErr.HTTPStatus()
	errorResponse := domainErr.ToErrorResponse()

	h.writeJSON(w, statusCode, errorResponse)
}

// writeJSON escribe una respuesta JSON con las cabeceras y código de estado correctos.
func (h *ItemHandler) writeJSON(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(data)
}
