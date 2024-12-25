package main

import (
	"log/slog"
	"os"

	"github.com/sokkalf/hubro/logging"
	"github.com/sokkalf/hubro/server"
)

func main() {
	logging.InitLogger("development")
	slog.Info("Starting Hubro 🦉")
	vendorDir := os.DirFS("view/assets/vendor")
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
