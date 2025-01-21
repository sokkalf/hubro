package admin

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/coder/websocket"
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
	mux.Handle("/", basicAuth(func(w http.ResponseWriter, r *http.Request) {
		h.RenderWithLayout(w, r, "admin/app", "admin/index", nil)
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
			conn.Write(ctx, t, []byte("Hello, World!"))
		}
	}))
}
