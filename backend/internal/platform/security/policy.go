package security

import (
	"strings"
	"unicode/utf8"
)

const (
	MinPasswordLength = 10
	MaxPasswordLength = 128
)

// commonPasswords contiene contraseñas frecuentes que se rechazan siempre.
var commonPasswords = map[string]struct{}{
	"1234567890": {}, "12345678910": {}, "0123456789": {}, "qwertyuiop": {},
	"contraseña": {}, "contrasena": {}, "contraseña1": {}, "contrasena1": {},
	"password12": {}, "password123": {}, "password1234": {}, "passw0rd123": {},
	"administrador": {}, "admin12345": {}, "admin123456": {}, "1q2w3e4r5t": {},
	"iloveyou12": {}, "abc1234567": {}, "qwerty1234": {}, "1111111111": {},
	"peru123456": {}, "chimbote123": {}, "sertech123": {}, "bienvenido1": {},
	"bienvenido123": {}, "superadmin1": {}, "superadmin123": {}, "whatsapp123": {},
}

// CheckPasswordPolicy devuelve un mensaje si la contraseña no cumple la
// política, o cadena vacía si la cumple.
func CheckPasswordPolicy(password, email string) string {
	n := utf8.RuneCountInString(password)
	switch {
	case n < MinPasswordLength:
		return "Debe tener al menos 10 caracteres"
	case n > MaxPasswordLength:
		return "Debe tener como máximo 128 caracteres"
	}
	lower := strings.ToLower(password)
	if _, common := commonPasswords[lower]; common {
		return "Es una contraseña demasiado común"
	}
	if local, _, ok := strings.Cut(strings.ToLower(email), "@"); ok && len(local) >= 4 && strings.Contains(lower, local) {
		return "No debe contener tu email"
	}
	if allSameRune(password) {
		return "No debe repetir un solo carácter"
	}
	return ""
}

func allSameRune(s string) bool {
	first, _ := utf8.DecodeRuneInString(s)
	for _, r := range s {
		if r != first {
			return false
		}
	}
	return true
}
