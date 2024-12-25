package main

import (
	"embed"
	"io/fs"
	"log/slog"

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
	config := server.Config{
		RootPath:  "/",
		VendorDir: vendorDir,
	}
	h := server.NewHubro(config)
	h.Use(logging.LogMiddleware())
	h.Start()
}
