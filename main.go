package main

import (
	"embed"
	"io/fs"
	"log/slog"
	"os"

	"github.com/sokkalf/hubro/server"
)

//go:embed vendor
var vendorDir embed.FS

func initLogger() {
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug})))
}

func main() {
	initLogger()
	slog.Info("Starting Hubro")
	vendorDir, err := fs.Sub(vendorDir, "vendor")
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
