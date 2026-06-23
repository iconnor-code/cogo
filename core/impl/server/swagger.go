package server

import (
	"io/fs"
	"net/http"

	"github.com/swaggest/swgui/v5emb"
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
	mux.Handle("/swagger/", v5emb.New(titleOrDefault(opt.Title), specPath, "/swagger/"))

	if apiHandler == nil {
		return mux
	}
	mux.Handle("/", apiHandler)
	return mux
}

func titleOrDefault(title string) string {
	if title == "" {
		return "Swagger UI"
	}
	return title
}
