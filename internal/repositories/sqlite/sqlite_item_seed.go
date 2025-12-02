package sqlite

import (
	"context"
	"encoding/json"
	"fmt"
	"project/internal/models"
)

// Seed inserta datos iniciales en la base de datos si aún no existen.

// Flujo del proceso:
// 1. Verifica si la tabla items ya contiene datos.
// 2. Si está vacía, construye una lista de items de ejemplo.
// 3. Serializa el campo Specifications a JSON para almacenarlo correctamente.
// 4. Inserta cada item en la base de datos usando SQL parametrizado.
func (r *SQLiteItemRepository) Seed(ctx context.Context) error {
	// Check if data already exists
	var count int
	if err := r.DB.QueryRowContext(ctx, "SELECT COUNT(*) FROM items").Scan(&count); err != nil {
		return fmt.Errorf("failed to check existing data: %w", err)
	}

	if count > 0 {
		return nil
	}

	seedItems := []models.Item{
		{
			Name:        "MacBook Pro 16\"",
			ImageURL:    "https://example.com/images/macbook-pro.jpg",
			Description: "Powerful laptop for professionals with M2 Pro chip",
			Price:       2499.99,
			Rating:      4.8,
			Specifications: models.Specifications{
				"processor":    "Apple M2 Pro",
				"memory":       "16GB",
				"storage":      "512GB SSD",
				"display":      "16.2-inch Liquid Retina XDR",
				"graphics":     "19-core GPU",
				"battery_life": "Up to 22 hours",
				"weight":       "2.15 kg",
			},
		},
		{
			Name:        "Dell XPS 15",
			ImageURL:    "https://example.com/images/dell-xps.jpg",
			Description: "Premium Windows laptop with stunning display",
			Price:       1899.99,
			Rating:      4.6,
			Specifications: models.Specifications{
				"processor":    "Intel Core i7-13700H",
				"memory":       "32GB",
				"storage":      "1TB SSD",
				"display":      "15.6-inch OLED 3.5K",
				"graphics":     "NVIDIA RTX 4050",
				"battery_life": "Up to 10 hours",
				"weight":       "1.92 kg",
			},
		},
		{
			Name:        "HP Spectre x360",
			ImageURL:    "https://example.com/images/hp-spectre.jpg",
			Description: "Versatile 2-in-1 convertible laptop",
			Price:       1499.99,
			Rating:      4.5,
			Specifications: models.Specifications{
				"processor":    "Intel Core i7-1355U",
				"memory":       "16GB",
				"storage":      "512GB SSD",
				"display":      "13.5-inch OLED",
				"graphics":     "Intel Iris Xe",
				"battery_life": "Up to 17 hours",
				"weight":       "1.36 kg",
				"touchscreen":  "Yes",
			},
		},
		{
			Name:        "Lenovo ThinkPad X1 Carbon",
			ImageURL:    "https://example.com/images/thinkpad.jpg",
			Description: "Business-class laptop with exceptional keyboard",
			Price:       1699.99,
			Rating:      4.7,
			Specifications: models.Specifications{
				"processor":    "Intel Core i7-1355U",
				"memory":       "16GB",
				"storage":      "512GB SSD",
				"display":      "14-inch WQXGA",
				"graphics":     "Intel Iris Xe",
				"battery_life": "Up to 15 hours",
				"weight":       "1.12 kg",
				"durability":   "MIL-STD tested",
			},
		},
		{
			Name:        "ASUS ROG Zephyrus G14",
			ImageURL:    "https://example.com/images/asus-rog.jpg",
			Description: "Gaming laptop with powerful GPU and compact design",
			Price:       1599.99,
			Rating:      4.4,
			Specifications: models.Specifications{
				"processor":    "AMD Ryzen 9 7940HS",
				"memory":       "16GB",
				"storage":      "1TB SSD",
				"display":      "14-inch QHD 165Hz",
				"graphics":     "NVIDIA RTX 4060",
				"battery_life": "Up to 8 hours",
				"weight":       "1.65 kg",
				"rgb_keyboard": "Yes",
			},
		},
	}

	insertQuery := `
		INSERT INTO items (name, image_url, description, price, rating, specifications)
		VALUES (?, ?, ?, ?, ?, ?)
	`

	for _, item := range seedItems {
		specsJSON, err := json.Marshal(item.Specifications)
		if err != nil {
			return fmt.Errorf("failed to marshal specifications: %w", err)
		}

		if _, err := r.DB.ExecContext(
			ctx,
			insertQuery,
			item.Name,
			item.ImageURL,
			item.Description,
			item.Price,
			item.Rating,
			specsJSON,
		); err != nil {
			return fmt.Errorf("failed to insert seed item: %w", err)
		}
	}

	return nil
}
