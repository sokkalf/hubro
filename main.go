package main

import (
	"log/slog"
	"os"

	"github.com/sokkalf/hubro/logging"
	"github.com/sokkalf/hubro/modules/page"
	"github.com/sokkalf/hubro/server"
)

func main() {
	logging.InitLogger("development")
	slog.Info("Starting Hubro 🦉")
	vendorDir := os.DirFS("view/assets/vendor")
	layoutDir := os.DirFS("view/layouts")
	templateDir := os.DirFS("view/templates")
	publicDir := os.DirFS("view/public")
	pagesDir := os.DirFS("pages")

	config := server.Config{
		RootPath:  "/",
		VendorDir: vendorDir,
		LayoutDir: layoutDir,
		TemplateDir: templateDir,
		PublicDir: publicDir,
		PagesDir: pagesDir,
	}
	h := server.NewHubro(config)
	h.Use(logging.LogMiddleware())
	h.AddModule("/page", page.Register)
	h.Start()
}
