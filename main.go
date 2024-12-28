package main

import (
	"encoding/json"
	"log/slog"
	"os"
	"time"

	"github.com/sokkalf/hubro/logging"
	"github.com/sokkalf/hubro/modules/feeds"
	"github.com/sokkalf/hubro/modules/page"
	"github.com/sokkalf/hubro/modules/redirects"
	"github.com/sokkalf/hubro/server"
)

func main() {
	start := time.Now()
	logging.InitLogger("development")
	slog.Info("Starting Hubro 🦉")
	vendorDir := os.DirFS("view/assets/vendor")
	layoutDir := os.DirFS("view/layouts")
	templateDir := os.DirFS("view/templates")
	publicDir := os.DirFS("view/public")
	pagesDir := os.DirFS("pages")
	blogDir := os.DirFS("blog")

	config := server.Config{
		RootPath:    "/",
		VendorDir:   vendorDir,
		LayoutDir:   layoutDir,
		TemplateDir: templateDir,
		PublicDir:   publicDir,
	}
	h := server.NewHubro(config)
	h.Use(logging.LogMiddleware())
	blogIndex := server.NewIndex("blog", h.RootPath+"blog")
	pageIndex := server.NewIndex("pages", h.RootPath+"page")
	h.AddModule("/page", page.Register, page.PageOptions{FilesDir: pagesDir, IndexSummary: false, IndexFunc: pageIndex.AddEntry})
	h.AddModule("/blog", page.Register, page.PageOptions{FilesDir: blogDir, IndexSummary: true, IndexFunc: blogIndex.AddEntry})
	pageIndex.SortBySortOrder()
	blogIndex.SortByDate()
	h.AddModule("/feeds", feeds.Register, blogIndex)
	b, err := os.ReadFile("legacyRoutes.json")
	if err != nil {
		slog.Info("No legacy routes found")
	} else {
		var legacyRoutes []redirects.PathRoutes
		err = json.Unmarshal(b, &legacyRoutes)
		if err != nil {
			slog.Error("Error unmarshalling legacy routes", "error", err)
		} else {
			for _, route := range legacyRoutes {
				h.AddModule(route.Path, redirects.Register, route)
			}
		}
	}
	h.Start(start)
}
