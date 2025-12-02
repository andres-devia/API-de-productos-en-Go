package sqlite

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"project/internal/models"
	"project/internal/repositories"
)

// GetAll recupera todos los items almacenados en la base de datos.
// Ejecuta una consulta SQL, escanea los resultados y convierte el JSON
// almacenado en la columna `specifications` a un mapa Go.
func (r *SQLiteItemRepository) GetAll(ctx context.Context) ([]models.Item, error) {
	query := `
		SELECT id, name, image_url, description, price, rating, specifications
		FROM items
		ORDER BY id
	`

	rows, err := r.DB.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query items: %w", err)
	}
	defer rows.Close()

	var items []models.Item

	for rows.Next() {
		var item models.Item
		var specsJSON string

		if err := rows.Scan(
			&item.ID,
			&item.Name,
			&item.ImageURL,
			&item.Description,
			&item.Price,
			&item.Rating,
			&specsJSON,
		); err != nil {
			return nil, fmt.Errorf("failed to scan item: %w", err)
		}

		if err := json.Unmarshal([]byte(specsJSON), &item.Specifications); err != nil {
			return nil, fmt.Errorf("failed to unmarshal specifications: %w", err)
		}

		items = append(items, item)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return items, nil
}

// GetByID obtiene un item específico buscándolo por su ID.
// Retorna nil si no se encuentra un registro con el ID dado.
// Si existe, convierte el JSON de specifications y devuelve el item completo.
func (r *SQLiteItemRepository) GetByID(ctx context.Context, id int64) (*models.Item, error) {
	query := `
		SELECT id, name, image_url, description, price, rating, specifications
		FROM items
		WHERE id = ?
	`

	var item models.Item
	var specsJSON string

	err := r.DB.QueryRowContext(ctx, query, id).Scan(
		&item.ID,
		&item.Name,
		&item.ImageURL,
		&item.Description,
		&item.Price,
		&item.Rating,
		&specsJSON,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			// Traducimos el error de la DB a un error de repositorio
			return nil, repositories.ErrNotFound
		}
		return nil, fmt.Errorf("error al consultar item: %w", err)
	}

	if err := json.Unmarshal([]byte(specsJSON), &item.Specifications); err != nil {
		return nil, fmt.Errorf("error al deserializar las especificaciones: %w", err)
	}

	return &item, nil
}

// GetByIDs obtiene múltiples items usando una lista de IDs.
// Construye dinámicamente los placeholders para una clausula IN
// y ejecuta una consulta parametrizada evitando SQL injection.
// Retorna un slice de items o error si falla la consulta, lectura de filas
// o parseo del JSON.
func (r *SQLiteItemRepository) GetByIDs(ctx context.Context, ids []int64) ([]models.Item, error) {
	if len(ids) == 0 {
		return []models.Item{}, nil
	}

	placeholders := ""
	for i := range ids {
		if i > 0 {
			placeholders += ","
		}
		placeholders += "?"
	}

	query := fmt.Sprintf(`
		SELECT id, name, image_url, description, price, rating, specifications
		FROM items
		WHERE id IN (%s)
		ORDER BY id
	`, placeholders)

	args := make([]interface{}, len(ids))
	for i, id := range ids {
		args[i] = id
	}

	rows, err := r.DB.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("error al consultar los items: %w", err)
	}
	defer rows.Close()

	var items []models.Item

	for rows.Next() {
		var item models.Item
		var specsJSON string

		if err := rows.Scan(
			&item.ID,
			&item.Name,
			&item.ImageURL,
			&item.Description,
			&item.Price,
			&item.Rating,
			&specsJSON,
		); err != nil {
			return nil, fmt.Errorf("error al escanear el ítem: %w", err)
		}

		if err := json.Unmarshal([]byte(specsJSON), &item.Specifications); err != nil {
			return nil, fmt.Errorf("error al deserializar las especificaciones: %w", err)
		}

		items = append(items, item)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error al iterar las filas: %w", err)
	}

	return items, nil
}
