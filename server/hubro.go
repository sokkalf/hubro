package server

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"path/filepath"
	"strings"
)

type middleware func (*Hubro) func(http.Handler) http.Handler

type Hubro struct {
	Mux *http.ServeMux
	Server *http.Server
	Templates *template.Template
	middlewares []middleware
}

func (h *Hubro) Use(m middleware) {
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
	}
	h.Templates, err = template.New("root").Funcs(funcMap).ParseGlob(filepath.Join("templates", "*.gohtml"))
	if err != nil {
		log.Fatalf("Error parsing templates: %v", err)
	}
	for _, t := range h.Templates.Templates() {
		fmt.Printf("Parsed template: %s\n", t.Name())
	}
}

func (h *Hubro) initStaticFiles() {
	fs := http.FileServer(http.Dir("./static"))
	fsWithDirectoryListingDisabled := func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if strings.HasSuffix(r.URL.Path, "/") || r.URL.Path == "" {
				http.Error(w, "403 directory listing not allowed", http.StatusForbidden)
				//fmt.Fprintf(w, "403 directory listing not allowed")
				return
			}
			h.ServeHTTP(w, r)
		})
	}
	h.Mux.Handle("/static/", http.StripPrefix("/static/", fsWithDirectoryListingDisabled(fs)))
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
	fmt.Fprintln(w, "Pong!")
}

func NewHubro() *Hubro {
	h := &Hubro{
		Mux: http.NewServeMux(),
		Server: &http.Server{
			Addr: ":8080",
		},
	}
	h.Mux.HandleFunc("/", h.indexHandler)
	h.Mux.HandleFunc("/ping", h.pingHandler)
	h.initTemplates()
	h.initStaticFiles()
	return h
}

func (h *Hubro) Start() {
	h.Server.Handler = h.GetHandler()
	fmt.Println("Listening on http://localhost:8080")
	http.ListenAndServe(":8080", h.Server.Handler)
}
