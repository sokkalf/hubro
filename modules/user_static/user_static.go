package userstatic

import (
	"io/fs"
	"log/slog"
	"net/http"

	"github.com/sokkalf/hubro/server"
)

func Register(prefix string, h *server.Hubro, mux *http.ServeMux, options any) {
	userStaticDir := options.(fs.FS)

	slog.Info("Registering User Static", "prefix", prefix)
	mux.Handle("GET /", h.FileServerWithDirectoryListingDisabled(http.FileServer(http.FS(userStaticDir))))
}
