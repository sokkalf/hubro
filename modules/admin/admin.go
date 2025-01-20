package admin

import (
	"log/slog"
	"net/http"

	"github.com/coder/websocket"
	"github.com/sokkalf/hubro/server"
)

func Register(prefix string, h *server.Hubro, mux *http.ServeMux, options interface{}) {
	slog.Info("Registering admin module")
	mux.Handle("/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Hello, World!"))
	}))
	mux.Handle("/ws", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, err := websocket.Accept(w, r, nil)
		if err != nil {
			slog.Error("Error accepting websocket connection", "error", err)
			return
		}
	}))
}
