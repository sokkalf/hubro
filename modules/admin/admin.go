package admin

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
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

func Register(prefix string, h *server.Hubro, mux *http.ServeMux, options interface{}) {
    slog.Info("Registering admin module")

    mux.Handle("/", basicAuth(adminIndexHandler(h)))
    mux.Handle("/edit", basicAuth(adminEditHandler(h)))
    mux.Handle("/ws", basicAuth(adminWebSocketHandler(h)))
}

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

func adminIndexHandler(h *server.Hubro) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        indices := index.GetIndices()
        h.RenderWithLayout(w, r, "admin/app", "admin/index", indices)
    }
}

func adminEditHandler(h *server.Hubro) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        slug := r.URL.Query().Get("p")
        idxName := r.URL.Query().Get("idx")

        i, err := getIndexByName(idxName)
        if err != nil {
			msg := err.Error()
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
            slog.Error(msg, "error", err)
            h.ErrorHandler(w, r, http.StatusInternalServerError, &msg)
            return
        }

        data := struct {
            Entry      *index.IndexEntry
            RawContent string
        }{
            Entry:      entry,
            RawContent: string(fileContent),
        }

        h.RenderWithLayout(w, r, "admin/app", "admin/edit", data)
    }
}

func adminWebSocketHandler(h *server.Hubro) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        conn, err := websocket.Accept(w, r, nil)
        if err != nil {
            slog.Error("Error accepting websocket connection", "error", err)
            return
        }
        defer conn.Close(websocket.StatusInternalError, "closing")

        ctx := context.Background()
        for {
            msgType, rawMsg, err := conn.Read(ctx)
            if err != nil {
                slog.Error("Error reading message", "error", err)
                return
            }

            var msg map[string]interface{}
            if err := json.Unmarshal(rawMsg, &msg); err != nil {
                slog.Error("Invalid JSON message", "error", err)
                return
            }

            switch msg["type"] {
            case "markdown":
                handleMarkdownMessage(ctx, conn, msgType, msg)

            case "load":
                handleLoadMessage(ctx, conn, msgType, msg)

            case "save":
                handleSaveMessage(ctx, conn, msg)

            default:
                slog.Debug("Received unknown message", "message", string(rawMsg), "type", msgType)
            }
        }
    }
}

func handleMarkdownMessage(ctx context.Context, conn *websocket.Conn, msgType websocket.MessageType, msg map[string]interface{}) {
    content, _ := msg["content"].(string)
    rendered, metaData, err := renderMarkdown([]byte(content))
    if err != nil {
        slog.Error("Error rendering markdown", "error", err)
        return
    }
    responses := map[string]interface{}{
        "type":    "markdown",
        "content": string(rendered),
        "id":      msg["id"],
        "meta":    metaData,
    }
    _ = writeJSON(ctx, conn, msgType, responses)
}

func handleLoadMessage(ctx context.Context, conn *websocket.Conn, msgType websocket.MessageType, msg map[string]interface{}) {
    fileSlug, _ := msg["id"].(string)
    idxName, _ := msg["idx"].(string)

    idx, err := getIndexByName(idxName)
    if err != nil {
        slog.Error(err.Error())
        return
    }

    entry := idx.GetEntryBySlug(fileSlug)
    if entry == nil {
        slog.Error("Entry not found", "slug", fileSlug)
        return
    }

    content, err := fs.ReadFile(idx.FilesDir, entry.FileName)
    if err != nil {
        slog.Error("Error reading file", "error", err)
        return
    }

    responses := map[string]interface{}{
        "type":    "filecontent",
        "content": string(content),
        "id":      entry.FileName,
    }
    _ = writeJSON(ctx, conn, msgType, responses)
}

func handleSaveMessage(ctx context.Context, conn *websocket.Conn, msg map[string]interface{}) {
    fileName, _ := msg["id"].(string)
    content, _ := msg["content"].(string)
    idxName, _ := msg["idx"].(string)

    idx, err := getIndexByName(idxName)
    if err != nil {
        slog.Error(err.Error())
        return
    }

    stat, err := fs.Stat(idx.FilesDir, fileName)
    if err != nil {
        slog.Error("Error getting file info", "error", err)
        return
    }

    path := idx.DirPath + "/" + fileName
    if err := os.WriteFile(path, []byte(content), stat.Mode()); err != nil {
        slog.Error("Error writing to file", "error", err)
        return
    }

    slog.Info("File saved", "file", path)
}

func getIndexByName(name string) (*index.Index, error) {
    if name == "" {
        return nil, fmt.Errorf("Index name not provided")
    }
    idx := index.GetIndex(name)
    if idx == nil {
        return nil, fmt.Errorf("Index not found: %s", name)
    }
    return idx, nil
}

func writeJSON(ctx context.Context, conn *websocket.Conn, msgType websocket.MessageType, v interface{}) error {
    data, err := json.Marshal(v)
    if err != nil {
        slog.Error("Error marshalling JSON", "error", err)
        return err
    }
    if wErr := conn.Write(ctx, msgType, data); wErr != nil {
        slog.Error("Error writing WebSocket message", "error", wErr)
        return wErr
    }
    return nil
}
