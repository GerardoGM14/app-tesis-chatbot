package reqmeta

import (
	"net/http"
	"net/http/httptest"
	"testing"

	chimw "github.com/go-chi/chi/v5/middleware"
)

func clientIPOf(wrap func(http.Handler) http.Handler, xff ...string) string {
	var got string
	h := Middleware(http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) { got = ClientIP(r.Context()) }))
	if wrap != nil {
		h = wrap(h)
	}
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "172.18.0.5:4321"
	for _, v := range xff {
		req.Header.Add("X-Forwarded-For", v)
	}
	req.Header.Set("X-Real-IP", "6.6.6.6")
	req.Header.Set("True-Client-IP", "6.6.6.6")
	h.ServeHTTP(httptest.NewRecorder(), req)
	return got
}

func TestMiddlewareClientIP(t *testing.T) {
	proxy := chimw.ClientIPFromXFFTrustedProxies(1)
	cases := []struct {
		label string
		wrap  func(http.Handler) http.Handler
		xff   []string
		want  string
	}{
		{"sin proxy se ignoran las cabeceras", nil, []string{"6.6.6.6"}, "172.18.0.5"},
		{"con proxy se usa la IP que agregó", proxy, []string{"190.10.20.30"}, "190.10.20.30"},
		{"una IP falsificada por el cliente no cuenta", proxy, []string{"6.6.6.6, 190.10.20.30"}, "190.10.20.30"},
		{"sin cabecera se usa la conexión", proxy, nil, "172.18.0.5"},
		{"con cabecera inválida se usa la conexión", proxy, []string{"basura"}, "172.18.0.5"},
	}
	for _, c := range cases {
		if got := clientIPOf(c.wrap, c.xff...); got != c.want {
			t.Errorf("%s: %q, se esperaba %q", c.label, got, c.want)
		}
	}
}

func TestHostOnly(t *testing.T) {
	if got := hostOnly("sin-puerto"); got != "sin-puerto" {
		t.Errorf("hostOnly: %q", got)
	}
}
