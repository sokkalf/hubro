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
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/sokkalf/hubro/cache"
	hc "github.com/sokkalf/hubro/config"
	"github.com/sokkalf/hubro/helpers"
	"github.com/sokkalf/hubro/index"
)

type Config struct {
	RootPath    string
	Port        int
	VendorDir   fs.FS
	TemplateDir fs.FS
	LayoutDir   fs.FS
	PublicDir   fs.FS
}

type Middleware func(*Hubro) func(http.Handler) http.Handler

type Hubro struct {
	Mux         *http.ServeMux
	Server      *http.Server
	Templates   *template.Template
	config      hc.HubroConfig
	middlewares []Middleware
	publicDir   fs.FS
}

type HubroModule func(string, *Hubro, *http.ServeMux, interface{})

const (
	rootLayout           = "app"
	errorLayout          = "errors/layout"
	defaultErrorTemplate = "errors/default"
)

var publicFileWhiteList = []string{"favicon.ico", "robots.txt", "sitemap.xml", "manifest.json", "apple-touch-icon.png",
	"security.txt", "android-chrome-192x192.png", "android-chrome-512x512.png", "browserconfig.xml", "site.webmanifest",
	"favicon.png", "favicon-16x16.png", "favicon-32x32.png", "favicon-96x96.png", "favicon-192x192.png", "favicon-512x512.png",
	"cache_manifest.json"}

var VendorLibs map[string]string = map[string]string{
	"htmx":         "/vendor/htmx/htmx.min.js",
	"alpine.js":    "/vendor/alpine.js/alpine.min.js",
	"highlight.js": "/vendor/highlight/highlight.min.js",
}

var addTemplateMutex sync.Mutex

func (h *Hubro) Use(m Middleware) {
	h.middlewares = append(h.middlewares, m)
}

func (h *Hubro) AddModule(prefix string, module HubroModule, options interface{}) {
	h.createSubMux(prefix, module, options)
}

func (h *Hubro) handlerWithMiddlewares(handler http.Handler) http.Handler {
	for _, m := range h.middlewares {
		handler = m(h)(handler)
	}
	return handler
}

// GetHandler returns the http.Handler for the Hubro instance with all middlewares applied.
func (h *Hubro) GetHandler() http.Handler {
	return h.handlerWithMiddlewares(h.Mux)
}

func (h *Hubro) createSubMux(prefix string, module HubroModule, options interface{}) *http.ServeMux {
	var mux *http.ServeMux
	if prefix == "" || prefix == "/" {
		prefix = "" // strip trailing slash if present
		mux = h.Mux
	} else {
		mux = http.NewServeMux()
		h.Mux.Handle(prefix+"/", http.StripPrefix(prefix, mux))
	}
	module(prefix, h, mux, options)
	return mux
}

func (h *Hubro) initTemplates(layoutDir fs.FS, templateDir fs.FS, modTime int64) {
	defaultFuncMap := template.FuncMap{
		"appTitle": func() string {
			return h.config.Title
		},
		"rootPath": func() string {
			return strings.TrimSuffix(h.config.RootPath, "/")
		},
		"staticPath": func(path string) string {
			return strings.TrimSuffix(h.config.RootPath, "/") + "/static/" + path
		},
		"vendorPath": func(path string) string {
			return strings.TrimSuffix(h.config.RootPath, "/") + "/vendor/" + path
		},
		"appCSS": func() string {
			return fmt.Sprintf("%s/static/app.css?v=%d", strings.TrimSuffix(h.config.RootPath, "/"), modTime)
		},
		"vendor": func(path string) string {
			return strings.TrimSuffix(h.config.RootPath, "/") + VendorLibs[path]
		},
		"yield": func() (string, error) {
			// overwritten when rendering with layout
			slog.Warn("yield called unexpectedly")
			return "", fmt.Errorf("yield called unexpectedly.")
		},
		"boosted": func() bool {
			return false
		},
		"format_date": func(date time.Time) string {
			return date.Format("2006-01-02")
		},
		"listPages": func(i string, filterTag string) []index.IndexEntry {
			entries := index.GetIndex(i)
			if entries == nil {
				return []index.IndexEntry{}
			} else {
				if filterTag == "" {
					return entries.Entries
				} else {
					var filteredEntries []index.IndexEntry
					for _, entry := range entries.Entries {
						if slices.Contains(entry.Tags, filterTag) {
							filteredEntries = append(filteredEntries, entry)
						}
					}
					return filteredEntries
				}
			}
		},
		"paginate": func(page int, entries []index.IndexEntry) []index.IndexEntry {
			perPage := hc.Config.PostsPerPage
			start := (page - 1) * perPage
			end := start + perPage
			if start > len(entries) {
				return []index.IndexEntry{}
			}
			if end > len(entries) {
				end = len(entries)
			}
			return entries[start:end]
		},
		"paginator": func(page int, entries []index.IndexEntry) template.HTML {
			slog.Warn("paginator called unexpectedly")
			return template.HTML("")
		},
		"getConfig": func() hc.HubroConfig {
			return h.config
		},
		"tagCloud": func(i string) template.HTML {
			entries := index.GetIndex(i)
			return helpers.GenerateTagCloud(entries)
		},
		"add": func(a, b int) int {
			return a + b
		},
	}

	h.Templates = template.New("root")
	fs.WalkDir(layoutDir, ".", func(path string, d fs.DirEntry, err error) error {
		if !d.IsDir() && strings.HasSuffix(path, ".gohtml") {
			go func() {
				start := time.Now()
				name := strings.TrimPrefix(path, "layouts/")
				name = strings.TrimSuffix(name, ".gohtml")
				content, err := fs.ReadFile(layoutDir, path)
				if err != nil {
					slog.Error("Error reading layout file", "layout", path, "error", err)
					panic(err)
				}
				addTemplateMutex.Lock()
				h.Templates = template.Must(h.Templates.New(name).Funcs(defaultFuncMap).Parse(string(content)))
				slog.Debug("Parsed layout", "layout", h.Templates.Name(), "duration", time.Since(start))
				addTemplateMutex.Unlock()
			}()
		}
		return nil
	})

	fs.WalkDir(templateDir, ".", func(path string, d fs.DirEntry, err error) error {
		if !d.IsDir() && strings.HasSuffix(path, ".gohtml") {
			go func() {
				start := time.Now()
				name := strings.TrimPrefix(path, "templates/")
				name = strings.TrimSuffix(name, ".gohtml")
				content, err := fs.ReadFile(templateDir, path)
				if err != nil {
					slog.Error("Error reading template file", "template", path, "error", err)
					panic(err)
				}
				addTemplateMutex.Lock()
				h.Templates = template.Must(h.Templates.New(name).Funcs(defaultFuncMap).Parse(string(content)))
				slog.Debug("Parsed template", "template", h.Templates.Name(), "duration", time.Since(start))
				addTemplateMutex.Unlock()
			}()
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
		w.Header().Set("Cache-Control", "public, max-age=31536000")
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
		tag := ""
		page := 1
		query := r.URL.Query()
		if query.Get("tag") != "" {
			tag = query.Get("tag")
		}
		if query.Get("p") != "" {
			var err error
			page, err = strconv.Atoi(query.Get("p"))
			if err != nil {
				page = 1
			}
		} else {
			page = 1
		}

		// Render the "index.gohtml" template
		h.Render(w, r, "blogindex", struct {
			FilterByTag string
			Title       string
			Page        int
		}{
			FilterByTag: tag,
			Title:       h.config.Description,
			Page:        page,
		})
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
			w.Header().Set("Cache-Control", "public, max-age=31536000")
			http.FileServer(fs).ServeHTTP(w, r)
		}
	}
}

func (h *Hubro) testHandler(w http.ResponseWriter, r *http.Request) {
	h.Render(w, r, "test", nil)
}

// pingHandler is a simple route that returns "Pong!" text.
// We set an HTMX response header ("HX-Trigger") to demonstrate sending events back to the client.
func (h *Hubro) pingHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("HX-Trigger", "pongReceived")
	fmt.Fprintln(w, `<h1 class="text-2xl">Pong!</h1>`)
}

func (h *Hubro) ErrorHandler(w http.ResponseWriter, r *http.Request, status int, message *string) {
	w.WriteHeader(status)
	errorTemplate := fmt.Sprintf("errors/%d", status)
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
	cacheKey := fmt.Sprintf("%s-%s", layoutName, templateName)
	if t := cache.Get(cacheKey); t != nil {
		t.ExecuteTemplate(w, layoutName, data)
		var err error
		if err != nil {
			slog.Error("can't render cached template", "template", cacheKey, "error", err)
			http.Error(w, "Failed to render cached template", http.StatusInternalServerError)
		}
		return
	}
	clone, err := h.Templates.Clone()
	if err != nil {
		slog.Error("can't clone templates", "error", err)
		http.Error(w, "Failed to render layout", http.StatusInternalServerError)
		return
	}

	funcs := template.FuncMap{
		"yield": func() (template.HTML, error) {
			buf := bytes.NewBuffer(nil)
			err := clone.ExecuteTemplate(buf, templateName, data)
			return template.HTML(buf.String()), err
		},
		"boosted": func() bool {
			return r.Header.Get("HX-Boosted") == "true"
		},
		"paginator": func(page int, entries []index.IndexEntry) template.HTML {
			totalPages := (len(entries) + hc.Config.PostsPerPage - 1) / hc.Config.PostsPerPage
			return helpers.Paginator(r.URL, page, totalPages, entries)
		},
	}
	clone.Funcs(funcs)
	cache.Put(cacheKey, clone)
	err = clone.ExecuteTemplate(w, layoutName, data)
	if err != nil {
		slog.Error("can't render layout", "layout", layoutName, "error", err)
		http.Error(w, "Failed to render layout", http.StatusInternalServerError)
	}
}

func (h *Hubro) RenderWithoutLayout(w http.ResponseWriter, r *http.Request, templateName string, data interface{}) {
	if data == nil {
		data = map[string]interface{}{}
	}
	clone, err := h.Templates.Clone()
	if err != nil {
		slog.Error("can't clone templates", "error", err)
		http.Error(w, "Failed to render template", http.StatusInternalServerError)
		return
	}
	err = clone.ExecuteTemplate(w, templateName, data)
	if err != nil {
		slog.Error("can't render template", "template", templateName, "error", err)
		http.Error(w, "Failed to render template", http.StatusInternalServerError)
	}
}

func (h *Hubro) Render(w http.ResponseWriter, r *http.Request, templateName string, data interface{}) {
	h.RenderWithLayout(w, r, rootLayout, templateName, data)
}

func NewHubro(config Config) *Hubro {
	h := &Hubro{
		config: *hc.Config,
		Mux:    http.NewServeMux(),
		Server: &http.Server{
			Addr: fmt.Sprintf(":%d", config.Port),
		},
		publicDir: config.PublicDir,
	}
	cssFileName := "view/static/app.css"
	assetModificationTime, err := os.Stat(cssFileName)
	if err != nil {
		slog.Error("Error getting file info, CSS file not found", "filename", cssFileName)
		panic(err)
	}
	go func() {
		h.initTemplates(config.LayoutDir, config.TemplateDir, assetModificationTime.ModTime().Unix())
	}()
	go func() {
		h.initStaticFiles()
	}()
	go func() {
		h.initVendorDir(config.VendorDir)
	}()
	h.Mux.HandleFunc("/", h.indexHandler)
	h.Mux.HandleFunc("GET /ping", h.pingHandler)
	h.Mux.HandleFunc("GET /test", h.testHandler)
	return h
}

func (h *Hubro) Start(startTime time.Time) {
	h.Server.Handler = h.GetHandler()
	port := h.config.Port
	slog.Info("Server started", "port", port, "rootPath", h.config.RootPath, "duration", time.Since(startTime))
	http.ListenAndServe(fmt.Sprintf(":%d", port), h.Server.Handler)
}
