package validate

import (
	"testing"

	"backend/internal/platform/apperr"
)

func TestRUC(t *testing.T) {
	cases := map[string]bool{
		"20131312955": true,  // RUC de SUNAT
		"20600000005": true,  // dígito verificador calculado
		"20131312954": false, // dígito verificador incorrecto
		"30131312955": false, // prefijo inválido
		"2013131295":  false, // longitud incorrecta
		"2013131295A": false, // caracteres no numéricos
	}
	for ruc, want := range cases {
		if got := RUC(ruc); got != want {
			t.Errorf("RUC(%q) = %v, se esperaba %v", ruc, got, want)
		}
	}
}

func TestRUCCheckDigitRejectsInvalidInput(t *testing.T) {
	if _, ok := RUCCheckDigit("123"); ok {
		t.Error("se esperaba rechazo por longitud")
	}
	if _, ok := RUCCheckDigit("12345678x0"); ok {
		t.Error("se esperaba rechazo por caracteres no numéricos")
	}
}

func TestEmail(t *testing.T) {
	valid := []string{"a@b.pe", "admin@empresa.com.pe"}
	invalid := []string{"", "sin-arroba", "a@b", "a b@c.pe"}
	for _, e := range valid {
		if !Email(e) {
			t.Errorf("Email(%q) debería ser válido", e)
		}
	}
	for _, e := range invalid {
		if Email(e) {
			t.Errorf("Email(%q) debería ser inválido", e)
		}
	}
}

func TestLength(t *testing.T) {
	if !Length("  abc  ", 3, 3) {
		t.Error("debería ignorar espacios en los extremos")
	}
	if Length("ab", 3, 10) {
		t.Error("debería rechazar textos cortos")
	}
	if !Length("ñandú", 5, 5) {
		t.Error("debería contar caracteres, no bytes")
	}
}

func TestValidatorKeepsFirstMessagePerField(t *testing.T) {
	v := New()
	if v.Err() != nil {
		t.Fatal("un validador vacío no debe devolver error")
	}
	v.Check(false, "email", "primero")
	v.Check(false, "email", "segundo")
	v.Check(true, "nombre", "no debe aparecer")

	appErr, ok := apperr.As(v.Err())
	if !ok {
		t.Fatal("se esperaba un *apperr.Error")
	}
	if appErr.Fields["email"] != "primero" || len(appErr.Fields) != 1 {
		t.Errorf("campos inesperados: %v", appErr.Fields)
	}
}
