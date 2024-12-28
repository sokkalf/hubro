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
	blogIndex := server.NewIndex("blog", h.RootPath + "blog")
	pageIndex := server.NewIndex("pages", h.RootPath + "page")
	h.AddModule("/page", page.Register, page.PageOptions{FilesDir: pagesDir, IndexSummary: false, IndexFunc: pageIndex.AddEntry})
	h.AddModule("/blog", page.Register, page.PageOptions{FilesDir: blogDir, IndexSummary: true, IndexFunc: blogIndex.AddEntry})
	pageIndex.SortBySortOrder()
	h.Start()
}
