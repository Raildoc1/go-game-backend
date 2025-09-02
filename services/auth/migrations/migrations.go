// Package migrations embeds SQL migration files for the auth service.
package migrations

import "embed"

// Files holds embedded SQL migration files.
//
//go:embed *.sql
var Files embed.FS
