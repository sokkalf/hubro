package page

import (
	"bytes"
	"html/template"
	"io/fs"
	"log/slog"
	"net/http"
	"strings"
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

func Register(h *server.Hubro, mux *http.ServeMux, options interface{}) {
	filesDir := options.(struct{ FilesDir fs.FS }).FilesDir
	md := goldmark.New(
		goldmark.WithExtensions(extension.GFM, meta.Meta),
		goldmark.WithParserOptions(parser.WithAutoHeadingID()),
		goldmark.WithRendererOptions(html.WithUnsafe()),
	)
	fs.WalkDir(filesDir, ".", func(path string, d fs.DirEntry, err error) error {
		if !d.IsDir() && strings.HasSuffix(path, ".md") {
			start := time.Now()
			name := strings.TrimSuffix(path, ".md")
			content, err := fs.ReadFile(filesDir, path)
			if err != nil {
				slog.Error("Error reading page file", "page", path, "error", err)
				goto next
			}
			var buf bytes.Buffer
			if err := md.Convert(content, &buf); err != nil {
				slog.Error("Error converting markdown", "page", path, "error", err)
				goto next
			}
			path := "/" + slugify(name)
			mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
				h.Render(w, r, "page.gohtml", struct {
					Title string
					Body  template.HTML
				}{
					Title: name,
					Body:  template.HTML(buf.String()),
				})
			})
			slog.Debug("Parsed page", "page", name, "path", path, "duration", time.Since(start))
		}
	next:
		return nil
	})
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		msg := "Page not found"
		h.ErrorHandler(w, r, http.StatusNotFound, &msg)
		return
	})
}
