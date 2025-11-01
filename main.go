package main

import (
	"context"
	"encoding/json"
	"io/fs"
	"log/slog"
	"os"
	"sync"
	"time"

	pagesAPI "github.com/sokkalf/hubro/api/pages"
	"github.com/sokkalf/hubro/config"
	"github.com/sokkalf/hubro/gzip"
	"github.com/sokkalf/hubro/helpers"
	"github.com/sokkalf/hubro/index"
	"github.com/sokkalf/hubro/logging"
	"github.com/sokkalf/hubro/modules/admin"
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
	closeFunc := logging.InitLogger()
	defer closeFunc()
	ctx := context.Background()
	tr := config.Config.Tracer
	spanCtx, span := tr.Start(ctx, "main")
	slog.InfoContext(spanCtx, "Starting Hubro 🦉")
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
	span.AddEvent("Initializing middleware")
	h.Use(logging.LogMiddleware())
	h.Use(gzip.GzipMiddleware())
	span.End()
	spanCtx, span = tr.Start(spanCtx, "module registration")
	span.AddEvent("Healthcheck module")
	h.AddModule("/healthz", healthcheck.Register, nil)
	span.End()
	spanCtx, span = tr.Start(spanCtx, "Adding pages and blog entries")
	var userStaticDir fs.FS
	usd, err := os.Stat(config.Config.UserStaticDir)
	if err != nil {
		slog.InfoContext(spanCtx, "No userfiles directory found")
	} else if usd.IsDir() {
		userStaticDir = os.DirFS(config.Config.UserStaticDir)
		h.AddModule("/userfiles", userstatic.Register, userStaticDir)
	} else {
		slog.ErrorContext(spanCtx, "User static directory is not a directory")
	}
	span.AddEvent("Creating indices for pages and blog entries")
	pageIndex := index.NewIndex("pages", config.Config.RootPath+"page")
	pageIndex.SetSortMode(index.SortBySortOrder)
	blogIndex := index.NewIndex("blog", config.Config.RootPath+"blog")
	blogIndex.SetSortMode(index.SortByDate)
	pagesDir, err := watchfs.WatchFS(config.Config.PagesDir, pageIndex)
	if err != nil {
		slog.ErrorContext(spanCtx, "Error watching pages directory", "error", err)
	}
	blogDir, err := watchfs.WatchFS(config.Config.BlogDir, blogIndex)
	if err != nil {
		slog.ErrorContext(spanCtx, "Error watching blog directory", "error", err)
	}
	pageIndex.FilesDir = *pagesDir
	blogIndex.FilesDir = *blogDir
	pageIndex.DirPath = config.Config.PagesDir
	blogIndex.DirPath = config.Config.BlogDir
	helpers.TagCloudInit(pageIndex)
	helpers.TagCloudInit(blogIndex)
	span.AddEvent("Searching for pages and blog entries")
	wg := sync.WaitGroup{}
	wg.Add(2)
	go func() {
		defer wg.Done()
		h.AddModule("/page", page.Register, page.PageOptions{Index: pageIndex, Ctx: spanCtx})
	}()
	go func() {
		defer wg.Done()
		h.AddModule("/blog", page.Register, page.PageOptions{Index: blogIndex, Ctx: spanCtx})
	}()
	wg.Wait()
	span.End()
	spanCtx, span = tr.Start(spanCtx, "Adding API endpoints, feeds and legacy routes")
	span.AddEvent("Adding API endpoints")
	h.AddModule("/api/pages", pagesAPI.Register, []*index.Index{pageIndex, blogIndex})
	if config.Config.AdminEnabled {
		h.AddModule("/admin", admin.Register, nil)
	}
	span.AddEvent("Adding feeds")
	if config.Config.FeedsEnabled {
		if blogIndex.Count() > 0 {
			h.AddModule("/feeds", feeds.Register, blogIndex)
		} else {
			config.Config.FeedsEnabled = false
			slog.InfoContext(spanCtx, "No blog entries found, skipping feeds")
		}
	}
	span.AddEvent("Adding legacy routes")
	b, err := os.ReadFile(config.Config.LegacyRoutesFile)
	if err != nil {
		slog.Info("No legacy routes found")
	} else {
		var legacyRoutes []redirects.PathRoutes
		err = json.Unmarshal(b, &legacyRoutes)
		if err != nil {
			slog.ErrorContext(spanCtx, "Error unmarshalling legacy routes", "error", err)
		} else {
			for _, route := range legacyRoutes {
				h.AddModule(route.Path, redirects.Register, route)
			}
		}
	}
	span.End()
	err = h.Start(start)
	if err != nil {
		slog.Error("Error starting Hubro", "error", err)
	}
}
