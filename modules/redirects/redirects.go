package redirects

import (
	"log/slog"
	"net/http"

	"github.com/sokkalf/hubro/server"
)

type Route struct {
    OldPath string `json:"oldPath"`
    NewPath string `json:"newPath"`
}

type PathRoutes struct {
    Path     string  `json:"path"`
    Routes   []Route `json:"routes"`
}

func Register(prefix string, h *server.Hubro, mux *http.ServeMux, options interface{}) {
	routes := options.(PathRoutes)
	for _, route := range routes.Routes {
		slog.Info("Registering redirect", "oldPath", prefix + route.OldPath, "newPath", route.NewPath)
		mux.HandleFunc(route.OldPath, func(w http.ResponseWriter, r *http.Request) {
			http.Redirect(w, r, route.NewPath, http.StatusMovedPermanently)
		})
	}
}
