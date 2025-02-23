package logging

import (
	"log/slog"
	"os"

	"github.com/sokkalf/hubro/config"
)

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
