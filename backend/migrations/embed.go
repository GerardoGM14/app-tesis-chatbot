// Package migrations embebe los scripts SQL en el binario.
package migrations

import "embed"

// FS contiene los archivos .sql en orden de aplicación.
//
//go:embed *.sql
var FS embed.FS
