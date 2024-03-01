package main

import (
	"log/slog"
	"os"
)

func slogSetup(json bool, tag string) {
	var logger *slog.Logger

	if json {
		logger = slog.New(slog.NewJSONHandler(os.Stdout, nil))
	} else {
		logger = slog.New(slog.NewTextHandler(os.Stdout, nil))
	}

	if tag != "" {
		logger = logger.With("tag", tag)
	}

	slog.SetDefault(logger)
}
