package main

import (
	"io/fs"
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
	blogDir := os.DirFS("blog")

	config := server.Config{
		RootPath:  "/",
		VendorDir: vendorDir,
		LayoutDir: layoutDir,
		TemplateDir: templateDir,
		PublicDir: publicDir,
	}
	h := server.NewHubro(config)
	h.Use(logging.LogMiddleware())
	pageIndex := server.NewIndex("pages", h.RootPath + "page")
	h.AddModule("/page", page.Register,
		struct{
			FilesDir fs.FS
			IndexFunc func(server.IndexEntry)
		}{FilesDir: pagesDir, IndexFunc: pageIndex.AddEntry})
	blogIndex := server.NewIndex("blog", h.RootPath + "blog")
	h.AddModule("/blog", page.Register,
		struct{
			FilesDir fs.FS
			IndexFunc func(server.IndexEntry)
		}{FilesDir: blogDir, IndexFunc: blogIndex.AddEntry})
	h.Start()
}
