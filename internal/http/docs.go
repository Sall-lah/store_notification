package http

import (
	"fmt"
	"net/http"

	"github.com/sall-lah/store_notification/docs"
)

// DocsHandler serves raw OpenAPI 3.1 specifications and interactive documentation interfaces.
type DocsHandler struct{}

// NewDocsHandler constructs a DocsHandler instance.
func NewDocsHandler() *DocsHandler {
	return &DocsHandler{}
}

// ServeOpenAPIYAML returns the embedded openapi.yaml specification.
func (d *DocsHandler) ServeOpenAPIYAML(w http.ResponseWriter, r *http.Request) {
	data, err := docs.FS.ReadFile("openapi.yaml")
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to read openapi.yaml: %v", err), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/yaml; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(data)
}

// ServeOpenAPIJSON returns the embedded openapi.json specification.
func (d *DocsHandler) ServeOpenAPIJSON(w http.ResponseWriter, r *http.Request) {
	data, err := docs.FS.ReadFile("openapi.json")
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to read openapi.json: %v", err), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(data)
}

// ServeScalar renders the modern interactive Scalar documentation interface configured with
// relative spec paths (./openapi.yaml) to ensure compatibility behind reverse proxies and API gateway prefixes.
func (d *DocsHandler) ServeScalar(w http.ResponseWriter, r *http.Request) {
	scalarHTML := `<!doctype html>
<html lang="en">
  <head>
    <title>Store Notification Service API - Scalar</title>
    <meta charset="utf-8" />
    <meta name="viewport" content="width=device-width, initial-scale=1" />
    <link rel="icon" type="image/svg+xml" href="data:image/svg+xml,<svg xmlns='http://www.w3.org/2000/svg' viewBox='0 0 100 100'><text y='.9em' font-size='90'>📬</text></svg>" />
    <style>
      body { margin: 0; padding: 0; }
    </style>
  </head>
  <body>
    <script
      id="api-reference"
      data-url="./openapi.yaml"
      data-proxy-url=""
      data-configuration='{"theme": "purple", "hideModels": false}'
    ></script>
    <script src="https://cdn.jsdelivr.net/npm/@scalar/api-reference"></script>
  </body>
</html>`

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(scalarHTML))
}

// ServeSwagger renders the classic interactive Swagger UI interface configured with
// relative spec paths (./openapi.yaml) to ensure compatibility behind reverse proxies and API gateway prefixes.
func (d *DocsHandler) ServeSwagger(w http.ResponseWriter, r *http.Request) {
	swaggerHTML := `<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <title>Store Notification Service API - Swagger UI</title>
  <link rel="stylesheet" type="text/css" href="https://unpkg.com/swagger-ui-dist@5.11.0/swagger-ui.css" />
  <style>
    html { box-sizing: border-box; overflow: -moz-scrollbars-vertical; overflow-y: scroll; }
    *, *:before, *:after { box-sizing: inherit; }
    body { margin:0; background: #fafafa; }
  </style>
</head>
<body>
<div id="swagger-ui"></div>
<script src="https://unpkg.com/swagger-ui-dist@5.11.0/swagger-ui-bundle.js"></script>
<script src="https://unpkg.com/swagger-ui-dist@5.11.0/swagger-ui-standalone-preset.js"></script>
<script>
window.onload = function() {
  window.ui = SwaggerUIBundle({
    url: "./openapi.yaml",
    dom_id: '#swagger-ui',
    deepLinking: true,
    presets: [
      SwaggerUIBundle.presets.apis,
      SwaggerUIStandalonePreset
    ],
    plugins: [
      SwaggerUIBundle.plugins.DownloadUrl
    ],
    layout: "StandaloneLayout"
  });
};
</script>
</body>
</html>`

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(swaggerHTML))
}
