package watchfs

import (
	"io/fs"
	"log/slog"
	"os"

	"github.com/fsnotify/fsnotify"
	"github.com/sokkalf/hubro/index"
)

func WatchFS(dir string, idx *index.Index) (*fs.FS, error) {
	fs := os.DirFS(dir)
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}
	watcher.Add(dir)
	go func() {
		for {
			select {
			case event := <-watcher.Events:
				if event.Op&fsnotify.Write == fsnotify.Write || event.Op&fsnotify.Create == fsnotify.Create || event.Op&fsnotify.Remove == fsnotify.Remove {
					slog.Info("File modified", "file", event.Name)
					idx.MsgBroker.Publish(index.Scanned)
				}
			case err := <-watcher.Errors:
				if err != nil {
					slog.Error("Error watching directory", "error", err)
				}
			}
		}
	}()

	return &fs, nil
}
