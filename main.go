package main

import (
	"flag"
	"fmt"
	"os"

	"log/slog"
)

var softwareVersion = "0.1.0"

var config Config
var devMode = true
var programLevel = new(slog.LevelVar)
var log = slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: programLevel}))

func main() {
	flag.BoolVar(&devMode, "dev", devMode, "dev mode")
	flag.Parse()
	args := flag.Args()

	if devMode {
		programLevel.Set(slog.LevelDebug)
	}
	cmd := "run"
	if len(args) > 0 {
		cmd = args[0]
	}
	switch cmd {
	case "init":
		initdb()
	case "version":
		fmt.Println(softwareVersion)
		os.Exit(0)
	case "run":
		serve()
	}
}
