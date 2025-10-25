package gzip

import (
	"compress/gzip"
	"net/http"
	"strings"

	"github.com/sokkalf/hubro/server"
)

type gzipResponseWriter struct {
	http.ResponseWriter
	Writer *gzip.Writer
}

func GzipMiddleware() server.Middleware {
	return func(h *server.Hubro) func(http.Handler) http.Handler {
		return func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
					next.ServeHTTP(w, r)
					return
				}
				gzWriter := gzip.NewWriter(w)
				defer gzWriter.Close()
				w.Header().Set("Content-Encoding", "gzip")
				gzResponseWriter := &gzipResponseWriter{ResponseWriter: w, Writer: gzWriter}
				next.ServeHTTP(gzResponseWriter, r)
			})
		}
	}
}

func (w *gzipResponseWriter) Write(b []byte) (int, error) {
	return w.Writer.Write(b)
}
