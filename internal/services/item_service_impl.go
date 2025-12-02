package services

import (
	"context"
	stdErrors "errors"
	"fmt"
	"project/internal/errors"
	"project/internal/models"
	"project/internal/repositories"
	"sort"
)

// ItemServiceImpl implementa la interfaz ItemService.
// Esta capa representa la lógica de negocio y orquesta
// las llamadas hacia el repositorio.
type ItemServiceImpl struct {
	repo repositories.ItemRepository
}

// NewItemService crea una nueva instancia del servicio.
func NewItemService(repo repositories.ItemRepository) ItemService {
	return &ItemServiceImpl{repo: repo}
}

// GetAllItems obtiene todos los ítems desde el repositorio.
// Si algo falla, envía un error de servidor interno.
func (s *ItemServiceImpl) GetAllItems(ctx context.Context) ([]models.Item, error) {
	items, err := s.repo.GetAll(ctx)
	if err != nil {
		return nil, errors.NewInternalServerError("error al obtener los items", err)
	}
	return items, nil
}

// GetItemByID devuelve un ítem según su ID.
// Si no existe, retorna un error de tipo NotFound.
func (s *ItemServiceImpl) GetItemByID(ctx context.Context, id int64) (*models.Item, error) {
	if id <= 0 {
		return nil, errors.NewValidationError("ID inválido", nil)
	}
	item, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if stdErrors.Is(err, repositories.ErrNotFound) {
			return nil, errors.NewNotFoundError("Item", id)
		}

		return nil, errors.NewInternalServerError("error al obtener el item", err)
	}

	return item, nil
}

// CompareItems compara múltiples ítems y genera un informe
// con rangos de precio, rating y especificaciones comunes/únicas.
func (s *ItemServiceImpl) CompareItems(ctx context.Context, itemIDs []int64) (*models.CompareResponse, error) {
	// Validación de reglas del negocio
	if len(itemIDs) < 2 {
		return nil, errors.NewValidationError("se requieren al menos 2 items para comparar", nil)
	}

	if len(itemIDs) > 10 {
		return nil, errors.NewValidationError("máximo 10 items pueden compararse a la vez", nil)
	}

	itemIDs = uniqueIDs(itemIDs)

	items, err := s.repo.GetByIDs(ctx, itemIDs)
	if err != nil {
		return nil, errors.NewInternalServerError("error al obtener los items para comparación", err)
	}

	if len(items) != len(itemIDs) {
		missingIDs := missingItemIDs(itemIDs, items)
		return nil, errors.NewNotFoundError(fmt.Sprintf("Items con IDs %v no encontrados", missingIDs), nil)
	}

	comparison := s.generateComparison(items)

	return &models.CompareResponse{
		Items:      items,
		Comparison: comparison,
	}, nil
}

// generateComparison construye los datos derivados necesarios
// para la respuesta de comparación.
func (s *ItemServiceImpl) generateComparison(items []models.Item) models.ComparisonDetails {
	if len(items) == 0 {
		return models.ComparisonDetails{}
	}

	prices := make([]float64, len(items))
	ratings := make([]float64, len(items))
	for i, item := range items {
		prices[i] = item.Price
		ratings[i] = item.Rating
	}
	sort.Float64s(prices)
	sort.Float64s(ratings)

	priceRange := models.PriceRange{Min: prices[0], Max: prices[len(prices)-1]}
	ratingRange := models.RatingRange{Min: ratings[0], Max: ratings[len(ratings)-1]}

	commonSpecs := s.findCommonSpecs(items)
	uniqueSpecs := s.findUniqueSpecs(items)

	return models.ComparisonDetails{
		PriceRange:  priceRange,
		RatingRange: ratingRange,
		CommonSpecs: commonSpecs,
		UniqueSpecs: uniqueSpecs,
	}
}

// findCommonSpecs retorna las especificaciones (keys)
// presentes en todos los ítems.
func (s *ItemServiceImpl) findCommonSpecs(items []models.Item) []string {
	if len(items) == 0 {
		return []string{}
	}

	specCounts := make(map[string]int)
	totalItems := len(items)

	for _, item := range items {
		for specKey := range item.Specifications {
			specCounts[specKey]++
		}
	}

	common := make([]string, 0, len(specCounts))
	for specKey, count := range specCounts {
		if count == totalItems {
			common = append(common, specKey)
		}
	}

	sort.Strings(common)
	return common
}

// findUniqueSpecs identifica las especificaciones únicas por item
func (s *ItemServiceImpl) findUniqueSpecs(items []models.Item) map[int64][]string {
	unique := make(map[int64][]string)
	specMap := make(map[string]map[int64]bool)

	for _, item := range items {
		for key := range item.Specifications {
			if specMap[key] == nil {
				specMap[key] = make(map[int64]bool)
			}
			specMap[key][item.ID] = true
		}
	}

	for spec, ids := range specMap {
		if len(ids) == 1 {
			for id := range ids {
				unique[id] = append(unique[id], spec)
			}
		}
	}

	for id := range unique {
		sort.Strings(unique[id])
	}

	return unique
}

// uniqueIDs elimina IDs duplicados de la lista
func uniqueIDs(ids []int64) []int64 {
	seen := make(map[int64]bool)
	result := []int64{}
	for _, id := range ids {
		if !seen[id] {
			seen[id] = true
			result = append(result, id)
		}
	}
	return result
}

// missingItemIDs retorna los IDs que no se encontraron en los items recuperados
func missingItemIDs(requested []int64, found []models.Item) []int64 {
	foundMap := make(map[int64]bool)
	for _, item := range found {
		foundMap[item.ID] = true
	}
	var missing []int64
	for _, id := range requested {
		if !foundMap[id] {
			missing = append(missing, id)
		}
	}
	return missing
}
