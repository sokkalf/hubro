package main

import (
	"embed"
	"io/fs"
	"log/slog"
	"os"

	"github.com/sokkalf/hubro/logging"
	"github.com/sokkalf/hubro/server"
)

//go:embed view/assets/vendor
var vendorDir embed.FS

func main() {
	logging.InitLogger("development")
	slog.Info("Starting Hubro 🦉")
	vendorDir, err := fs.Sub(vendorDir, "view/assets/vendor")
	if err != nil {
		panic(err)
	}
	layoutDir := os.DirFS("view/layouts")
	templateDir := os.DirFS("view/templates")

	config := server.Config{
		RootPath:  "/",
		VendorDir: vendorDir,
		LayoutDir: layoutDir,
		TemplateDir: templateDir,
	}
	h := server.NewHubro(config)
	h.Use(logging.LogMiddleware())
	h.Start()
}
