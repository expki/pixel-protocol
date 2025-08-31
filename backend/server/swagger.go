package server

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"net/http"

	"gopkg.in/yaml.v3"
)

//go:embed swagger.yaml
var swaggerYAML []byte

// SwaggerJSON converts the embedded YAML to JSON for compatibility
func getSwaggerJSON() ([]byte, error) {
	var data interface{}
	if err := yaml.Unmarshal(swaggerYAML, &data); err != nil {
		return nil, err
	}
	return json.Marshal(data)
}

// HandleSwagger serves the Swagger UI
func (s *Server) HandleSwagger(w http.ResponseWriter, r *http.Request) {
	// Serve different content based on the path
	path := r.URL.Path

	switch path {
	case "/swagger", "/swagger/":
		// Serve the Swagger UI HTML
		serveSwaggerUI(w, r)
	case "/swagger/swagger.json":
		// Serve the OpenAPI specification as JSON
		serveSwaggerJSON(w, r)
	case "/swagger/swagger.yaml":
		// Serve the OpenAPI specification as YAML
		serveSwaggerYAML(w, r)
	default:
		http.NotFound(w, r)
	}
}

func serveSwaggerUI(w http.ResponseWriter, r *http.Request) {
	html := `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <title>Pixel Protocol API - Swagger UI</title>
    <link rel="stylesheet" href="https://cdn.jsdelivr.net/npm/swagger-ui-dist@5.9.0/swagger-ui.css">
    <style>
        html {
            box-sizing: border-box;
            overflow: -moz-scrollbars-vertical;
            overflow-y: scroll;
        }
        *, *:before, *:after {
            box-sizing: inherit;
        }
        body {
            margin: 0;
            background: #fafafa;
        }
    </style>
</head>
<body>
    <div id="swagger-ui"></div>
    <script src="https://cdn.jsdelivr.net/npm/swagger-ui-dist@5.28.0/swagger-ui-bundle.js"></script>
    <script src="https://cdn.jsdelivr.net/npm/swagger-ui-dist@5.28.0/swagger-ui-standalone-preset.js"></script>
    <script>
        window.onload = function() {
            window.ui = SwaggerUIBundle({
                url: "/swagger/swagger.json",
                dom_id: '#swagger-ui',
                deepLinking: true,
                presets: [
                    SwaggerUIBundle.presets.apis,
                    SwaggerUIStandalonePreset
                ],
                plugins: [
                    SwaggerUIBundle.plugins.DownloadUrl
                ],
                layout: "StandaloneLayout",
                defaultModelsExpandDepth: 1,
                defaultModelExpandDepth: 1,
                docExpansion: "none",
                filter: true,
                showExtensions: true,
                showCommonExtensions: true,
                validatorUrl: null
            });
        };
    </script>
</body>
</html>`

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(html))
}

func serveSwaggerJSON(w http.ResponseWriter, r *http.Request) {
	jsonData, err := getSwaggerJSON()
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to convert swagger spec: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Write(jsonData)
}

func serveSwaggerYAML(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/x-yaml")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Write(swaggerYAML)
}
