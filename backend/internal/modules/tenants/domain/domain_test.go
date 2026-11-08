package domain

import (
	"strings"
	"testing"
)

func TestSlugify(t *testing.T) {
	cases := map[string]string{
		"Sertech Perú E.I.R.L.":  "sertech-peru-e-i-r-l",
		"  Ferretería  Ñandú  ":  "ferreteria-nandu",
		"ACME 2026":              "acme-2026",
		"***":                    "",
		strings.Repeat("a", 100): strings.Repeat("a", 80),
	}
	for in, want := range cases {
		if got := Slugify(in); got != want {
			t.Errorf("Slugify(%q) = %q, se esperaba %q", in, got, want)
		}
	}
}

func TestCreateInput(t *testing.T) {
	in := CreateInput{Nombre: "  Mi   Empresa ", RUC: " 20600000005 ", AdminNombre: " Ana ", AdminEmail: " ANA@X.PE "}.Normalize()
	if in.Nombre != "Mi Empresa" || in.RUC != "20600000005" || in.AdminEmail != "ana@x.pe" || in.AdminNombre != "Ana" {
		t.Errorf("normalización inesperada: %+v", in)
	}
	if err := in.Validate(); err != nil {
		t.Errorf("entrada válida rechazada: %v", err)
	}
	if (CreateInput{Nombre: "***", RUC: "123"}).Validate() == nil {
		t.Error("se esperaba rechazo")
	}
}

func TestValidateName(t *testing.T) {
	if ValidateName("Nueva Empresa") != nil {
		t.Error("nombre válido rechazado")
	}
	if ValidateName("-") == nil || ValidateName("!!!!") == nil {
		t.Error("nombre inválido aceptado")
	}
}
