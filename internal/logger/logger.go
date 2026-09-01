package logger

import (
	"log/slog"
	"os"
	"strings"
)

var Log = slog.New(
	slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}),
)

func Init(env string) {
	opts := &slog.HandlerOptions{Level: slog.LevelInfo}

	var handler slog.Handler
	if strings.EqualFold(env, "production") {
		handler = slog.NewJSONHandler(os.Stdout, opts)
	} else {
		handler = slog.NewTextHandler(os.Stdout, opts)
	}

	Log = slog.New(handler)
}
