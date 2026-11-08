// Package domain contiene las entidades y reglas de las empresas (tenants).
package domain

import (
	"strings"
	"time"
	"unicode"

	"github.com/google/uuid"
	"golang.org/x/text/runes"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"

	"backend/internal/platform/apperr"
	"backend/internal/platform/validate"
)

// Estados de una empresa.
const (
	EstadoActivo     = "activo"
	EstadoSuspendido = "suspendido"
)

var (
	ErrRUCTaken  = apperr.New(apperr.KindConflict, "ruc_registrado", "Ya existe una empresa con ese RUC")
	ErrNameTaken = apperr.New(apperr.KindConflict, "nombre_registrado", "Ya existe una empresa con un nombre equivalente")
)

// Tenant es una empresa cliente de la plataforma.
type Tenant struct {
	ID       uuid.UUID `json:"id"`
	Nombre   string    `json:"nombre"`
	RUC      string    `json:"ruc"`
	Slug     string    `json:"slug"`
	Estado   string    `json:"estado"`
	Usuarios int       `json:"usuarios"`
	CreadoEn time.Time `json:"creadoEn"`
}

// CreateInput son los datos para dar de alta una empresa con su primer administrador.
type CreateInput struct {
	Nombre      string `json:"nombre"`
	RUC         string `json:"ruc"`
	AdminNombre string `json:"adminNombre"`
	AdminEmail  string `json:"adminEmail"`
}

// Normalize limpia espacios.
func (in CreateInput) Normalize() CreateInput {
	in.Nombre = strings.Join(strings.Fields(in.Nombre), " ")
	in.RUC = strings.TrimSpace(in.RUC)
	in.AdminNombre = strings.TrimSpace(in.AdminNombre)
	in.AdminEmail = strings.ToLower(strings.TrimSpace(in.AdminEmail))
	return in
}

// Validate verifica los datos de la empresa. Los del administrador los valida
// el módulo de usuarios.
func (in CreateInput) Validate() error {
	v := validate.New()
	v.Check(validate.Length(in.Nombre, 2, 120), "nombre", "Ingresa un nombre de 2 a 120 caracteres")
	v.Check(Slugify(in.Nombre) != "", "nombre", "El nombre debe contener letras o números")
	v.Check(validate.RUC(in.RUC), "ruc", "Ingresa un RUC válido de 11 dígitos")
	return v.Err()
}

// ValidateName verifica un nombre de empresa.
func ValidateName(nombre string) error {
	v := validate.New()
	v.Check(validate.Length(nombre, 2, 120), "nombre", "Ingresa un nombre de 2 a 120 caracteres")
	v.Check(Slugify(nombre) != "", "nombre", "El nombre debe contener letras o números")
	return v.Err()
}

// Created es la empresa creada con la contraseña temporal de su administrador.
type Created struct {
	Empresa          Tenant `json:"empresa"`
	AdminEmail       string `json:"adminEmail"`
	PasswordTemporal string `json:"passwordTemporal"`
}

// Slugify genera un identificador legible: minúsculas, sin tildes y con guiones.
func Slugify(nombre string) string {
	t := transform.Chain(norm.NFD, runes.Remove(runes.In(unicode.Mn)), norm.NFC)
	plain, _, err := transform.String(t, strings.ToLower(nombre))
	if err != nil {
		plain = strings.ToLower(nombre)
	}
	var b strings.Builder
	dash := false
	for _, r := range plain {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			b.WriteRune(r)
			dash = false
			continue
		}
		if b.Len() > 0 && !dash {
			b.WriteByte('-')
			dash = true
		}
	}
	slug := strings.TrimSuffix(b.String(), "-")
	if len(slug) > 80 {
		slug = strings.TrimSuffix(slug[:80], "-")
	}
	return slug
}
