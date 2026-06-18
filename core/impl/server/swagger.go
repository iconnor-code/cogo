package server

import (
	"fmt"
	"io/fs"
	"net/http"
	"strings"
)

type SwaggerOption struct {
	Title    string
	SpecFile string
	SpecFS   fs.FS
}

func NewSwaggerHandler(apiHandler http.Handler, opt SwaggerOption) http.Handler {
	mux := http.NewServeMux()
	specPath := "/swagger/openapi.json"

	mux.HandleFunc("/swagger", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/swagger/", http.StatusMovedPermanently)
	})
	mux.HandleFunc(specPath, func(w http.ResponseWriter, r *http.Request) {
		data, err := fs.ReadFile(opt.SpecFS, opt.SpecFile)
		if err != nil {
			http.Error(w, "openapi spec not found", http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		_, _ = w.Write(data)
	})
	mux.HandleFunc("/swagger/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/swagger/" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = fmt.Fprintf(w, swaggerHTML, htmlTitle(opt.Title), specPath)
	})

	if apiHandler == nil {
		return mux
	}
	mux.Handle("/", apiHandler)
	return mux
}

func htmlTitle(title string) string {
	if strings.TrimSpace(title) == "" {
		return "Swagger UI"
	}
	replacer := strings.NewReplacer(
		"&", "&amp;",
		"<", "&lt;",
		">", "&gt;",
		`"`, "&quot;",
		"'", "&#39;",
	)
	return replacer.Replace(title)
}

const swaggerHTML = `<!doctype html>
<html lang="en">
<head>
  <meta charset="utf-8">
  <title>%s</title>
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <link rel="stylesheet" href="https://unpkg.com/swagger-ui-dist@5/swagger-ui.css">
  <style>
    body { margin: 0; background: #fafafa; }
  </style>
</head>
<body>
  <div id="swagger-ui"></div>
  <script src="https://unpkg.com/swagger-ui-dist@5/swagger-ui-bundle.js"></script>
  <script>
    window.onload = function() {
      window.ui = SwaggerUIBundle({
        url: "%s",
        dom_id: "#swagger-ui",
        deepLinking: true,
        presets: [SwaggerUIBundle.presets.apis],
        layout: "BaseLayout"
      });
    };
  </script>
</body>
</html>`
