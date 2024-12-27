package page

import (
	"bytes"
	"html/template"
	"io/fs"
	"log/slog"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gosimple/slug"
	"github.com/yuin/goldmark"
	meta "github.com/yuin/goldmark-meta"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer/html"

	"github.com/sokkalf/hubro/server"
)

func slugify(s string) string {
	return slug.Make(s)
}

func parse(h *server.Hubro, mux *http.ServeMux, md goldmark.Markdown, path string, filesDir fs.FS, indexFunc func(server.IndexEntry)) {
	var handlerPath string
	var title string
	var author string
	var visible bool = true
	var sortOrder int
	var metaData map[string]interface{}
	var buf bytes.Buffer
	var context parser.Context

	start := time.Now()
	name := strings.TrimSuffix(path, ".md")
	content, err := fs.ReadFile(filesDir, path)
	if err != nil {
		slog.Error("Error reading page file", "page", path, "error", err)
		goto next
	}
	context = parser.NewContext()
	if err := md.Convert(content, &buf, parser.WithContext(context)); err != nil {
		slog.Error("Error converting markdown", "page", path, "error", err)
		goto next
	}
	metaData = meta.Get(context)
	if t, ok := metaData["title"]; ok {
		title = t.(string)
	} else {
		title = name
	}
	if a, ok := metaData["author"]; ok {
		author = a.(string)
	}
	if v, ok := metaData["visible"]; ok {
		visible = v.(bool)
	}
	if s, ok := metaData["sortOrder"]; ok {
		sortOrder = s.(int)
	} else {
		sortOrder = 0
	}

	handlerPath = "/" + slugify(title)
	indexFunc(server.IndexEntry{
		Title: title,
		Author: author,
		Visible: visible,
		Metadata: metaData,
		Path: handlerPath,
		SortOrder: sortOrder,
	})
	mux.HandleFunc(handlerPath, func(w http.ResponseWriter, r *http.Request) {
		h.Render(w, r, "page.gohtml", struct {
			Title string
			Author string
			Visible bool
			Body  template.HTML
			Metadata map[string]interface{}
		}{
			Title: title,
			Author: author,
			Visible: visible,
			Body:  template.HTML(buf.String()),
			Metadata: metaData,
		})
	})
	slog.Debug("Parsed page", "page", name, "title", title, "path", handlerPath, "duration", time.Since(start))
	next:
}

func Register(h *server.Hubro, mux *http.ServeMux, options interface{}) {
	start := time.Now()
	var wg sync.WaitGroup
	opts := options.(struct{
		FilesDir fs.FS
		IndexFunc func(server.IndexEntry)
	})
	filesDir := opts.FilesDir
	indexFunc := opts.IndexFunc
	md := goldmark.New(
		goldmark.WithExtensions(extension.GFM, meta.Meta),
		goldmark.WithParserOptions(parser.WithAutoHeadingID()),
		goldmark.WithRendererOptions(html.WithUnsafe()),
	)
	fs.WalkDir(filesDir, ".", func(path string, d fs.DirEntry, err error) error {
		if !d.IsDir() && strings.HasSuffix(path, ".md") {
			wg.Add(1)
			go func() {
				defer wg.Done()
				parse(h, mux, md, path, filesDir, indexFunc)
			}()
		}
		return nil
	})
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		msg := "Page not found"
		h.ErrorHandler(w, r, http.StatusNotFound, &msg)
		return
	})
	wg.Wait()
	slog.Debug("Registered pages", "duration", time.Since(start))
}
