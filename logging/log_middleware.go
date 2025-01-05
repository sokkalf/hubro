package logging

import (
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/sokkalf/hubro/server"
)

// Borrowed from https://stackoverflow.com/a/78381482
type CustomResponseWriter struct {
	responseWriter http.ResponseWriter
	StatusCode     int
}

func ExtendResponseWriter(w http.ResponseWriter) *CustomResponseWriter {
	return &CustomResponseWriter{w, 0}
}

func (w *CustomResponseWriter) Write(b []byte) (int, error) {
	return w.responseWriter.Write(b)
}

func (w *CustomResponseWriter) Header() http.Header {
	return w.responseWriter.Header()
}

func (w *CustomResponseWriter) WriteHeader(statusCode int) {
	// receive status code from this method
	w.StatusCode = statusCode
	w.responseWriter.WriteHeader(statusCode)

	return
}

func (w *CustomResponseWriter) Done() {
	// if the `w.WriteHeader` wasn't called, set status code to 200 OK
	if w.StatusCode == 0 {
		w.StatusCode = http.StatusOK
	}

	return
}

func getRemoteAddr(r *http.Request) (string, bool) {
	xff := r.Header.Get("X-Forwarded-For")
	if xff != "" {
		parts := strings.Split(xff, ",")
		return strings.TrimSpace(parts[0]), true
	}
	return r.RemoteAddr, false
}

func LogMiddleware() server.Middleware {
	return func(h *server.Hubro) func(http.Handler) http.Handler {
		return func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				start := time.Now()
				ew := ExtendResponseWriter(w)
				next.ServeHTTP(ew, r)
				ew.Done()
				if r.URL.Path == "/healthz" || r.URL.Path == "/healthz/" {
					// Don't log health checks
					return
				}
				userAgent := r.Header.Get("User-Agent")
				hxBoosted := r.Header.Get("HX-Boosted")
				if hxBoosted == "" {
					hxBoosted = "false"
				}
				remoteAddr, proxied := getRemoteAddr(r)
				query := r.URL.Query().Encode()
				if query != "" {
					slog.Info(fmt.Sprintf("%s %s?%s", r.Method, r.URL.Path, query),
						"remoteaddr", remoteAddr, "proxied", proxied, "status", ew.StatusCode,
						"user-agent", userAgent, "hx-boosted", hxBoosted, "duration", time.Since(start))
				} else {
					slog.Info(fmt.Sprintf("%s %s", r.Method, r.URL.Path),
						"remoteaddr", remoteAddr, "proxied", proxied, "status", ew.StatusCode,
						"user-agent", userAgent, "hx-boosted", hxBoosted, "duration", time.Since(start))
				}
			})
		}
	}
}
