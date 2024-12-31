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

	"github.com/sokkalf/hubro/index"
	"github.com/sokkalf/hubro/server"
	"github.com/sokkalf/hubro/utils"
)

type PageOptions struct {
	FilesDir fs.FS
	Index    *index.Index
}

func slugify(s string) string {
	return slug.Make(s)
}

func parse(prefix string, h *server.Hubro, mux *http.ServeMux, md goldmark.Markdown, path string, opts PageOptions) {
	var title string
	var description string
	var author string
	var visible bool = true
	var hideAuthor bool = false
	var sortOrder int
	var tags []string
	var summary *template.HTML
	var body *template.HTML
	var date time.Time
	var buf bytes.Buffer
	indexFunc := opts.Index.AddEntry

	start := time.Now()
	name := strings.TrimSuffix(path, ".md")
	content, err := fs.ReadFile(opts.FilesDir, path)
	if err != nil {
		slog.Error("Error reading page file", "page", path, "error", err)
		return
	}
	context := parser.NewContext()
	if err := md.Convert(content, &buf, parser.WithContext(context)); err != nil {
		slog.Error("Error converting markdown", "page", path, "error", err)
		return
	}
	metaData := meta.Get(context)
	if t, ok := metaData["title"]; ok {
		title = t.(string)
		delete(metaData, "title")
	} else {
		title = name
	}
	if d, ok := metaData["description"]; ok {
		description = d.(string)
		delete(metaData, "description")
	}
	if a, ok := metaData["author"]; ok {
		author = a.(string)
		delete(metaData, "author")
	}
	if v, ok := metaData["visible"]; ok {
		visible = v.(bool)
		delete(metaData, "visible")
	}
	if s, ok := metaData["sortOrder"]; ok {
		sortOrder = s.(int)
		delete(metaData, "sortOrder")
	} else {
		sortOrder = 0
	}
	if h, ok := metaData["hideAuthor"]; ok {
		hideAuthor = h.(bool)
		delete(metaData, "hideAuthor")
	}
	if t, ok := metaData["tags"]; ok {
		tags = make([]string, 0)
		for _, tag := range t.([]interface{}) {
			tags = append(tags, tag.(string))
		}
		delete(metaData, "tags")
	} else {
		tags = []string{}
	}
	if d, ok := metaData["date"]; ok {
		date = utils.ParseDate(d.(string))
		delete(metaData, "date")
	} else {
		date = time.Time{}
	}

	b := template.HTML(buf.String())
	body = &b

	var sum template.HTML
	s := strings.SplitN(buf.String(), "<!--more-->", 2)[0]
	if s == "" {
		// If there is no summary, use the body as the summary
		sum = b
	} else {
		sum = template.HTML(s)
	}
	summary = &sum

	slug := slugify(title)
	handlerPath := "/" + slugify(title)
	err = indexFunc(index.IndexEntry{
		Id:          slug,
		Title:       title,
		Description: description,
		Author:      author,
		Visible:     visible,
		Metadata:    metaData,
		Path:        handlerPath,
		SortOrder:   sortOrder,
		HideAuthor:  hideAuthor,
		Tags:        tags,
		Date:        date,
		Summary:     summary,
		Body:        body,
	})
	if err != nil {
		slog.Warn("Error adding page to index", "page", name, "error", err)
		return
	}
	slog.Debug("Parsed page", "page", name, "title", title, "path", prefix+handlerPath, "duration", time.Since(start))
}

func handler(h *server.Hubro, mux *http.ServeMux, index *index.Index) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		slug := strings.TrimPrefix(r.URL.Path, "/")
		entry := index.GetEntry(slug)
		if entry != nil {
			h.Render(w, r, "page", entry)
			return
		} else {
			msg := "Page not found"
			h.ErrorHandler(w, r, http.StatusNotFound, &msg)
			return
		}
	}
}

func scanMarkdownFiles(prefix string, h *server.Hubro, mux *http.ServeMux, opts PageOptions) {
	var wg sync.WaitGroup
	md := goldmark.New(
		goldmark.WithExtensions(extension.GFM, meta.Meta),
		goldmark.WithParserOptions(parser.WithAutoHeadingID()),
		goldmark.WithRendererOptions(html.WithUnsafe()),
	)
	fs.WalkDir(opts.FilesDir, ".", func(path string, d fs.DirEntry, err error) error {
		if !d.IsDir() && strings.HasSuffix(path, ".md") {
			wg.Add(1)
			go func() {
				defer wg.Done()
				parse(prefix, h, mux, md, path, opts)
			}()
		}
		return nil
	})
	wg.Wait()
}

func Register(prefix string, h *server.Hubro, mux *http.ServeMux, options interface{}) {
	start := time.Now()
	opts, ok := options.(PageOptions)
	if !ok {
		slog.Error("Invalid options for page module")
	}
	scanMarkdownFiles(prefix, h, mux, opts)
	mux.HandleFunc("/", handler(h, mux, opts.Index))
	slog.Info("Registered pages", "duration", time.Since(start))
}
