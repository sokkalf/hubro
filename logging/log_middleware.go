package logging

import (
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/sokkalf/hubro/server"
)

func LogMiddleware() server.Middleware {
	return func(h *server.Hubro) func(http.Handler) http.Handler {
		return func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				start := time.Now()
				next.ServeHTTP(w, r)
				slog.Info(fmt.Sprintf("%s %s", r.Method, r.URL.Path), "remoteaddr", r.RemoteAddr, "time", time.Since(start))
			})
		}
	}
}
