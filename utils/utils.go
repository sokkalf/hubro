package utils

import (
	"net/http"
	"strings"
)

func FileServerWithDirectoryListingDisabled(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/") || r.URL.Path == "" {
			http.Error(w, "403 directory listing not allowed", http.StatusForbidden)
			return
		}
		h.ServeHTTP(w, r)
	})
}
