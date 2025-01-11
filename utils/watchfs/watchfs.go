package watchfs

import (
	"io/fs"
	"log/slog"
	"os"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/sokkalf/hubro/index"
)

const debounceDuration = 500 * time.Millisecond

func WatchFS(dir string, idx *index.Index) (*fs.FS, error) {
	fsys := os.DirFS(dir)
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}

	if err := watcher.Add(dir); err != nil {
		return nil, err
	}
	// Subdirectories
	err = fs.WalkDir(fsys, ".", func(path string, d fs.DirEntry, err error) error {
		if d.IsDir() && path != "." {
			if err := watcher.Add(dir + "/" + path); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		slog.Warn("Error walking directory", "error", err)
	}

	trigger := make(chan struct{})

	go func() {
		var timer *time.Timer

		for {
			select {
			case event := <-watcher.Events:
				if event.Op&(fsnotify.Write|fsnotify.Create|fsnotify.Remove) != 0 {
					fileInfo, err := os.Stat(event.Name)
					if err != nil && !event.Has(fsnotify.Remove) {
						slog.Error("Error getting file info", "error", err)
						continue
					} else if err != nil && event.Has(fsnotify.Remove) {
						// If the file was removed, we don't need to do anything
					}
					if err == nil && fileInfo.IsDir() {
						slog.Info("Watching new directory", "directory", event.Name)
						watcher.Add(event.Name)
					}

					// Reset the timer every time we get a relevant event
					if timer != nil {
						timer.Stop()
					}
					timer = time.AfterFunc(debounceDuration, func() {
						trigger <- struct{}{}
					})
				}
			case err := <-watcher.Errors:
				if err != nil {
					slog.Error("Error watching directory", "error", err)
				}
			}
		}
	}()

	go func() {
		for range trigger {
			slog.Info("Starting directory scan")
			idx.MsgBroker.Publish(index.Scanned)
		}
	}()

	return &fsys, nil
}
