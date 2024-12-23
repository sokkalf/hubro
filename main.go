package main

import (
	"embed"
	"io/fs"

	"github.com/sokkalf/hubro/server"
)

//go:embed vendor
var vendorDir embed.FS

func main() {
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
