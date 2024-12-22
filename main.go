package main

import (
	"html/template"

	"github.com/sokkalf/hubro/server"
)

var tmpl *template.Template

func main() {
	h := server.NewHubro()
	h.Start()
}
