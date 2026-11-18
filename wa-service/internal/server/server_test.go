package server

import (
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"wa-service/internal/hmacauth"
)

var secret = []byte("secreto-interno-de-pruebas-32-bytes")

func TestHealthz(t *testing.T) {
	rec := httptest.NewRecorder()
	New(secret).ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/healthz", nil))
	if rec.Code != http.StatusOK {
		t.Errorf("healthz: %d", rec.Code)
	}
}

func TestInternalRoutesRequireSignature(t *testing.T) {
	h := New(secret)

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/internal/sessions/abc/status", nil))
	if rec.Code != http.StatusUnauthorized {
		t.Errorf("sin firma: %d", rec.Code)
	}

	now := time.Now().Unix()
	req := httptest.NewRequest(http.MethodGet, "/internal/sessions/abc/status", nil)
	req.Header.Set(hmacauth.HeaderTimestamp, strconv.FormatInt(now, 10))
	req.Header.Set(hmacauth.HeaderSignature, hmacauth.Sign(secret, now, nil))
	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusNotImplemented {
		t.Errorf("con firma: %d", rec.Code)
	}
}
