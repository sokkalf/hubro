package logging

import (
	"log/slog"
	"os"

	"github.com/sokkalf/hubro/config"
)

func InitLogger() {
	env := config.Config.Environment
	if env != "development" {
		if config.Config.GelfEndpoint != nil {
			InitGelfLog(slog.LevelInfo, *config.Config.GelfEndpoint)
		} else {
			slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo})))
		}
	} else {
		InitTintLog(slog.LevelDebug)
	}
}
