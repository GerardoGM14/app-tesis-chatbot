// Package hmacauth firma y verifica las peticiones internas entre el backend y
// wa-service. La firma es HMAC-SHA256 de "timestamp.cuerpo"; el timestamp
// limita la ventana en la que una petición capturada podría repetirse.
package hmacauth

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"net/http"
	"strconv"
	"time"
)

const (
	HeaderSignature = "X-Signature"
	HeaderTimestamp = "X-Timestamp"
	// MaxSkew es la diferencia máxima aceptada entre relojes.
	MaxSkew = 5 * time.Minute
	// maxBody limita el cuerpo que se lee para verificar la firma.
	maxBody = 1 << 20
)

// Sign devuelve la firma hexadecimal de timestamp y cuerpo.
func Sign(secret []byte, timestamp int64, body []byte) string {
	mac := hmac.New(sha256.New, secret)
	mac.Write([]byte(strconv.FormatInt(timestamp, 10)))
	mac.Write([]byte{'.'})
	mac.Write(body)
	return hex.EncodeToString(mac.Sum(nil))
}

// Verify comprueba la firma en tiempo constante y que el timestamp esté dentro de la ventana.
func Verify(secret []byte, timestamp int64, body []byte, signature string, now time.Time) bool {
	sent := time.Unix(timestamp, 0)
	if now.Sub(sent) > MaxSkew || sent.Sub(now) > MaxSkew {
		return false
	}
	expected := Sign(secret, timestamp, body)
	return hmac.Equal([]byte(expected), []byte(signature))
}

// Middleware rechaza con 401 las peticiones sin firma válida.
func Middleware(secret []byte, now func() time.Time) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			timestamp, err := strconv.ParseInt(r.Header.Get(HeaderTimestamp), 10, 64)
			if err != nil {
				http.Error(w, `{"error":{"code":"firma_invalida"}}`, http.StatusUnauthorized)
				return
			}
			body, err := io.ReadAll(http.MaxBytesReader(w, r.Body, maxBody))
			if err != nil {
				http.Error(w, `{"error":{"code":"cuerpo_invalido"}}`, http.StatusBadRequest)
				return
			}
			if !Verify(secret, timestamp, body, r.Header.Get(HeaderSignature), now()) {
				http.Error(w, `{"error":{"code":"firma_invalida"}}`, http.StatusUnauthorized)
				return
			}
			r.Body = io.NopCloser(bytes.NewReader(body))
			next.ServeHTTP(w, r)
		})
	}
}
