package page

import (
	"bytes"
	"html/template"
	"io/fs"
	"log/slog"
	"net/http"
	"strings"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark-meta"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer/html"

	"github.com/sokkalf/hubro/server"
)

func Register(h *server.Hubro, mux *http.ServeMux) {
	md := goldmark.New(
		goldmark.WithExtensions(extension.GFM, meta.Meta),
		goldmark.WithParserOptions(parser.WithAutoHeadingID()),
		goldmark.WithRendererOptions(html.WithUnsafe()),
	)
	fs.WalkDir(h.PagesDir, ".", func(path string, d fs.DirEntry, err error) error {
		if !d.IsDir() && strings.HasSuffix(path, ".md") {
			name := strings.TrimSuffix(path, ".md")
			content, err := fs.ReadFile(h.PagesDir, path)
			if err != nil {
				slog.Error("Error reading page file", "page", path, "error", err)
				goto next
			}
			var buf bytes.Buffer
			if err := md.Convert(content, &buf); err != nil {
				slog.Error("Error converting markdown", "page", path, "error", err)
				goto next
			}
			mux.HandleFunc("/"+name, func(w http.ResponseWriter, r *http.Request) {
				h.Render(w, r, "page.gohtml", struct {
					Title string
					Body  template.HTML
				}{
					Title: name,
					Body:  template.HTML(buf.String()),
				})
			})
			slog.Debug("Parsed page", "page", name)
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
