package server

import (
	"fmt"
	"html/template"
	"io/fs"
	"log"
	"log/slog"
	"net/http"
	"os"
	"strings"
)
type Config struct {
	RootPath string
	VendorDir fs.FS
}

type Middleware func(*Hubro) func(http.Handler) http.Handler

type Hubro struct {
	Mux         *http.ServeMux
	Server      *http.Server
	Templates   *template.Template
	RootPath    string
	middlewares []Middleware
}

var VendorLibs map[string]string = map[string]string{
	"htmx": "/vendor/htmx/htmx.min.js",
	"alpine.js": "/vendor/alpine.js/alpine.min.js",
}

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
	funcMap := template.FuncMap{
		"title": func() string {
			return "Hubro"
		},
		"staticPath": func(path string) string {
			return strings.TrimSuffix(h.RootPath, "/") + "/static/" + path
		},
		"vendorPath": func(path string) string {
			return strings.TrimSuffix(h.RootPath, "/") + "/vendor/" + path
		},
		"appCSS": func() string {
			stat, err := os.Stat("static/app.css")
			if err != nil {
				log.Fatalf("Error getting file info, CSS file not found : %v", err)
			}
			return fmt.Sprintf("%s/static/app.css?v=%d", strings.TrimSuffix(h.RootPath, "/"), stat.ModTime().Unix())
		},
		"vendor": func(path string) string {
			return strings.TrimSuffix(h.RootPath, "/") + VendorLibs[path]
		},
	}

	h.Templates = template.New("root")
	templateDir := os.DirFS("templates")
	fs.WalkDir(templateDir, ".", func(path string, d fs.DirEntry, err error) error {
		if !d.IsDir() && strings.HasSuffix(path, ".gohtml") {
			name := strings.TrimPrefix(path, "templates/")
			content, err := fs.ReadFile(templateDir, path)
			h.Templates, err = h.Templates.New(name).Funcs(funcMap).Parse(string(content))
			if err != nil {
				log.Fatalf("Error parsing template: %v", err)
			} else {
				slog.Debug("Parsed template", "template", h.Templates.Name())
			}
		}
		return nil
	})
}

func (hu *Hubro) fileServerWithDirectoryListingDisabled(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/") || r.URL.Path == "" {
			msg := "Directory listing is disallowed"
			hu.ErrorHandler(w, r, http.StatusForbidden, &msg)
			return
		}
		h.ServeHTTP(w, r)
	})
}

func (h *Hubro) initStaticFiles() {
	fs := http.FileServer(http.Dir("./static"))
	h.Mux.Handle("GET /static/", http.StripPrefix("/static/", h.fileServerWithDirectoryListingDisabled(fs)))
}

func (h *Hubro) initVendorDir(vendorDir fs.FS) {
	fs := http.FileServer(http.FS(vendorDir))
	h.Mux.Handle("GET /vendor/", http.StripPrefix("/vendor/", h.fileServerWithDirectoryListingDisabled(fs)))
}

func (h *Hubro) indexHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		msg := "Page not found"
		h.ErrorHandler(w, r, http.StatusNotFound, &msg)
		return
	}
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

func (h *Hubro) ErrorHandler(w http.ResponseWriter, r *http.Request, status int, message *string) {
	w.WriteHeader(status)
	errorTemplate := fmt.Sprintf("errors/%d.gohtml", status)
	err := h.Templates.ExecuteTemplate(w, errorTemplate, struct {
		Status  int
		Message *string
	}{
		Status:  status,
		Message: message,
	})
	if err != nil {
		slog.Error("can't render template for error", "status", status, "error", err)
		fmt.Fprintf(w, "Error %d\n", status)
	}
	return
}

func NewHubro(config Config) *Hubro {
	h := &Hubro{
		RootPath: config.RootPath,
		Mux: http.NewServeMux(),
		Server: &http.Server{
			Addr: ":8080",
		},
	}
	h.initTemplates()
	h.initStaticFiles()
	h.initVendorDir(config.VendorDir)
	h.Mux.HandleFunc("GET /", h.indexHandler)
	h.Mux.HandleFunc("GET /ping", h.pingHandler)
	return h
}

func (h *Hubro) Start() {
	h.Server.Handler = h.GetHandler()
	port := 8080
	slog.Info("Server started", "port", port, "rootPath", h.RootPath)
	http.ListenAndServe(fmt.Sprintf(":%d", port), h.Server.Handler)
}
