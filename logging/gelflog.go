package logging

import (
	"os"
	"log/slog"

	"github.com/Graylog2/go-gelf/gelf"
	sloggraylog "github.com/samber/slog-graylog/v2"
	slogmulti "github.com/samber/slog-multi"
)


func InitGelfLog(logLevel slog.Level, gelfEndpoint string) {
	gelfWriter, err := gelf.NewWriter(gelfEndpoint)
	if err != nil {
		slog.Error("Error creating GELF writer", "Error", err)
	}
	gelfWriter.CompressionType = gelf.CompressGzip
	seqLogger := sloggraylog.Option{Level: logLevel, Writer: gelfWriter}.NewGraylogHandler()
	textLogger := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: logLevel})
	logger := slog.New(slogmulti.Fanout(seqLogger, textLogger))
	slog.SetDefault(logger.With("appname", "hubro"))
}
