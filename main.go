package main

import (
    "fmt"
    "html/template"
    "log"
    "net/http"
    "path/filepath"
)

var tmpl *template.Template

func main() {
    // Parse all templates in the templates directory
    var err error
    tmpl, err = template.ParseGlob(filepath.Join("templates", "*.gohtml"))
    if err != nil {
        log.Fatalf("Error parsing templates: %v", err)
    }
	for _, t := range tmpl.Templates() {
		fmt.Printf("Parsed template: %s\n", t.Name())
	}

    // Serve static files from the "static" directory at "/static/" path
    fs := http.FileServer(http.Dir("./static"))
    http.Handle("/static/", http.StripPrefix("/static/", fs))

    // Routes
    http.HandleFunc("/", indexHandler)
    http.HandleFunc("/ping", pingHandler)

    // Start server
    fmt.Println("Listening on http://localhost:8080")
    log.Fatal(http.ListenAndServe(":8080", nil))
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
    // Render the "index.gohtml" template
    err := tmpl.ExecuteTemplate(w, "index.gohtml", nil)
    if err != nil {
        http.Error(w, "Failed to render template", http.StatusInternalServerError)
    }
}

// pingHandler is a simple route that returns "Pong!" text.
// We set an HTMX response header ("HX-Trigger") to demonstrate sending events back to the client.
func pingHandler(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("HX-Trigger", "pongReceived")
    fmt.Fprintln(w, "Pong!")
}
