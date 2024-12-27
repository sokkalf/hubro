package server

import (
	"bytes"
	"fmt"
	"html/template"
	"io/fs"
	"log/slog"
	"net/http"
	"os"
	"slices"
	"strings"
)

type Config struct {
	RootPath    string
	VendorDir   fs.FS
	TemplateDir fs.FS
	LayoutDir   fs.FS
	PublicDir   fs.FS
	PagesDir    fs.FS
}

type Middleware func(*Hubro) func(http.Handler) http.Handler

type Hubro struct {
	Mux         *http.ServeMux
	Server      *http.Server
	Layouts     *template.Template
	Templates   *template.Template
	PagesDir    fs.FS
	RootPath    string
	middlewares []Middleware
	publicDir   fs.FS
}

type HubroModule func(*Hubro, *http.ServeMux)

const (
	rootLayout           = "app.gohtml"
	errorLayout          = "errors/layout.gohtml"
	defaultErrorTemplate = "errors/default.gohtml"
)

var publicFileWhiteList = []string{"favicon.ico", "robots.txt", "sitemap.xml", "manifest.json", "apple-touch-icon.png",
	"security.txt", "android-chrome-192x192.png", "android-chrome-512x512.png", "browserconfig.xml", "site.webmanifest",
	"favicon.png", "favicon-16x16.png", "favicon-32x32.png", "favicon-96x96.png", "favicon-192x192.png", "favicon-512x512.png",
	"cache_manifest.json"}

var VendorLibs map[string]string = map[string]string{
	"htmx":      "/vendor/htmx/htmx.min.js",
	"alpine.js": "/vendor/alpine.js/alpine.min.js",
}

func (h *Hubro) Use(m Middleware) {
	h.middlewares = append(h.middlewares, m)
}

func (h *Hubro) AddModule(prefix string, module HubroModule) {
	h.createSubMux(prefix, module)
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

func (h *Hubro) createSubMux(prefix string, module HubroModule) *http.ServeMux {
	mux := http.NewServeMux()
	module(h, mux)
	h.Mux.Handle(prefix+"/", http.StripPrefix(prefix, mux))
	return mux
}

func (h *Hubro) initTemplates(layoutDir fs.FS, templateDir fs.FS, modTime int64) {
	defaultFuncMap := template.FuncMap{
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
			return fmt.Sprintf("%s/static/app.css?v=%d", strings.TrimSuffix(h.RootPath, "/"), modTime)
		},
		"vendor": func(path string) string {
			return strings.TrimSuffix(h.RootPath, "/") + VendorLibs[path]
		},
		"yield": func() (string, error) {
			// overwritten when rendering with layout
			return "", fmt.Errorf("yield called unexpectedly.")
		},
	}

	h.Layouts = template.New("root_layout")
	fs.WalkDir(layoutDir, ".", func(path string, d fs.DirEntry, err error) error {
		if !d.IsDir() && strings.HasSuffix(path, ".gohtml") {
			name := strings.TrimPrefix(path, "layouts/")
			content, err := fs.ReadFile(layoutDir, path)
			if err != nil {
				slog.Error("Error reading layout file", "layout", path, "error", err)
				panic(err)
			}
			h.Layouts = template.Must(h.Layouts.New(name).Funcs(defaultFuncMap).Parse(string(content)))
			slog.Debug("Parsed layout", "layout", h.Layouts.Name())
		}
		return nil
	})

	h.Templates = template.New("root")
	fs.WalkDir(templateDir, ".", func(path string, d fs.DirEntry, err error) error {
		if !d.IsDir() && strings.HasSuffix(path, ".gohtml") {
			name := strings.TrimPrefix(path, "templates/")
			content, err := fs.ReadFile(templateDir, path)
			if err != nil {
				slog.Error("Error reading template file", "template", path, "error", err)
				panic(err)
			}
			h.Templates = template.Must(h.Templates.New(name).Funcs(defaultFuncMap).Parse(string(content)))
			slog.Debug("Parsed template", "template", h.Templates.Name())
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
	fs := http.FileServer(http.Dir("./view/static"))
	h.Mux.Handle("GET /static/", http.StripPrefix("/static/", h.fileServerWithDirectoryListingDisabled(fs)))
}

func (h *Hubro) initVendorDir(vendorDir fs.FS) {
	fs := http.FileServer(http.FS(vendorDir))
	h.Mux.Handle("GET /vendor/", http.StripPrefix("/vendor/", h.fileServerWithDirectoryListingDisabled(fs)))
}

func (h *Hubro) indexHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/" {
		// Render the "index.gohtml" template
		h.Render(w, r, "index.gohtml", nil)
	} else {
		fs := http.FS(h.publicDir)
		if !slices.Contains(publicFileWhiteList, strings.TrimPrefix(r.URL.Path, "/")) {
			msg := "Page not found"
			h.ErrorHandler(w, r, http.StatusNotFound, &msg)
			return
		}
		file, err := fs.Open(r.URL.Path)
		if err != nil {
			msg := "Page not found"
			h.ErrorHandler(w, r, http.StatusNotFound, &msg)
			return
		} else {
			file.Close()
			http.FileServer(fs).ServeHTTP(w, r)
		}
	}
}

func (h *Hubro) testHandler(w http.ResponseWriter, r *http.Request) {
	h.Render(w, r, "test.gohtml", nil)
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
	data := struct {
		Status  int
		Message *string
	}{
		Status:  status,
		Message: message,
	}
	if h.Templates.Lookup(errorTemplate) == nil {
		slog.Warn("Error template for error status not found", "status", status, "template", errorTemplate)
		errorTemplate = defaultErrorTemplate
	}
	h.RenderWithLayout(w, r, errorLayout, errorTemplate, data)
	return
}

func (h *Hubro) RenderWithLayout(w http.ResponseWriter, r *http.Request, layoutName string, templateName string, data interface{}) {
	if data == nil {
		data = map[string]interface{}{}
	}
	funcs := template.FuncMap{
		"yield": func() (template.HTML, error) {
			buf := bytes.NewBuffer(nil)
			err := h.Templates.ExecuteTemplate(buf, templateName, data)
			return template.HTML(buf.String()), err
		},
	}
	layout := h.Layouts.Lookup(layoutName)
	if layout == nil {
		slog.Warn("Layout not found, falling back to default layout", "layout", layoutName)
		layout = h.Layouts.Lookup(rootLayout)
	}
	layoutClone, _ := layout.Clone()
	layoutClone.Funcs(funcs)
	err := layoutClone.Execute(w, data)
	if err != nil {
		slog.Error("can't render layout", "layout", layoutName, "error", err)
		http.Error(w, "Failed to render layout", http.StatusInternalServerError)
	}
}

func (h *Hubro) Render(w http.ResponseWriter, r *http.Request, templateName string, data interface{}) {
	h.RenderWithLayout(w, r, rootLayout, templateName, data)
}

func NewHubro(config Config) *Hubro {
	h := &Hubro{
		RootPath: config.RootPath,
		Mux:      http.NewServeMux(),
		Server: &http.Server{
			Addr: ":8080",
		},
		PagesDir:  config.PagesDir,
		publicDir: config.PublicDir,
	}
	assetModificationTime, err := os.Stat("view/static/app.css")
	if err != nil {
		slog.Error("Error getting file info, CSS file not found")
		panic(err)
	}
	h.initTemplates(config.LayoutDir, config.TemplateDir, assetModificationTime.ModTime().Unix())
	h.initStaticFiles()
	h.initVendorDir(config.VendorDir)
	h.Mux.HandleFunc("/", h.indexHandler)
	h.Mux.HandleFunc("GET /ping", h.pingHandler)
	h.Mux.HandleFunc("GET /test", h.testHandler)
	return h
}

func (h *Hubro) Start() {
	h.Server.Handler = h.GetHandler()
	port := 8080
	slog.Info("Server started", "port", port, "rootPath", h.RootPath)
	http.ListenAndServe(fmt.Sprintf(":%d", port), h.Server.Handler)
}
