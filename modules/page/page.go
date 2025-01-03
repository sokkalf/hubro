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
type indexedPage struct {
	path    string
	modTime time.Time
}

var indexedPages = make(map[*index.Index][]indexedPage)
var indexedPagesMutex sync.RWMutex

func slugify(s string) string {
	return slug.Make(s)
}

func parse(prefix string, h *server.Hubro, mux *http.ServeMux, md goldmark.Markdown, path string, opts PageOptions, isUpdate bool) error {
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

	var indexFunc func(index.IndexEntry) error
	if isUpdate {
		indexFunc = opts.Index.UpdateEntry
	} else {
		indexFunc = opts.Index.AddEntry
	}

	start := time.Now()
	name := strings.TrimSuffix(path, ".md")
	content, err := fs.ReadFile(opts.FilesDir, path)
	if err != nil {
		slog.Error("Error reading page file", "page", path, "error", err)
		return err
	}
	context := parser.NewContext()
	if err := md.Convert(content, &buf, parser.WithContext(context)); err != nil {
		slog.Error("Error converting markdown", "page", path, "error", err)
		return err
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
		Id:          path,
		Slug:        slug,
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
		FileName:    path,
	})
	if err != nil {
		slog.Warn("Error adding page to index", "page", name, "error", err, "index", opts.Index.GetName())
		return err
	}
	slog.Debug("Parsed page", "page", name, "title", title, "path", prefix+handlerPath, "duration", time.Since(start))
	return nil
}

func handler(h *server.Hubro, mux *http.ServeMux, index *index.Index) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		slug := strings.TrimPrefix(r.URL.Path, "/")
		entry := index.GetEntryBySlug(slug)
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

func scanMarkdownFiles(prefix string, h *server.Hubro, mux *http.ServeMux, opts PageOptions) int {
	md := goldmark.New(
		goldmark.WithExtensions(extension.GFM, meta.Meta),
		goldmark.WithParserOptions(parser.WithAutoHeadingID()),
		goldmark.WithRendererOptions(html.WithUnsafe()),
	)
	filesScanned := 0
	filesScannedList := make([]string, 0)
	fs.WalkDir(opts.FilesDir, ".", func(path string, d fs.DirEntry, err error) error {
		if !d.IsDir() && strings.HasSuffix(path, ".md") {
			fi, _ := d.Info()
			indexedPagesMutex.Lock()
			if indexedPages[opts.Index] == nil {
				slog.Debug("Initializing indexed pages", "index", opts.Index.GetName())
				indexedPages[opts.Index] = make([]indexedPage, 0)
			}
			modTime := fi.ModTime()
			var alreadyIndexed bool = false
			var isUpdate bool = false
			for _, p := range indexedPages[opts.Index] {
				if p.path == path && p.modTime == modTime {
					alreadyIndexed = true
				} else if p.path == path && p.modTime != modTime {
					isUpdate = true
				}
			}
			idxVal := indexedPage{path: path, modTime: modTime}
			indexedPagesMutex.Unlock()
			if !alreadyIndexed {
				var err error
				if isUpdate {
					err = parse(prefix, h, mux, md, path, opts, true)
				} else {
					err = parse(prefix, h, mux, md, path, opts, false)
				}
				if err != nil {
					slog.Error("Error parsing page", "page", path, "error", err)
				} else {
					if !isUpdate {
						indexedPagesMutex.Lock()
						indexedPages[opts.Index] = append(indexedPages[opts.Index], idxVal)
						filesScanned++
						indexedPagesMutex.Unlock()
					} else {
						indexedPagesMutex.Lock()
						for i, p := range indexedPages[opts.Index] {
							if p.path == path {
								indexedPages[opts.Index][i] = idxVal
							}
						}
						filesScanned++
						indexedPagesMutex.Unlock()
					}
				}
			}
			filesScannedList = append(filesScannedList, path)
		}
		return nil
	})
	deletedFiles := make([]string, 0)
	indexedPagesMutex.Lock()
	for _, p := range indexedPages[opts.Index] {
		var found bool = false
		for _, f := range filesScannedList {
			if p.path == f {
				found = true
			}
		}
		if !found {
			deletedFiles = append(deletedFiles, p.path)
		}
	}
	for _, f := range deletedFiles {
		slog.Debug("Removing deleted page", "page", f, "index", opts.Index.GetName())
		opts.Index.DeleteEntry(f)
		for i, p := range indexedPages[opts.Index] {
			if p.path == f {
				indexedPages[opts.Index] = append(indexedPages[opts.Index][:i], indexedPages[opts.Index][i+1:]...)
			}
		}
		filesScanned++
	}
	indexedPagesMutex.Unlock()

	return filesScanned
}

func Register(prefix string, h *server.Hubro, mux *http.ServeMux, options interface{}) {
	start := time.Now()
	opts, ok := options.(PageOptions)
	if !ok {
		slog.Error("Invalid options for page module")
	}
	scanMarkdownFiles(prefix, h, mux, opts)
	opts.Index.Sort()
	mux.HandleFunc("/", handler(h, mux, opts.Index))
	slog.Info("Registered pages", "duration", time.Since(start))
	slog.Debug("Scanning for new pages every 30 seconds", "index", opts.Index.GetName())

	go func() {
		for {
			time.Sleep(30 * time.Second)
			n := scanMarkdownFiles(prefix, h, mux, opts)
			if n > 0 {
				slog.Info("Found new or updated pages", "index", opts.Index.GetName(), "new", n)
				opts.Index.Sort()
				opts.Index.MsgBroker.Publish(index.Reset)
			}
		}
	}()
}
