// Package migrations embeds the order service's SQL migrations.
package migrations

import "embed"

//go:embed *.sql
var FS embed.FS
