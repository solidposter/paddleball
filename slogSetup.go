package main

import (
	"log/slog"
	"os"
)

func slogSetup(json bool, tag string) {
	var logger *slog.Logger

	if json {
		opts := &slog.HandlerOptions{
			ReplaceAttr: renameTimeAttr,
		}
		logger = slog.New(slog.NewJSONHandler(os.Stdout, opts))
	} else {
		logger = slog.New(slog.NewTextHandler(os.Stdout, nil))
	}

	if tag != "" {
		logger = logger.With("Tag", tag)
	}

	slog.SetDefault(logger)
}

func renameTimeAttr(groups []string, a slog.Attr) slog.Attr {
	if a.Key == slog.TimeKey {
		a.Key = "TimestampUtc"
	}
	return a
}
