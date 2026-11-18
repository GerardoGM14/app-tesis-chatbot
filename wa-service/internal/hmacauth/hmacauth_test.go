package hmacauth

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"
)

var secret = []byte("secreto-interno-de-pruebas-32-bytes")

func TestSignAndVerify(t *testing.T) {
	now := time.Unix(1_800_000_000, 0)
	body := []byte(`{"tenant":"a"}`)
	sig := Sign(secret, now.Unix(), body)

	if !Verify(secret, now.Unix(), body, sig, now) {
		t.Fatal("una firma válida debe aceptarse")
	}
	cases := map[string]bool{
		"cuerpo alterado": Verify(secret, now.Unix(), []byte(`{"tenant":"b"}`), sig, now),
		"otra clave":      Verify([]byte("otra"), now.Unix(), body, sig, now),
		"firma vacía":     Verify(secret, now.Unix(), body, "", now),
		"muy antigua":     Verify(secret, now.Unix(), body, sig, now.Add(MaxSkew+time.Second)),
		"del futuro":      Verify(secret, now.Unix(), body, sig, now.Add(-MaxSkew-time.Second)),
	}
	for name, accepted := range cases {
		if accepted {
			t.Errorf("%s: no debe aceptarse", name)
		}
	}
}

func TestMiddleware(t *testing.T) {
	now := time.Unix(1_800_000_000, 0)
	var received string
	h := Middleware(secret, func() time.Time { return now })(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b, _ := io.ReadAll(r.Body)
		received = string(b)
		w.WriteHeader(http.StatusOK)
	}))

	send := func(ts, sig, body string) int {
		req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
		req.Header.Set(HeaderTimestamp, ts)
		req.Header.Set(HeaderSignature, sig)
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, req)
		return rec.Code
	}

	body := `{"ok":true}`
	ts := strconv.FormatInt(now.Unix(), 10)
	if code := send(ts, Sign(secret, now.Unix(), []byte(body)), body); code != http.StatusOK || received != body {
		t.Errorf("petición firmada: %d, cuerpo %q", code, received)
	}
	if code := send(ts, "firma-falsa", body); code != http.StatusUnauthorized {
		t.Errorf("firma falsa: %d", code)
	}
	if code := send("no-numero", "x", body); code != http.StatusUnauthorized {
		t.Errorf("timestamp inválido: %d", code)
	}
	big := strings.Repeat("a", maxBody+1)
	if code := send(ts, Sign(secret, now.Unix(), []byte(big)), big); code != http.StatusBadRequest {
		t.Errorf("cuerpo excesivo: %d", code)
	}
}
