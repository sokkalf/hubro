package logging

import (
	"log/slog"
	"os"
)

func InitLogger(env string) {
	if env != "development" {
		slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo})))
	} else {
		InitTintLog(slog.LevelDebug)
	}
}
