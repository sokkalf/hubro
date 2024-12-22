package main

import (
	"fmt"
	"html/template"
	"net/http"

	"github.com/sokkalf/hubro/server"
)

var tmpl *template.Template

func main() {
	h := server.NewHubro()

	// Start server
	fmt.Println("Listening on http://localhost:8080")
	http.ListenAndServe(":8080", h.Mux)
}
