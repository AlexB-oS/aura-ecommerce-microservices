// Package migrations embeds the user service's SQL migration files into the
// binary. Embedding means the compiled service is fully self-contained: there
// is no need to ship .sql files alongside the container or mount them in.
package migrations

import "embed"

//go:embed *.sql
var FS embed.FS
