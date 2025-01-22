package admin

import (
	"context"
	"io/fs"
	"log/slog"
	"net/http"

	"github.com/coder/websocket"
	"github.com/sokkalf/hubro/index"
	"github.com/sokkalf/hubro/server"
)

func basicAuth(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		username, password, ok := r.BasicAuth()
		if !ok || username != "admin" || password != "admin" {
			w.Header().Set("WWW-Authenticate", `Basic realm="Restricted"`)
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		h(w, r)
	}
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

			slog.Debug("received message", "message", string(b), "type", t)
			conn.Write(ctx, t, []byte("received: "+string(b)))
		}
	}))
}
