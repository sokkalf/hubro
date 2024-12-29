package logging

import (
	"fmt"
	"log/slog"
	"net/http"
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

func LogMiddleware() server.Middleware {
	return func(h *server.Hubro) func(http.Handler) http.Handler {
		return func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				start := time.Now()
				ew := ExtendResponseWriter(w)
				next.ServeHTTP(ew, r)
				ew.Done()
				query := r.URL.Query().Encode()
				if query != "" {
					slog.Info(fmt.Sprintf("%s %s?%s", r.Method, r.URL.Path, query), "remoteaddr", r.RemoteAddr, "status", ew.StatusCode, "duration", time.Since(start))
				} else {
					slog.Info(fmt.Sprintf("%s %s", r.Method, r.URL.Path), "remoteaddr", r.RemoteAddr, "status", ew.StatusCode, "duration", time.Since(start))
				}
			})
		}
	}
}
