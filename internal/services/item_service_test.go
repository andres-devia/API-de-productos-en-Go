package services

import (
	"context"
	"database/sql"
	"project/internal/errors"
	"project/internal/models"
	"project/internal/repositories"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockItemRepository es una implementación mock de ItemRepository para pruebas
type MockItemRepository struct {
	mock.Mock
}

func (m *MockItemRepository) GetAll(ctx context.Context) ([]models.Item, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.Item), args.Error(1)
}

func (m *MockItemRepository) GetByID(ctx context.Context, id int64) (*models.Item, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Item), args.Error(1)
}

func (m *MockItemRepository) GetByIDs(ctx context.Context, ids []int64) ([]models.Item, error) {
	args := m.Called(ctx, ids)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.Item), args.Error(1)
}

func (m *MockItemRepository) Seed(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockItemRepository) Close() error {
	args := m.Called()
	return args.Error(0)
}

// TestService_GetItemByID_OK: El repo devuelve un item
func TestService_GetItemByID_OK(t *testing.T) {
	mockRepo := new(MockItemRepository)
	service := NewItemService(mockRepo)

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

	mockRepo.On("GetByID", mock.Anything, int64(1)).Return(expectedItem, nil)

	item, err := service.GetItemByID(context.Background(), 1)

	assert.NoError(t, err)
	assert.NotNil(t, item)
	assert.Equal(t, expectedItem.ID, item.ID)
	assert.Equal(t, expectedItem.Name, item.Name)
	assert.Equal(t, expectedItem.Price, item.Price)
	assert.Equal(t, expectedItem.Rating, item.Rating)

	mockRepo.AssertExpectations(t)
}

// TestService_GetAllItems_OK: Prueba el camino feliz para obtener todos los items
func TestService_GetAllItems_OK(t *testing.T) {
	mockRepo := new(MockItemRepository)
	service := NewItemService(mockRepo)

	expectedItems := []models.Item{
		{ID: 1, Name: "Item 1"},
		{ID: 2, Name: "Item 2"},
	}

	mockRepo.On("GetAll", mock.Anything).Return(expectedItems, nil)

	items, err := service.GetAllItems(context.Background())

	assert.NoError(t, err)
	assert.NotNil(t, items)
	assert.Len(t, items, 2)
	assert.Equal(t, expectedItems, items)
	mockRepo.AssertExpectations(t)
}

// TestService_GetAllItems_RepoError: Prueba un error del repositorio al obtener todos los items
func TestService_GetAllItems_RepoError(t *testing.T) {
	mockRepo := new(MockItemRepository)
	service := NewItemService(mockRepo)

	repoError := sql.ErrConnDone
	mockRepo.On("GetAll", mock.Anything).Return(nil, repoError)

	items, err := service.GetAllItems(context.Background())

	assert.Error(t, err)
	assert.Nil(t, items)

	domainErr, ok := err.(*errors.DomainError)
	assert.True(t, ok)
	assert.Equal(t, errors.ErrorCodeInternalServer, domainErr.Code)
	assert.Equal(t, repoError, domainErr.Unwrap()) // Verifica que el error original se preserva
	mockRepo.AssertExpectations(t)
}

// TestService_GetItemByID_InvalidID: Prueba con id = 0. Debe devolver ValidationError sin llamar al repo
func TestService_GetItemByID_InvalidID(t *testing.T) {
	mockRepo := new(MockItemRepository)
	service := NewItemService(mockRepo)

	item, err := service.GetItemByID(context.Background(), 0)

	assert.Error(t, err)
	assert.Nil(t, item)

	domainErr, ok := err.(*errors.DomainError)
	assert.True(t, ok)
	assert.Equal(t, errors.ErrorCodeValidation, domainErr.Code)
	assert.Contains(t, domainErr.Message, "ID inválido")

	mockRepo.AssertExpectations(t)
}

// TestService_GetItemByID_InvalidID_Negative: Prueba con id negativo
func TestService_GetItemByID_InvalidID_Negative(t *testing.T) {
	mockRepo := new(MockItemRepository)
	service := NewItemService(mockRepo)

	item, err := service.GetItemByID(context.Background(), -1)

	assert.Error(t, err)
	assert.Nil(t, item)

	domainErr, ok := err.(*errors.DomainError)
	assert.True(t, ok)
	assert.Equal(t, errors.ErrorCodeValidation, domainErr.Code)

	mockRepo.AssertExpectations(t)
}

// TestService_GetItemByID_RepoNotFound: El repo devuelve repositories.ErrNotFound.
// El servicio debe traducirlo a un errors.DomainError con código NOT_FOUND
func TestService_GetItemByID_RepoNotFound(t *testing.T) {
	mockRepo := new(MockItemRepository)
	service := NewItemService(mockRepo)

	mockRepo.On("GetByID", mock.Anything, int64(999)).Return(nil, repositories.ErrNotFound)

	item, err := service.GetItemByID(context.Background(), 999)

	assert.Error(t, err)
	assert.Nil(t, item)

	domainErr, ok := err.(*errors.DomainError)
	assert.True(t, ok)
	assert.Equal(t, errors.ErrorCodeNotFound, domainErr.Code)
	assert.Contains(t, domainErr.Message, "Item")
	assert.Contains(t, domainErr.Message, "999")

	// Verificar que el error fue traducido correctamente desde ErrNotFound
	// Nota: El servicio traduce ErrNotFound a DomainError, por lo que errors.Is
	// puede no funcionar directamente, pero el código y mensaje son correctos

	mockRepo.AssertExpectations(t)
}

// TestService_GetItemByID_RepoInternalError: El repo devuelve un error genérico (ej. sql.ErrConnDone).
// El servicio debe traducirlo a un errors.DomainError con código INTERNAL_SERVER_ERROR
func TestService_GetItemByID_RepoInternalError(t *testing.T) {
	mockRepo := new(MockItemRepository)
	service := NewItemService(mockRepo)

	repoError := sql.ErrConnDone
	mockRepo.On("GetByID", mock.Anything, int64(1)).Return(nil, repoError)

	item, err := service.GetItemByID(context.Background(), 1)

	assert.Error(t, err)
	assert.Nil(t, item)

	domainErr, ok := err.(*errors.DomainError)
	assert.True(t, ok)
	assert.Equal(t, errors.ErrorCodeInternalServer, domainErr.Code)
	assert.Contains(t, domainErr.Message, "error al obtener el item")

	// Verificar que el error subyacente se preserva
	assert.Equal(t, repoError, domainErr.Unwrap())

	mockRepo.AssertExpectations(t)
}

// TestService_CompareItems_OK: Happy path. Proporciona 3 items mockeados
// y valida que la lógica de common_specs, unique_specs y price_range es correcta
func TestService_CompareItems_OK(t *testing.T) {
	mockRepo := new(MockItemRepository)
	service := NewItemService(mockRepo)

	items := []models.Item{
		{
			ID:     1,
			Name:   "Item 1",
			Price:  100.0,
			Rating: 4.5,
			Specifications: models.Specifications{
				"color":    "red",
				"size":     "large",
				"material": "cotton", // común
			},
		},
		{
			ID:     2,
			Name:   "Item 2",
			Price:  200.0,
			Rating: 4.0,
			Specifications: models.Specifications{
				"color":    "blue",
				"size":     "medium",
				"material": "cotton", // común
			},
		},
		{
			ID:     3,
			Name:   "Item 3",
			Price:  150.0,
			Rating: 4.8,
			Specifications: models.Specifications{
				"color":    "green",
				"material": "cotton", // común
				"brand":    "Nike",   // único
			},
		},
	}

	mockRepo.On("GetByIDs", mock.Anything, []int64{1, 2, 3}).Return(items, nil)

	response, err := service.CompareItems(context.Background(), []int64{1, 2, 3})

	assert.NoError(t, err)
	assert.NotNil(t, response)
	assert.Len(t, response.Items, 3)

	// Validar price_range
	assert.Equal(t, 100.0, response.Comparison.PriceRange.Min)
	assert.Equal(t, 200.0, response.Comparison.PriceRange.Max)

	// Validar rating_range
	assert.Equal(t, 4.0, response.Comparison.RatingRange.Min)
	assert.Equal(t, 4.8, response.Comparison.RatingRange.Max)

	// Validar common_specs (debe contener "material" y "color" que están en los 3 items)
	assert.Contains(t, response.Comparison.CommonSpecs, "material")
	assert.Contains(t, response.Comparison.CommonSpecs, "color")
	assert.Len(t, response.Comparison.CommonSpecs, 2)
	// "size" NO debe estar en common_specs porque no está en el item 3
	assert.NotContains(t, response.Comparison.CommonSpecs, "size")

	// Validar unique_specs
	// Item 1: no debe tener specs únicas (todas sus specs están en otros items o son comunes)
	assert.Empty(t, response.Comparison.UniqueSpecs[int64(1)], "Item 1 no debería tener specs únicas")

	// Item 2: no debe tener specs únicas (todas sus specs están en otros items o son comunes)
	assert.Empty(t, response.Comparison.UniqueSpecs[int64(2)], "Item 2 no debería tener specs únicas")

	// Item 3: solo debe tener "brand" como spec única
	assert.Contains(t, response.Comparison.UniqueSpecs[int64(3)], "brand")
	assert.Len(t, response.Comparison.UniqueSpecs[int64(3)], 1, "Item 3 solo debería tener 'brand' como spec única")

	// "size" NO debe estar en unique_specs porque aparece en más de un item (1 y 2)
	assert.NotContains(t, response.Comparison.UniqueSpecs[int64(1)], "size")
	assert.NotContains(t, response.Comparison.UniqueSpecs[int64(2)], "size")
	assert.NotContains(t, response.Comparison.UniqueSpecs[int64(3)], "size")

	mockRepo.AssertExpectations(t)
}

// TestService_CompareItems_ValidationError_TooFew: Llama con 1 ID. Debe devolver ValidationError
func TestService_CompareItems_ValidationError_TooFew(t *testing.T) {
	mockRepo := new(MockItemRepository)
	service := NewItemService(mockRepo)

	response, err := service.CompareItems(context.Background(), []int64{1})

	assert.Error(t, err)
	assert.Nil(t, response)

	domainErr, ok := err.(*errors.DomainError)
	assert.True(t, ok)
	assert.Equal(t, errors.ErrorCodeValidation, domainErr.Code)
	assert.Contains(t, domainErr.Message, "al menos 2 items")

	// Verificar que el repo NO fue llamado
	mockRepo.AssertExpectations(t)
}

// TestService_CompareItems_ValidationError_Empty: Llama con 0 IDs
func TestService_CompareItems_ValidationError_Empty(t *testing.T) {
	mockRepo := new(MockItemRepository)
	service := NewItemService(mockRepo)

	response, err := service.CompareItems(context.Background(), []int64{})

	assert.Error(t, err)
	assert.Nil(t, response)

	domainErr, ok := err.(*errors.DomainError)
	assert.True(t, ok)
	assert.Equal(t, errors.ErrorCodeValidation, domainErr.Code)

	mockRepo.AssertExpectations(t)
}

// TestService_CompareItems_ValidationError_TooMany: Llama con 11 IDs. Debe devolver ValidationError
func TestService_CompareItems_ValidationError_TooMany(t *testing.T) {
	mockRepo := new(MockItemRepository)
	service := NewItemService(mockRepo)

	ids := make([]int64, 11)
	for i := range ids {
		ids[i] = int64(i + 1)
	}

	response, err := service.CompareItems(context.Background(), ids)

	assert.Error(t, err)
	assert.Nil(t, response)

	domainErr, ok := err.(*errors.DomainError)
	assert.True(t, ok)
	assert.Equal(t, errors.ErrorCodeValidation, domainErr.Code)
	assert.Contains(t, domainErr.Message, "máximo 10 items")

	// Verificar que el repo NO fue llamado
	mockRepo.AssertExpectations(t)
}

// TestService_CompareItems_RepoError: El repo devuelve un error. Debe devolver InternalServerError
func TestService_CompareItems_RepoError(t *testing.T) {
	mockRepo := new(MockItemRepository)
	service := NewItemService(mockRepo)

	repoError := sql.ErrConnDone
	mockRepo.On("GetByIDs", mock.Anything, []int64{1, 2}).Return(nil, repoError)

	response, err := service.CompareItems(context.Background(), []int64{1, 2})

	assert.Error(t, err)
	assert.Nil(t, response)

	domainErr, ok := err.(*errors.DomainError)
	assert.True(t, ok)
	assert.Equal(t, errors.ErrorCodeInternalServer, domainErr.Code)
	assert.Contains(t, domainErr.Message, "error al obtener los items para comparación")

	mockRepo.AssertExpectations(t)
}

// TestService_CompareItems_MissingItems: Pide 3 IDs ([1, 2, 3])
// pero el repo solo devuelve 2 ([1, 2]). Debe devolver un NotFoundError
func TestService_CompareItems_MissingItems(t *testing.T) {
	mockRepo := new(MockItemRepository)
	service := NewItemService(mockRepo)

	// El repo solo devuelve 2 items cuando se piden 3
	foundItems := []models.Item{
		{ID: 1, Name: "Item 1", Price: 100.0, Rating: 4.5},
		{ID: 2, Name: "Item 2", Price: 200.0, Rating: 4.0},
	}

	mockRepo.On("GetByIDs", mock.Anything, []int64{1, 2, 3}).Return(foundItems, nil)

	response, err := service.CompareItems(context.Background(), []int64{1, 2, 3})

	assert.Error(t, err)
	assert.Nil(t, response)

	domainErr, ok := err.(*errors.DomainError)
	assert.True(t, ok)
	assert.Equal(t, errors.ErrorCodeNotFound, domainErr.Code)
	assert.Contains(t, domainErr.Message, "Items")
	assert.Contains(t, domainErr.Message, "3") // El ID 3 no fue encontrado

	mockRepo.AssertExpectations(t)
}

// TestService_CompareItems_DuplicateIDs: Prueba que los IDs duplicados se eliminen correctamente
func TestService_CompareItems_DuplicateIDs(t *testing.T) {
	mockRepo := new(MockItemRepository)
	service := NewItemService(mockRepo)

	items := []models.Item{
		{ID: 1, Name: "Item 1", Price: 100.0, Rating: 4.5},
		{ID: 2, Name: "Item 2", Price: 200.0, Rating: 4.0},
	}

	// Se pasan IDs duplicados [1, 2, 1], pero uniqueIDs debería convertirlos a [1, 2]
	mockRepo.On("GetByIDs", mock.Anything, []int64{1, 2}).Return(items, nil)

	response, err := service.CompareItems(context.Background(), []int64{1, 2, 1})

	assert.NoError(t, err)
	assert.NotNil(t, response)
	assert.Len(t, response.Items, 2)

	mockRepo.AssertExpectations(t)
}

// TestService_CompareItems_EmptySpecs: Prueba con items que no tienen especificaciones
func TestService_CompareItems_EmptySpecs(t *testing.T) {
	mockRepo := new(MockItemRepository)
	service := NewItemService(mockRepo)

	items := []models.Item{
		{ID: 1, Name: "Item 1", Price: 100.0, Rating: 4.5, Specifications: models.Specifications{}},
		{ID: 2, Name: "Item 2", Price: 200.0, Rating: 4.0, Specifications: models.Specifications{}},
	}

	mockRepo.On("GetByIDs", mock.Anything, []int64{1, 2}).Return(items, nil)

	response, err := service.CompareItems(context.Background(), []int64{1, 2})

	assert.NoError(t, err)
	assert.NotNil(t, response)
	assert.Empty(t, response.Comparison.CommonSpecs)
	assert.Empty(t, response.Comparison.UniqueSpecs)

	mockRepo.AssertExpectations(t)
}

// TestService_CompareItems_AllCommonSpecs: Prueba cuando todos los items tienen las mismas especificaciones
func TestService_CompareItems_AllCommonSpecs(t *testing.T) {
	mockRepo := new(MockItemRepository)
	service := NewItemService(mockRepo)

	items := []models.Item{
		{
			ID:     1,
			Name:   "Item 1",
			Price:  100.0,
			Rating: 4.5,
			Specifications: models.Specifications{
				"color":    "red",
				"material": "cotton",
			},
		},
		{
			ID:     2,
			Name:   "Item 2",
			Price:  200.0,
			Rating: 4.0,
			Specifications: models.Specifications{
				"color":    "red",
				"material": "cotton",
			},
		},
	}

	mockRepo.On("GetByIDs", mock.Anything, []int64{1, 2}).Return(items, nil)

	response, err := service.CompareItems(context.Background(), []int64{1, 2})

	assert.NoError(t, err)
	assert.NotNil(t, response)
	assert.Len(t, response.Comparison.CommonSpecs, 2)
	assert.Contains(t, response.Comparison.CommonSpecs, "color")
	assert.Contains(t, response.Comparison.CommonSpecs, "material")
	assert.Empty(t, response.Comparison.UniqueSpecs)

	mockRepo.AssertExpectations(t)
}
