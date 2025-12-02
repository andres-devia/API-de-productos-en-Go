package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"project/internal/errors"
	"project/internal/models"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockItemService es una implementación mock de ItemService para pruebas
// Esto nos permite probar los handlers sin implementar realmente los servicios
type MockItemService struct {
	mock.Mock
}

func (m *MockItemService) GetAllItems(ctx context.Context) ([]models.Item, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.Item), args.Error(1)
}

func (m *MockItemService) GetItemByID(ctx context.Context, id int64) (*models.Item, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Item), args.Error(1)
}

func (m *MockItemService) CompareItems(ctx context.Context, itemIDs []int64) (*models.CompareResponse, error) {
	args := m.Called(ctx, itemIDs)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.CompareResponse), args.Error(1)
}

// setupChiRouter crea un router chi real con el handler inyectado para tests más robustos
func setupChiRouter(t *testing.T, handler *ItemHandler) *chi.Mux {
	r := chi.NewRouter()
	r.Route("/api/v1", func(r chi.Router) {
		r.Route("/items", func(r chi.Router) {
			r.Get("/", handler.GetAllItems)
			r.Get("/{id}", handler.GetItemByID)
			r.Post("/compare", handler.CompareItems)
		})
	})
	return r
}

// TestGetAllItems_OK: Happy path (200), valida el body del JSON
func TestGetAllItems_OK(t *testing.T) {
	mockService := new(MockItemService)
	handler := NewItemHandler(mockService)
	router := setupChiRouter(t, handler)

	items := []models.Item{
		{
			ID:          1,
			Name:        "Test Item 1",
			Price:       100.0,
			Rating:      4.5,
			Description: "Description 1",
			ImageURL:    "https://example.com/image1.jpg",
			Specifications: models.Specifications{
				"color": "red",
			},
		},
		{
			ID:          2,
			Name:        "Test Item 2",
			Price:       200.0,
			Rating:      4.0,
			Description: "Description 2",
			ImageURL:    "https://example.com/image2.jpg",
			Specifications: models.Specifications{
				"color": "blue",
			},
		},
	}

	mockService.On("GetAllItems", mock.Anything).Return(items, nil)

	req := httptest.NewRequest("GET", "/api/v1/items", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

	var response []models.Item
	err := json.NewDecoder(w.Body).Decode(&response)
	assert.NoError(t, err)
	assert.Len(t, response, 2)
	assert.Equal(t, items[0].ID, response[0].ID)
	assert.Equal(t, items[0].Name, response[0].Name)
	assert.Equal(t, items[1].ID, response[1].ID)
	assert.Equal(t, items[1].Name, response[1].Name)

	mockService.AssertExpectations(t)
}

// TestGetAllItems_EmptyList: Cuando no hay items, devuelve lista vacía
func TestGetAllItems_EmptyList(t *testing.T) {
	mockService := new(MockItemService)
	handler := NewItemHandler(mockService)
	router := setupChiRouter(t, handler)

	mockService.On("GetAllItems", mock.Anything).Return([]models.Item{}, nil)

	req := httptest.NewRequest("GET", "/api/v1/items", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

	var response []models.Item
	err := json.NewDecoder(w.Body).Decode(&response)
	assert.NoError(t, err)
	assert.Empty(t, response)

	mockService.AssertExpectations(t)
}

// TestGetAllItems_InternalError: El servicio devuelve un error genérico (500), valida el ErrorResponse
func TestGetAllItems_InternalError(t *testing.T) {
	mockService := new(MockItemService)
	handler := NewItemHandler(mockService)
	router := setupChiRouter(t, handler)

	domainErr := errors.NewInternalServerError("error al obtener los items", nil)
	mockService.On("GetAllItems", mock.Anything).Return(nil, domainErr)

	req := httptest.NewRequest("GET", "/api/v1/items", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

	var errorResp errors.ErrorResponse
	err := json.NewDecoder(w.Body).Decode(&errorResp)
	assert.NoError(t, err)
	assert.True(t, errorResp.Error)
	assert.Equal(t, errors.ErrorCodeInternalServer, errorResp.Code)

	mockService.AssertExpectations(t)
}

// TestGetItemByID_OK: Happy path (200), valida el body del JSON
func TestGetItemByID_OK(t *testing.T) {
	mockService := new(MockItemService)
	handler := NewItemHandler(mockService)
	router := setupChiRouter(t, handler)

	expectedItem := &models.Item{
		ID:          1,
		Name:        "Test Item",
		Price:       100.0,
		Rating:      4.5,
		Description: "Test Description",
		ImageURL:    "https://example.com/image.jpg",
		Specifications: models.Specifications{
			"color": "red",
			"size":  "large",
		},
	}

	mockService.On("GetItemByID", mock.Anything, int64(1)).Return(expectedItem, nil)

	req := httptest.NewRequest("GET", "/api/v1/items/1", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

	var response models.Item
	err := json.NewDecoder(w.Body).Decode(&response)
	assert.NoError(t, err)
	assert.Equal(t, expectedItem.ID, response.ID)
	assert.Equal(t, expectedItem.Name, response.Name)
	assert.Equal(t, expectedItem.Price, response.Price)
	assert.Equal(t, expectedItem.Rating, response.Rating)

	mockService.AssertExpectations(t)
}

// TestGetItemByID_NotFound: El servicio devuelve NewNotFoundError (404), valida el ErrorResponse
func TestGetItemByID_NotFound(t *testing.T) {
	mockService := new(MockItemService)
	handler := NewItemHandler(mockService)
	router := setupChiRouter(t, handler)

	domainErr := errors.NewNotFoundError("Item", 999)
	mockService.On("GetItemByID", mock.Anything, int64(999)).Return(nil, domainErr)

	req := httptest.NewRequest("GET", "/api/v1/items/999", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
	assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

	var errorResp errors.ErrorResponse
	err := json.NewDecoder(w.Body).Decode(&errorResp)
	assert.NoError(t, err)
	assert.True(t, errorResp.Error)
	assert.Equal(t, errors.ErrorCodeNotFound, errorResp.Code)
	assert.Contains(t, errorResp.Message, "Item")
	assert.Contains(t, errorResp.Message, "999")

	mockService.AssertExpectations(t)
}

// TestGetItemByID_InvalidID: El handler recibe un ID no numérico (ej. "abc").
// Debe devolver 400 (BAD_REQUEST) sin llamar al servicio
func TestGetItemByID_InvalidID(t *testing.T) {
	mockService := new(MockItemService)
	handler := NewItemHandler(mockService)
	router := setupChiRouter(t, handler)

	req := httptest.NewRequest("GET", "/api/v1/items/abc", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

	var errorResp errors.ErrorResponse
	err := json.NewDecoder(w.Body).Decode(&errorResp)
	assert.NoError(t, err)
	assert.True(t, errorResp.Error)
	assert.Equal(t, errors.ErrorCodeBadRequest, errorResp.Code)
	assert.Contains(t, errorResp.Message, "formato de id")

	// Verificar que el servicio NO fue llamado
	mockService.AssertExpectations(t)
}

// TestGetItemByID_InternalError: El servicio devuelve un error genérico (500), valida el ErrorResponse
func TestGetItemByID_InternalError(t *testing.T) {
	mockService := new(MockItemService)
	handler := NewItemHandler(mockService)
	router := setupChiRouter(t, handler)

	domainErr := errors.NewInternalServerError("error al obtener el item", nil)
	mockService.On("GetItemByID", mock.Anything, int64(1)).Return(nil, domainErr)

	req := httptest.NewRequest("GET", "/api/v1/items/1", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

	var errorResp errors.ErrorResponse
	err := json.NewDecoder(w.Body).Decode(&errorResp)
	assert.NoError(t, err)
	assert.True(t, errorResp.Error)
	assert.Equal(t, errors.ErrorCodeInternalServer, errorResp.Code)

	mockService.AssertExpectations(t)
}

// TestCompareItems_OK: Happy path (200), valida el CompareResponse
func TestCompareItems_OK(t *testing.T) {
	mockService := new(MockItemService)
	handler := NewItemHandler(mockService)
	router := setupChiRouter(t, handler)

	expectedResponse := &models.CompareResponse{
		Items: []models.Item{
			{ID: 1, Name: "Item 1", Price: 100.0, Rating: 4.5, Specifications: models.Specifications{"color": "red"}},
			{ID: 2, Name: "Item 2", Price: 200.0, Rating: 4.0, Specifications: models.Specifications{"color": "blue"}},
		},
		Comparison: models.ComparisonDetails{
			PriceRange:  models.PriceRange{Min: 100.0, Max: 200.0},
			RatingRange: models.RatingRange{Min: 4.0, Max: 4.5},
			CommonSpecs: []string{},
			UniqueSpecs: map[int64][]string{1: {"color"}, 2: {"color"}},
		},
	}

	mockService.On("CompareItems", mock.Anything, []int64{1, 2}).Return(expectedResponse, nil)

	requestBody := models.CompareRequest{ItemIDs: []int64{1, 2}}
	bodyBytes, _ := json.Marshal(requestBody)

	req := httptest.NewRequest("POST", "/api/v1/items/compare", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

	var response models.CompareResponse
	err := json.NewDecoder(w.Body).Decode(&response)
	assert.NoError(t, err)
	assert.Len(t, response.Items, 2)
	assert.Equal(t, expectedResponse.Comparison.PriceRange.Min, response.Comparison.PriceRange.Min)
	assert.Equal(t, expectedResponse.Comparison.PriceRange.Max, response.Comparison.PriceRange.Max)

	mockService.AssertExpectations(t)
}

// TestCompareItems_BadRequest_MalformedJSON: Envía un JSON inválido. Debe devolver 400 (BAD_REQUEST)
func TestCompareItems_BadRequest_MalformedJSON(t *testing.T) {
	mockService := new(MockItemService)
	handler := NewItemHandler(mockService)
	router := setupChiRouter(t, handler)

	req := httptest.NewRequest("POST", "/api/v1/items/compare", bytes.NewReader([]byte(`{"item_ids": [1, 2`)))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

	var errorResp errors.ErrorResponse
	err := json.NewDecoder(w.Body).Decode(&errorResp)
	assert.NoError(t, err)
	assert.True(t, errorResp.Error)
	assert.Equal(t, errors.ErrorCodeBadRequest, errorResp.Code)
	assert.Contains(t, errorResp.Message, "cuerpo de la petición")

	// Verificar que el servicio NO fue llamado
	mockService.AssertExpectations(t)
}

// TestCompareItems_ValidationError: El servicio devuelve NewValidationError
// (ej. < 2 IDs). Debe devolver 422 (VALIDATION_ERROR)
func TestCompareItems_ValidationError(t *testing.T) {
	mockService := new(MockItemService)
	handler := NewItemHandler(mockService)
	router := setupChiRouter(t, handler)

	domainErr := errors.NewValidationError("se requieren al menos 2 items para comparar", nil)
	mockService.On("CompareItems", mock.Anything, []int64{1}).Return(nil, domainErr)

	requestBody := models.CompareRequest{ItemIDs: []int64{1}}
	bodyBytes, _ := json.Marshal(requestBody)

	req := httptest.NewRequest("POST", "/api/v1/items/compare", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
	assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

	var errorResp errors.ErrorResponse
	err := json.NewDecoder(w.Body).Decode(&errorResp)
	assert.NoError(t, err)
	assert.True(t, errorResp.Error)
	assert.Equal(t, errors.ErrorCodeValidation, errorResp.Code)

	mockService.AssertExpectations(t)
}

// TestCompareItems_NotFound: El servicio devuelve NewNotFoundError
// (ej. un ID no existe). Debe devolver 404 (NOT_FOUND)
func TestCompareItems_NotFound(t *testing.T) {
	mockService := new(MockItemService)
	handler := NewItemHandler(mockService)
	router := setupChiRouter(t, handler)

	domainErr := errors.NewNotFoundError("Items con IDs [999]", nil)
	mockService.On("CompareItems", mock.Anything, []int64{1, 999}).Return(nil, domainErr)

	requestBody := models.CompareRequest{ItemIDs: []int64{1, 999}}
	bodyBytes, _ := json.Marshal(requestBody)

	req := httptest.NewRequest("POST", "/api/v1/items/compare", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
	assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

	var errorResp errors.ErrorResponse
	err := json.NewDecoder(w.Body).Decode(&errorResp)
	assert.NoError(t, err)
	assert.True(t, errorResp.Error)
	assert.Equal(t, errors.ErrorCodeNotFound, errorResp.Code)

	mockService.AssertExpectations(t)
}

// TestCompareItems_BodyTooLarge: Simula un body que excede http.MaxBytesReader
func TestCompareItems_BodyTooLarge(t *testing.T) {
	mockService := new(MockItemService)
	handler := NewItemHandler(mockService)
	router := setupChiRouter(t, handler)

	// Crear un body que exceda el límite de 1MB
	largeBody := make([]byte, 2*1024*1024) // 2MB
	for i := range largeBody {
		largeBody[i] = 'a'
	}

	req := httptest.NewRequest("POST", "/api/v1/items/compare", bytes.NewReader(largeBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	// http.MaxBytesReader debería causar un error al leer el body
	assert.Equal(t, http.StatusBadRequest, w.Code)

	var errorResp errors.ErrorResponse
	err := json.NewDecoder(w.Body).Decode(&errorResp)
	assert.NoError(t, err)
	assert.True(t, errorResp.Error)

	// Verificar que el servicio NO fue llamado
	mockService.AssertExpectations(t)
}
