// Package migrations embeds SQL migration files for profile storage service.
package migrations

import "embed"

// Files holds embedded SQL migration files.
//
//go:embed *.sql
var Files embed.FS
