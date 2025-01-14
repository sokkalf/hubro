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

func getOrDefault[T any](m map[string]interface{}, key string, defaultVal T) T {
	raw, ok := m[key]
	if !ok {
		return defaultVal
	}
	val, ok := raw.(T)
	if !ok {
		return defaultVal
	}
	delete(m, key)
	return val
}

func parse(prefix string, md goldmark.Markdown, path string, opts PageOptions, isUpdate bool) error {
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
	title := getOrDefault(metaData, "title", name)
	shortTitle := getOrDefault(metaData, "shortTitle", title)
	description := getOrDefault(metaData, "description", "")
	author := getOrDefault(metaData, "author", "")
	visible := getOrDefault(metaData, "visible", true)
	sortOrder := getOrDefault(metaData, "sortOrder", 0)
	hideAuthor := getOrDefault(metaData, "hideAuthor", false)
	hideTitle := getOrDefault(metaData, "hideTitle", false)

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
		ShortTitle:  shortTitle,
		Description: description,
		Author:      author,
		Visible:     visible,
		Metadata:    metaData,
		Path:        handlerPath,
		SortOrder:   sortOrder,
		HideAuthor:  hideAuthor,
		HideTitle:   hideTitle,
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

func handler(h *server.Hubro, index *index.Index) http.HandlerFunc {
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

func scanMarkdownFiles(prefix string, opts PageOptions) (filesScanned, numNew, numUpdated, numDeleted int) {
	md := goldmark.New(
		goldmark.WithExtensions(extension.GFM, meta.Meta),
		goldmark.WithParserOptions(parser.WithAutoHeadingID()),
		goldmark.WithRendererOptions(html.WithUnsafe()),
	)
	filesScannedList := make([]string, 0)
	fs.WalkDir(opts.FilesDir, ".", func(path string, d fs.DirEntry, err error) error {
		if !d.IsDir() && strings.HasSuffix(path, ".md") {
			fi, err := d.Info()
			if err != nil {
				slog.Error("Error getting file info", "file", path, "error", err)
				return err
			}
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
				err := parse(prefix, md, path, opts, isUpdate)
				if err != nil {
					slog.Error("Error parsing page", "page", path, "error", err)
				} else {
					if !isUpdate {
						indexedPagesMutex.Lock()
						indexedPages[opts.Index] = append(indexedPages[opts.Index], idxVal)
						filesScanned++
						numNew++
						indexedPagesMutex.Unlock()
					} else {
						indexedPagesMutex.Lock()
						for i, p := range indexedPages[opts.Index] {
							if p.path == path {
								indexedPages[opts.Index][i] = idxVal
							}
						}
						filesScanned++
						numUpdated++
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
				numDeleted++
			}
		}
		filesScanned++
	}
	indexedPagesMutex.Unlock()

	return filesScanned, numNew, numUpdated, numDeleted
}

func Register(prefix string, h *server.Hubro, mux *http.ServeMux, options interface{}) {
	start := time.Now()
	opts, ok := options.(PageOptions)
	if !ok {
		slog.Error("Invalid options for page module")
	}
	scanMarkdownFiles(prefix, opts)
	opts.Index.Sort()
	mux.HandleFunc("/", handler(h, opts.Index))
	slog.Info("Registered pages", "duration", time.Since(start))

	go func() {
		msgChan := opts.Index.MsgBroker.Subscribe()
		for {
			switch <-msgChan {
			case index.Scanned:
				start := time.Now()
				n, nNew, nUpdated, nDeleted := scanMarkdownFiles(prefix, opts)
				if n > 0 {
					slog.Info("Found new or updated pages", "index", opts.Index.GetName(),
						"new", nNew, "updated", nUpdated, "deleted", nDeleted, "duration", time.Since(start))
					opts.Index.Sort()
					opts.Index.MsgBroker.Publish(index.Updated)
				}
			default: // Ignore other messages
			}
		}
	}()
}
