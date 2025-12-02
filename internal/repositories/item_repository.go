package repositories

import (
	"context"
	"project/internal/models"
)

// ItemRepository define la interfaz para el acceso a datos de Items.
// Esta abstracción permite desacoplar la lógica de negocio de la capa de persistencia,
// facilitando pruebas unitarias, mantenibilidad y la posibilidad de intercambiar
// implementaciones.
// Sigue el patrón Repository y aplica el Principio de Inversión de Dependencias.
type ItemRepository interface {
	// GetAll obtiene todos los items almacenados en el repositorio.
	GetAll(ctx context.Context) ([]models.Item, error)

	// GetByID busca un item por su identificador único (ID).
	GetByID(ctx context.Context, id int64) (*models.Item, error)

	// GetByIDs obtiene múltiples items a partir de una lista de IDs.
	GetByIDs(ctx context.Context, ids []int64) ([]models.Item, error)

	// Seed inicializa la base de datos con datos de prueba o datos por defecto.
	Seed(ctx context.Context) error

	// Close cierra la conexión o recursos asociados al repositorio.
	// Esto es esencial para liberar recursos del sistema o conexiones abiertas.
	Close() error
}
