package pages

import (
	"log/slog"
	"net/http"

	"github.com/sokkalf/hubro/api"
	"github.com/sokkalf/hubro/index"
	"github.com/sokkalf/hubro/server"
)

func pageIndex(h *server.Hubro, entries []index.IndexEntry) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		slog.Info("Serving index")
		w.Header().Set("Content-Type", "text/html")
		// Response for HTMX
		h.RenderWithoutLayout(w, r, "api/index", entries)
	}
}

func Register(prefix string, h *server.Hubro, mux *http.ServeMux, options any) {
	indices := options.([]*index.Index)

	slog.Info("Registering API", "prefix", prefix)
	for i := range indices {
		indices[i].RLock()
		endpoint := "/" + indices[i].GetName() + "/index"
		mux.HandleFunc("GET "+endpoint, pageIndex(h, indices[i].GetEntries()))
		api.RegisterOptionsHandler(endpoint, mux)
		slog.Info("Registered endpoint", "endpoint", prefix+endpoint)
		indices[i].RUnlock()
	}
}
