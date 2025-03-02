package logging

import (
	"log/slog"
	"os"
	"time"

	"github.com/sokkalf/hubro/config"
)

func replaceDuration(groups []string, attr slog.Attr) slog.Attr {
	if attr.Key == "duration" {
		if d, ok := attr.Value.Any().(time.Duration); ok {
			ms := float64(d.Microseconds()) / 1000.0
			attr.Value = slog.Float64Value(ms)
		}
	}
	return attr
}

func InitLogger() func() {
	env := config.Config.Environment
	if env != "development" {
		if config.Config.GelfEndpoint != nil {
			return InitGelfLog(slog.LevelInfo, *config.Config.GelfEndpoint)
		} else if config.Config.SeqEndpoint != nil {
			return InitSeqLog(slog.LevelInfo, *config.Config.SeqEndpoint, *config.Config.SeqAPIKey)
		} else {
			slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo})))
		}
	} else {
		return InitTintLog(slog.LevelDebug)
	}
	return func() {
		// No cleanup needed
	}
}
