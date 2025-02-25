package logging

import (
	"log/slog"
	"os"
	"time"

	slogmulti "github.com/samber/slog-multi"
	"github.com/sokkalf/hubro/config"
	slogseq "github.com/sokkalf/slog-seq"
	"go.opentelemetry.io/otel/sdk/trace"
)

func InitSeqLog(logLevel slog.Level, seqEndpoint string, seqAPIKey string) (closeFunc func()) {
	opts := &slog.HandlerOptions{Level: logLevel, ReplaceAttr: replaceDuration, AddSource: true}
	_, handler := slogseq.NewLogger(seqEndpoint,
		slogseq.WithAPIKey(seqAPIKey),
		slogseq.WithFlushInterval(2 * time.Second),
		slogseq.WithBatchSize(100),
		slogseq.WithHandlerOptions(opts),
	)

	textLogger := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: logLevel})
	logger := slog.New(slogmulti.Fanout(handler, textLogger))
	slog.SetDefault(logger.With("appname", "hubro").
		With("appversion", config.Config.Version).
		With("environment", config.Config.Environment))

	spanProcessor := trace.NewSimpleSpanProcessor(&slogseq.LoggingSpanProcessor{Handler: handler})
	tp := trace.NewTracerProvider(trace.WithSpanProcessor(spanProcessor), trace.WithSampler(trace.AlwaysSample()))
	config.Config.Tracer = tp.Tracer("hubro")
	return func() {
		handler.Close()
	}
}
