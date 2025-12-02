package errors

import (
	"fmt"
	"net/http"
)

// ErrorCode es un tipo de dato que representa un código de error estandarizado para las respuestas de la API.
type ErrorCode string

const (
	ErrorCodeNotFound        ErrorCode = "NOT_FOUND"
	ErrorCodeBadRequest      ErrorCode = "BAD_REQUEST"
	ErrorCodeInternalServer  ErrorCode = "INTERNAL_SERVER_ERROR"
	ErrorCodeValidation      ErrorCode = "VALIDATION_ERROR"
	ErrorCodeTooManyRequests ErrorCode = "TOO_MANY_REQUESTS"
)

// DomainError es un tipo de dato que representa un error de dominio.
type DomainError struct {
	Code    ErrorCode
	Message string
	Err     error
}

// Error implementa la interfaz error.
func (e *DomainError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %s (%v)", e.Code, e.Message, e.Err)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// Unwrap devuelve el error subyacente para el soporte de wrapping de errores.
func (e *DomainError) Unwrap() error {
	return e.Err
}

// NewDomainError crea un nuevo error de dominio con soporte de wrapping de errores.
func NewDomainError(code ErrorCode, message string, err error) *DomainError {
	return &DomainError{
		Code:    code,
		Message: message,
		Err:     err,
	}
}

// ErrorResponse es un tipo de dato que representa la respuesta de error estandarizada para la API.
type ErrorResponse struct {
	Error   bool      `json:"error"`
	Message string    `json:"message"`
	Code    ErrorCode `json:"code"`
}

// ToErrorResponse convierte un ErrorDomain en un ErrorResponse.
func (e *DomainError) ToErrorResponse() ErrorResponse {
	return ErrorResponse{
		Error:   true,
		Message: e.Message,
		Code:    e.Code,
	}
}

// HTTPStatus devuelve el código de estado HTTP apropiado para el error de dominio.
func (e *DomainError) HTTPStatus() int {
	switch e.Code {
	case ErrorCodeNotFound:
		return http.StatusNotFound
	case ErrorCodeBadRequest:
		return http.StatusBadRequest
	case ErrorCodeValidation:
		return http.StatusUnprocessableEntity
	case ErrorCodeTooManyRequests:
		return http.StatusTooManyRequests
	case ErrorCodeInternalServer:
		return http.StatusInternalServerError
	default:
		return http.StatusInternalServerError
	}
}

// Funciones helper para crear errores de dominio comunes
func NewNotFoundError(resource string, id interface{}) *DomainError {
	return NewDomainError(
		ErrorCodeNotFound,
		fmt.Sprintf("%s with id %v not found", resource, id),
		nil,
	)
}

// NewBadRequestError crea un error de dominio de tipo "solicitud inválida"
func NewBadRequestError(message string, err error) *DomainError {
	return NewDomainError(ErrorCodeBadRequest, message, err)
}

// NewValidationError crea un error de dominio de tipo "validación"
func NewValidationError(message string, err error) *DomainError {
	return NewDomainError(ErrorCodeValidation, message, err)
}

// NewInternalServerError crea un error de dominio de tipo "error interno del servidor"
func NewInternalServerError(message string, err error) *DomainError {
	return NewDomainError(ErrorCodeInternalServer, message, err)
}
