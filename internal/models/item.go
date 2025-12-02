package models

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
)

// Item representa un producto dentro de la capa de dominio.
// Este modelo se utiliza tanto para la persistencia como para la exposición
// en la API.
type Item struct {
	ID             int64          `json:"id" db:"id"`
	Name           string         `json:"name" db:"name"`
	ImageURL       string         `json:"image_url" db:"image_url"`
	Description    string         `json:"description" db:"description"`
	Price          float64        `json:"price" db:"price"`
	Rating         float64        `json:"rating" db:"rating"`
	Specifications Specifications `json:"specifications" db:"specifications"`
}

// Specifications define un mapa genérico utilizado para almacenar características
// dinámicas de un producto.
type Specifications map[string]interface{}

// Value implementa la interfaz driver.Valuer, permitiendo que el tipo Specifications
// sea convertido automáticamente a JSON antes de ser almacenado en la base de datos.
func (s Specifications) Value() (driver.Value, error) {
	if s == nil {
		return nil, nil
	}
	return json.Marshal(s)
}

// Scan implementa la interfaz sql.Scanner, permitiendo que los datos almacenados como JSON
// en la base de datos sean convertidos nuevamente a un mapa Specifications.
//
// La función valida el tipo recibido y realiza la deserialización correspondiente.
func (s *Specifications) Scan(value interface{}) error {
	if value == nil {
		*s = nil
		return nil
	}

	bytes, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("failed to deserialize specifications: value is not a []byte")
	}

	return json.Unmarshal(bytes, s)
}

// CompareRequest representa la estructura esperada en el cuerpo de la solicitud
// cuando un cliente solicita comparar múltiples ítems.
//
// Se valida que:
//   - Debe enviarse una lista de IDs.
//   - Debe contener al menos 2 ítems.
//   - No debe exceder los 10 ítems.
type CompareRequest struct {
	ItemIDs []int64 `json:"item_ids" validate:"required,min=2,max=10"`
}

// CompareResponse representa la estructura enviada como respuesta al cliente
// después de realizar la comparación de ítems.
type CompareResponse struct {
	Items      []Item            `json:"items"`
	Comparison ComparisonDetails `json:"comparison"`
}

// ComparisonDetails contiene el resultado del análisis comparativo entre ítems.
type ComparisonDetails struct {
	PriceRange  PriceRange         `json:"price_range"`
	RatingRange RatingRange        `json:"rating_range"`
	CommonSpecs []string           `json:"common_specs"`
	UniqueSpecs map[int64][]string `json:"unique_specs"`
}

// PriceRange representa el precio mínimo y máximo entre un conjunto de ítems.
type PriceRange struct {
	Min float64 `json:"min"`
	Max float64 `json:"max"`
}

// RatingRange representa la calificación mínima y máxima entre múltiples ítems.
type RatingRange struct {
	Min float64 `json:"min"`
	Max float64 `json:"max"`
}
