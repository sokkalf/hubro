package logging

import (
	"os"
	"time"
	"log/slog"

	"github.com/lmittmann/tint"
)

func InitTintLog(logLevel slog.Level) (closeFunc func()) {
	handler := tint.NewHandler(os.Stdout, &tint.Options{
		Level:      logLevel,
		TimeFormat: time.StampMilli,
	})
	slog.SetDefault(slog.New(handler))

	return func() {
		// No cleanup needed
	}
}
