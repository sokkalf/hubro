package api

import (
	"log/slog"
	"net/http"

	"github.com/sokkalf/hubro/config"
)

func RegisterOptionsHandler(prefix string, mux *http.ServeMux) {
	slog.Debug("Registering OPTIONS handler", "prefix", prefix)
	mux.HandleFunc("OPTIONS " + prefix, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", config.Config.BaseURL)
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, hx-current-url, hx-request")
		w.WriteHeader(http.StatusOK)
	})
}
