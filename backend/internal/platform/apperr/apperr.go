// Package apperr define los errores de aplicación sin depender del transporte.
// La capa HTTP traduce cada Kind a un código de estado.
package apperr

import "errors"

// Kind clasifica un error para decidir cómo se informa al cliente.
type Kind int

const (
	KindInvalid Kind = iota + 1
	KindUnauthorized
	KindForbidden
	KindNotFound
	KindConflict
	KindTooManyRequests
)

// Error es un error esperado del negocio, con un mensaje apto para el usuario.
type Error struct {
	Kind    Kind
	Code    string
	Message string
	Fields  map[string]string
}

func (e *Error) Error() string { return e.Code + ": " + e.Message }

// New crea un error de aplicación.
func New(kind Kind, code, message string) *Error {
	return &Error{Kind: kind, Code: code, Message: message}
}

// Validation crea un error de validación con el mensaje de cada campo inválido.
func Validation(fields map[string]string) *Error {
	return &Error{Kind: KindInvalid, Code: "validacion", Message: "Revisa los datos enviados", Fields: fields}
}

// As extrae un *Error de la cadena de errores.
func As(err error) (*Error, bool) {
	var appErr *Error
	if errors.As(err, &appErr) {
		return appErr, true
	}
	return nil, false
}

var (
	ErrNotFound     = New(KindNotFound, "no_encontrado", "El recurso no existe")
	ErrForbidden    = New(KindForbidden, "prohibido", "No tienes permiso para esta acción")
	ErrUnauthorized = New(KindUnauthorized, "no_autenticado", "Inicia sesión para continuar")
	ErrTooMany      = New(KindTooManyRequests, "demasiadas_solicitudes", "Demasiadas solicitudes, espera un momento")
)
