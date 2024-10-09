package main

import (
	"os"

	"log/slog"
)

const SoftwareVersion = "0.1.1"

var config Config

// TODO parameterize
var programLevel = new(slog.LevelVar)
var log = slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: programLevel}))
