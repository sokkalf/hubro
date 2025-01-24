package admin

import (
	"bytes"
	"context"
	"encoding/json"
	"io/fs"
	"log/slog"
	"net/http"
	"os"

	"github.com/coder/websocket"
	"github.com/sokkalf/hubro/config"
	"github.com/sokkalf/hubro/index"
	"github.com/sokkalf/hubro/modules/page"
	"github.com/sokkalf/hubro/server"
	meta "github.com/yuin/goldmark-meta"
	"github.com/yuin/goldmark/parser"
)

func basicAuth(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		username, password, ok := r.BasicAuth()
		if !ok || username != "admin" || password != config.Config.AdminPassword {
			w.Header().Set("WWW-Authenticate", `Basic realm="Restricted"`)
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		h(w, r)
	}
}

func renderMarkdown(markdown []byte) ([]byte, map[string]interface{}, error) {
	md := page.GetMarkdownParser()
	var buf bytes.Buffer
	context := parser.NewContext()

	err := md.Convert(markdown, &buf, parser.WithContext(context))
	if err != nil {
		return nil, nil, err
	}

	metaData := meta.Get(context)
	return buf.Bytes(), metaData, nil
}

func Register(prefix string, h *server.Hubro, mux *http.ServeMux, options interface{}) {
	slog.Info("Registering admin module")

	indices := make([]*index.Index, 0)
	for _, i := range index.GetIndices() {
		indices = append(indices, i)
	}

	mux.Handle("/", basicAuth(func(w http.ResponseWriter, r *http.Request) {
		h.RenderWithLayout(w, r, "admin/app", "admin/index", indices)
	}))
	mux.Handle("/edit", basicAuth(func(w http.ResponseWriter, r *http.Request) {
		slug := r.URL.Query().Get("p")
		idx := r.URL.Query().Get("idx")
		if idx == "" {
			msg := "Index not found"
			h.ErrorHandler(w, r, http.StatusNotFound, &msg)
			return
		}
		i := index.GetIndex(idx)
		if i == nil {
			msg := "Index not found"
			h.ErrorHandler(w, r, http.StatusNotFound, &msg)
			return
		}
		entry := i.GetEntryBySlug(slug)
		if entry == nil {
			msg := "Entry not found"
			h.ErrorHandler(w, r, http.StatusNotFound, &msg)
			return
		}
		fileContent, err := fs.ReadFile(i.FilesDir, entry.FileName)
		if err != nil {
			msg := "Error reading file"
			slog.Error("Error reading file", "error", err)
			h.ErrorHandler(w, r, http.StatusInternalServerError, &msg)
			return
		}

		data := struct {
			Entry *index.IndexEntry
			RawContent string
		}{entry, string(fileContent)}

		h.RenderWithLayout(w, r, "admin/app", "admin/edit", data)
	}))
	mux.Handle("/ws", basicAuth(func(w http.ResponseWriter, r *http.Request) {
		conn, err := websocket.Accept(w, r, nil)
		if err != nil {
			slog.Error("Error accepting websocket connection", "error", err)
			return
		}
		defer conn.CloseNow()
		ctx := context.Background()

		for {
			t, b, err := conn.Read(ctx)
			if err != nil {
				// ...
				slog.Error("Error reading message", "error", err)
				return
			}
			var msg map[string]interface{}
			json.Unmarshal(b, &msg)
			switch msg["type"] {
			case "markdown":
				rendered, metaData, err := renderMarkdown([]byte(msg["content"].(string)))
				if err != nil {
					slog.Error("Error rendering markdown", "error", err)
				}
				responses := make(map[string]interface{})
				responses["type"] = "markdown"
				responses["content"] = string(rendered)
				responses["id"] = msg["id"]
				responses["meta"] = metaData
				b, _ := json.Marshal(responses)
				conn.Write(ctx, t, b)
			case "load":
				fileName := msg["id"].(string)
				idxName := msg["idx"].(string)
				idx := index.GetIndex(idxName)
				if idx == nil {
					slog.Error("Index not found", "index", idxName)
					return
				}
				fileName = idx.GetEntryBySlug(fileName).FileName
				content, err := fs.ReadFile(idx.FilesDir, fileName)
				if err != nil {
					slog.Error("Error reading file", "error", err)
					return
				}
				responses := make(map[string]interface{})
				responses["type"] = "filecontent"
				responses["content"] = string(content)
				responses["id"] = fileName
				b, _ := json.Marshal(responses)
				conn.Write(ctx, t, b)
			case "save":
				fileName := msg["id"].(string)
				content := msg["content"].(string)
				_ = content
				idxName := msg["idx"].(string)
				idx := index.GetIndex(idxName)
				if idx == nil {
					slog.Error("Index not found", "index", idxName)
					return
				}
				stat, err := fs.Stat(idx.FilesDir, fileName)
				if err != nil {
					slog.Error("Error getting file info", "error", err)
					return
				}
				path := idx.DirPath + "/" + fileName
				err = os.WriteFile(path, []byte(content), stat.Mode())
				if err != nil {
					slog.Error("Error writing to file", "error", err)
					return
				}
				slog.Info("File saved", "file", path)
			default:
				slog.Debug("received unknown message", "message", string(b), "type", t)
			}
		}
	}))
}
