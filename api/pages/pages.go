package pages

import (
	"log/slog"
	"net/http"

	"github.com/sokkalf/hubro/index"
	"github.com/sokkalf/hubro/server"
)

func pageIndex(h *server.Hubro, entries []index.IndexEntry) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		slog.Info("Serving index")
		// Response for HTMX
		h.RenderWithoutLayout(w, r, "api/index", entries)
	}
}

func Register(prefix string, h *server.Hubro, mux *http.ServeMux, options interface{}) {
	indices := options.([]*index.Index)

	slog.Info("Registering API", "prefix", prefix)
	for _, i := range indices {
		endpoint := "/" + i.GetName() + "/index"
		mux.HandleFunc(endpoint, pageIndex(h, i.Entries))
		slog.Info("Registered endpoint", "endpoint", prefix + endpoint)
	}
}
