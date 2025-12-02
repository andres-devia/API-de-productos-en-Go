package repositories

import "errors"

// ErrNotFound es devuelto por la capa de repositorio cuando
// no se puede encontrar un recurso específico.
var ErrNotFound = errors.New("recurso no encontrado")
