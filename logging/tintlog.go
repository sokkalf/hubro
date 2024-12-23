package logging

import (
	"os"
	"time"
	"log/slog"

	"github.com/lmittmann/tint"
)

func InitTintLog(logLevel slog.Level) {
	slog.SetDefault(slog.New(
		tint.NewHandler(os.Stdout, &tint.Options{
			Level:      logLevel,
			TimeFormat: time.StampMilli,
		}),
	))
}
