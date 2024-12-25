package main

import (
	"embed"
	"io/fs"
	"log/slog"

	"github.com/sokkalf/hubro/logging"
	"github.com/sokkalf/hubro/server"
)

//go:embed web/assets/vendor
var vendorDir embed.FS

func main() {
	logging.InitLogger("development")
	slog.Info("Starting Hubro")
	vendorDir, err := fs.Sub(vendorDir, "web/assets/vendor")
	if err != nil {
		panic(err)
	}
	config := server.Config{
		RootPath:  "/",
		VendorDir: vendorDir,
	}
	h := server.NewHubro(config)
	h.Start()
}
