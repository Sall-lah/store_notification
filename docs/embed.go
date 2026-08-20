package docs

import "embed"

// FS embeds openapi.yaml and openapi.json directly into the service binary.
//
//go:embed openapi.yaml openapi.json
var FS embed.FS
