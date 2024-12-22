package server

import (
	"fmt"
	"html/template"
	"io/fs"
	"log"
	"net/http"
	"path/filepath"

	"github.com/sokkalf/hubro/utils"
)

type Middleware func(*Hubro) func(http.Handler) http.Handler

type Hubro struct {
	Mux         *http.ServeMux
	Server      *http.Server
	Templates   *template.Template
	middlewares []Middleware
}

var VendorLibs map[string]string = map[string]string{"htmx": "/vendor/htmx/htmx.min.js"}

func (h *Hubro) Use(m Middleware) {
	h.middlewares = append(h.middlewares, m)
}

func (h *Hubro) handlerWithMiddlewares(handler http.Handler) http.Handler {
	for _, m := range h.middlewares {
		handler = m(h)(handler)
	}
	return handler
}

func (h *Hubro) GetHandler() http.Handler {
	return h.handlerWithMiddlewares(h.Mux)
}

func (h *Hubro) initTemplates() {
	var err error
	funcMap := template.FuncMap{
		"title": func() string {
			return "Hubro"
		},
		"staticPath": func(path string) string {
			return "/static/" + path
		},
		"vendorPath": func(path string) string {
			return "/vendor/" + path
		},
		"vendor": func(path string) string {
			return VendorLibs[path]
		},
	}

	h.Templates, err = template.New("root").
		Funcs(funcMap).
		ParseGlob(filepath.Join("templates", "*.gohtml"))

	if err != nil {
		log.Fatalf("Error parsing templates: %v", err)
	}
	for _, t := range h.Templates.Templates() {
		fmt.Printf("Parsed template: %s\n", t.Name())
	}
}

func (h *Hubro) initStaticFiles() {
	fs := http.FileServer(http.Dir("./static"))
	h.Mux.Handle("GET /static/", http.StripPrefix("/static/", utils.FileServerWithDirectoryListingDisabled(fs)))
}

func (h *Hubro) initVendorDir(vendorDir fs.FS) {
	fs := http.FileServer(http.FS(vendorDir))
	h.Mux.Handle("GET /vendor/", http.StripPrefix("/vendor/", utils.FileServerWithDirectoryListingDisabled(fs)))
}

func (h *Hubro) indexHandler(w http.ResponseWriter, r *http.Request) {
	// Render the "index.gohtml" template
	err := h.Templates.ExecuteTemplate(w, "index.gohtml", nil)
	if err != nil {
		http.Error(w, "Failed to render template", http.StatusInternalServerError)
	}
}

// pingHandler is a simple route that returns "Pong!" text.
// We set an HTMX response header ("HX-Trigger") to demonstrate sending events back to the client.
func (h *Hubro) pingHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("HX-Trigger", "pongReceived")
	fmt.Fprintln(w, `<h1 class="text-2xl">Pong!</h1>`)
}

func NewHubro(vendorDir fs.FS) *Hubro {
	h := &Hubro{
		Mux: http.NewServeMux(),
		Server: &http.Server{
			Addr: ":8080",
		},
	}
	h.initTemplates()
	h.initStaticFiles()
	h.initVendorDir(vendorDir)
	h.Mux.HandleFunc("GET /", h.indexHandler)
	h.Mux.HandleFunc("GET /ping", h.pingHandler)
	return h
}

func (h *Hubro) Start() {
	h.Server.Handler = h.GetHandler()
	fmt.Println("Listening on http://localhost:8080")
	http.ListenAndServe(":8080", h.Server.Handler)
}
