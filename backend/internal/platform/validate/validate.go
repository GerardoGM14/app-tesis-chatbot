// Package validate reúne reglas de validación reutilizables.
package validate

import (
	"regexp"
	"strings"
	"unicode/utf8"

	"backend/internal/platform/apperr"
)

// Validator acumula el primer mensaje de error de cada campo.
type Validator struct {
	fields map[string]string
}

// New crea un validador vacío.
func New() *Validator {
	return &Validator{fields: map[string]string{}}
}

// Check registra el mensaje si la condición no se cumple.
func (v *Validator) Check(ok bool, field, message string) {
	if ok {
		return
	}
	if _, exists := v.fields[field]; !exists {
		v.fields[field] = message
	}
}

// Err devuelve un error de validación o nil si todo es correcto.
func (v *Validator) Err() error {
	if len(v.fields) == 0 {
		return nil
	}
	return apperr.Validation(v.fields)
}

var emailPattern = regexp.MustCompile(`^[^\s@]+@[^\s@]+\.[^\s@]+$`)

// Email valida el formato básico de un correo.
func Email(value string) bool {
	return len(value) <= 254 && emailPattern.MatchString(value)
}

// Length valida la cantidad de caracteres sin contar espacios en los extremos.
func Length(value string, minLen, maxLen int) bool {
	n := utf8.RuneCountInString(strings.TrimSpace(value))
	return n >= minLen && n <= maxLen
}

var rucWeights = [10]int{5, 4, 3, 2, 7, 6, 5, 4, 3, 2}

// RUCCheckDigit calcula el dígito verificador SUNAT para los 10 primeros dígitos.
func RUCCheckDigit(first10 string) (byte, bool) {
	if len(first10) != 10 {
		return 0, false
	}
	sum := 0
	for i := range 10 {
		c := first10[i]
		if c < '0' || c > '9' {
			return 0, false
		}
		sum += int(c-'0') * rucWeights[i]
	}
	digit := 11 - sum%11
	switch digit {
	case 10:
		digit = 0
	case 11:
		digit = 1
	}
	return byte('0' + digit), true
}

// RUC valida un RUC peruano: 11 dígitos, prefijo válido y dígito verificador.
func RUC(value string) bool {
	if len(value) != 11 {
		return false
	}
	switch value[:2] {
	case "10", "15", "17", "20":
	default:
		return false
	}
	check, ok := RUCCheckDigit(value[:10])
	return ok && check == value[10]
}
