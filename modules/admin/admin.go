package admin

import (
	"log/slog"
	"net/http"

	"github.com/sokkalf/hubro/server"
)

func Register(prefix string, h *server.Hubro, mux *http.ServeMux, options interface{}) {
	slog.Info("Registering admin module")
	mux.Handle("/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Admin page"))
	}))
}
