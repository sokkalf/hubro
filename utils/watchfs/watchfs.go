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

	trigger := make(chan struct{})

	go func() {
		var timer *time.Timer

		for {
			select {
			case event := <-watcher.Events:
				if event.Op&(fsnotify.Write|fsnotify.Create|fsnotify.Remove) != 0 {
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
