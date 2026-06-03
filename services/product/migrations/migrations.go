// Package migrations embeds the product service's SQL migration files so the
// compiled binary is self-contained.
package migrations

import "embed"

//go:embed *.sql
var FS embed.FS
