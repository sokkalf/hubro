package main

import (
	"encoding/json"
	"log/slog"
	"os"
	"time"

	"github.com/sokkalf/hubro/config"
	"github.com/sokkalf/hubro/logging"
	"github.com/sokkalf/hubro/modules/feeds"
	"github.com/sokkalf/hubro/modules/page"
	"github.com/sokkalf/hubro/modules/redirects"
	"github.com/sokkalf/hubro/server"
	"github.com/sokkalf/hubro/index"
)

// Overwritten by the build system
var Version = "0.0.1-dev"

func main() {
	start := time.Now()
	config.Init()
	config.Config.Version = Version
	logging.InitLogger("development")
	slog.Info("Starting Hubro 🦉")
	vendorDir := os.DirFS("view/assets/vendor")
	layoutDir := os.DirFS("view/layouts")
	templateDir := os.DirFS("view/templates")
	publicDir := os.DirFS("view/public")
	pagesDir := os.DirFS(config.Config.PagesDir)
	blogDir := os.DirFS(config.Config.BlogDir)

	cfg := server.Config{
		RootPath:    config.Config.RootPath,
		Port:        config.Config.Port,
		VendorDir:   vendorDir,
		LayoutDir:   layoutDir,
		TemplateDir: templateDir,
		PublicDir:   publicDir,
	}
	h := server.NewHubro(cfg)
	h.Use(logging.LogMiddleware())
	blogIndex := index.NewIndex("blog", config.Config.RootPath+"blog")
	pageIndex := index.NewIndex("pages", config.Config.RootPath+"page")
	h.AddModule("/page", page.Register, page.PageOptions{FilesDir: pagesDir, IndexSummary: false, IndexFunc: pageIndex.AddEntry})
	h.AddModule("/blog", page.Register, page.PageOptions{FilesDir: blogDir, IndexSummary: true, IndexFunc: blogIndex.AddEntry})
	pageIndex.SortBySortOrder()
	blogIndex.SortByDate()
	if config.Config.FeedsEnabled {
		if blogIndex.Count() > 0 {
			h.AddModule("/feeds", feeds.Register, blogIndex)
		} else {
			config.Config.FeedsEnabled = false
			slog.Info("No blog entries found, skipping feeds")
		}
	}
	b, err := os.ReadFile(config.Config.LegacyRoutesFile)
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
