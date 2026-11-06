// Package httpx contiene utilidades HTTP: respuestas JSON, traducción de
// errores y decodificación segura del cuerpo de las peticiones.
package httpx

import (
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/google/uuid"

	"backend/internal/platform/apperr"
)

// MaxBodyBytes limita el tamaño del cuerpo de las peticiones.
const MaxBodyBytes = 1 << 20

type errorBody struct {
	Error errorPayload `json:"error"`
}

type errorPayload struct {
	Code      string            `json:"code"`
	Message   string            `json:"message"`
	Fields    map[string]string `json:"fields,omitempty"`
	RequestID string            `json:"requestId,omitempty"`
}

// JSON escribe v como JSON con el código de estado indicado.
func JSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	if v != nil {
		_ = json.NewEncoder(w).Encode(v)
	}
}

// NoContent responde 204.
func NoContent(w http.ResponseWriter) {
	w.WriteHeader(http.StatusNoContent)
}

var statusByKind = map[apperr.Kind]int{
	apperr.KindInvalid:         http.StatusUnprocessableEntity,
	apperr.KindUnauthorized:    http.StatusUnauthorized,
	apperr.KindForbidden:       http.StatusForbidden,
	apperr.KindNotFound:        http.StatusNotFound,
	apperr.KindConflict:        http.StatusConflict,
	apperr.KindTooManyRequests: http.StatusTooManyRequests,
}

// Error traduce err a una respuesta JSON. Los errores no esperados se
// registran y se responden con un 500 genérico, sin detalles internos.
func Error(w http.ResponseWriter, r *http.Request, err error) {
	reqID := middleware.GetReqID(r.Context())
	if appErr, ok := apperr.As(err); ok {
		status, known := statusByKind[appErr.Kind]
		if !known {
			status = http.StatusBadRequest
		}
		JSON(w, status, errorBody{Error: errorPayload{
			Code: appErr.Code, Message: appErr.Message, Fields: appErr.Fields, RequestID: reqID,
		}})
		return
	}
	slog.ErrorContext(r.Context(), "error no controlado", "error", err, "request_id", reqID, "path", r.URL.Path)
	JSON(w, http.StatusInternalServerError, errorBody{Error: errorPayload{
		Code: "error_interno", Message: "Ocurrió un error inesperado", RequestID: reqID,
	}})
}

var errBadJSON = apperr.New(apperr.KindInvalid, "json_invalido", "El cuerpo de la petición no es un JSON válido")

// Decode lee un único objeto JSON, rechaza campos desconocidos y limita el tamaño.
func Decode(w http.ResponseWriter, r *http.Request, dst any) error {
	if ct := r.Header.Get("Content-Type"); !strings.HasPrefix(ct, "application/json") {
		return apperr.New(apperr.KindInvalid, "content_type_invalido", "Se esperaba Content-Type application/json")
	}
	dec := json.NewDecoder(http.MaxBytesReader(w, r.Body, MaxBodyBytes))
	dec.DisallowUnknownFields()
	if err := dec.Decode(dst); err != nil {
		var tooLarge *http.MaxBytesError
		if errors.As(err, &tooLarge) {
			return apperr.New(apperr.KindInvalid, "cuerpo_muy_grande", "El cuerpo de la petición es demasiado grande")
		}
		return errBadJSON
	}
	if err := dec.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		return errBadJSON
	}
	return nil
}

// PathUUID lee un parámetro de ruta como UUID. Un valor inválido se trata como
// recurso inexistente para no revelar el formato interno.
func PathUUID(r *http.Request, name string) (uuid.UUID, error) {
	id, err := uuid.Parse(chi.URLParam(r, name))
	if err != nil {
		return uuid.Nil, apperr.ErrNotFound
	}
	return id, nil
}
