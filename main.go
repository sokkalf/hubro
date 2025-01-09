package main

import (
	"encoding/json"
	"io/fs"
	"log/slog"
	"os"
	"sync"
	"time"

	pagesAPI "github.com/sokkalf/hubro/api/pages"
	"github.com/sokkalf/hubro/config"
	"github.com/sokkalf/hubro/helpers"
	"github.com/sokkalf/hubro/index"
	"github.com/sokkalf/hubro/logging"
	"github.com/sokkalf/hubro/modules/feeds"
	"github.com/sokkalf/hubro/modules/healthcheck"
	"github.com/sokkalf/hubro/modules/page"
	"github.com/sokkalf/hubro/modules/redirects"
	userstatic "github.com/sokkalf/hubro/modules/user_static"
	"github.com/sokkalf/hubro/server"
	"github.com/sokkalf/hubro/utils/watchfs"
)

// Overwritten by the build system
var Version = "v0.0.1-dev"

func main() {
	start := time.Now()
	config.Init()
	config.Config.Version = Version
	logging.InitLogger()
	slog.Info("Starting Hubro 🦉")
	vendorDir := os.DirFS("view/assets/vendor")
	layoutDir := os.DirFS("view/layouts")
	templateDir := os.DirFS("view/templates")
	publicDir := os.DirFS("view/public")


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
	h.AddModule("/healthz", healthcheck.Register, nil)
	var userStaticDir fs.FS
	usd, err := os.Stat(config.Config.UserStaticDir)
	if err != nil {
		slog.Info("No userfiles directory found")
	} else if usd.IsDir() {
		userStaticDir = os.DirFS(config.Config.UserStaticDir)
		h.AddModule("/userfiles", userstatic.Register, userStaticDir)
	} else {
		slog.Error("User static directory is not a directory")
	}
	pageIndex := index.NewIndex("pages", config.Config.RootPath+"page")
	pageIndex.SetSortMode(index.SortBySortOrder)
	blogIndex := index.NewIndex("blog", config.Config.RootPath+"blog")
	blogIndex.SetSortMode(index.SortByDate)
	pagesDir, err := watchfs.WatchFS(config.Config.PagesDir, pageIndex)
	if err != nil {
		slog.Error("Error watching pages directory", "error", err)
	}
	blogDir, err := watchfs.WatchFS(config.Config.BlogDir, blogIndex)
	if err != nil {
		slog.Error("Error watching blog directory", "error", err)
	}
	helpers.TagCloudInit(pageIndex)
	helpers.TagCloudInit(blogIndex)
	wg := sync.WaitGroup{}
	wg.Add(2)
	go func () {
		defer wg.Done()
		h.AddModule("/page", page.Register, page.PageOptions{FilesDir: *pagesDir, Index: pageIndex})
	}()
	go func () {
		defer wg.Done()
		h.AddModule("/blog", page.Register, page.PageOptions{FilesDir: *blogDir, Index: blogIndex})
	}()
	wg.Wait()
	h.AddModule("/api/pages", pagesAPI.Register, []*index.Index{pageIndex, blogIndex})

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
