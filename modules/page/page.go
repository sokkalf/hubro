package page

import (
	"fmt"
	"net/http"

	"github.com/sokkalf/hubro/server"
)

func Register(h *server.Hubro, mux *http.ServeMux) () {
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Hello, you've requested: %s\n", r.URL.Path)
	})
}
