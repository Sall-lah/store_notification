package templates

import "embed"

// FS embeds all email HTML and plaintext templates into the compiled binary.
// This guarantees zero-dependency runtime asset resolution across environments.
//
//go:embed layout/*.html auth/*.html auth/*.txt order/*.html admin/*.html
var FS embed.FS
