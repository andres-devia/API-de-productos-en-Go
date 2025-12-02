package services

import (
	"context"
	"project/internal/models"
)

// ItemService define la interfaz que representa la lógica de negocio
// relacionada con los ítems. Esta capa actúa como intermediaria entre
// los controladores (handlers) y la capa de persistencia (repositories).
type ItemService interface {
	// GetAllItems obtiene todos los ítems disponibles en el sistema.
	// Recibe un contexto para controlar tiempos de ejecución o cancelaciones.
	GetAllItems(ctx context.Context) ([]models.Item, error)

	// GetItemByID obtiene un ítem por su identificador único.
	// Retorna un puntero a Item si existe, o un error si no se encuentra.
	GetItemByID(ctx context.Context, id int64) (*models.Item, error)

	// CompareItems recibe una lista de IDs y devuelve una estructura
	// con la información necesaria para comparar esos ítems.
	// Si algún ID no existe, devuelve un error.
	CompareItems(ctx context.Context, itemIDs []int64) (*models.CompareResponse, error)
}
