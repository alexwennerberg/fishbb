package main

import (
	"flag"
	"fmt"
	"os"
)

var softwareVersion = "0.1.0"

var config Config
var devMode = true

func main() {
	config = NewConfig() // TODO config story
	flag.BoolVar(&devMode, "dev", devMode, "dev mode")
	flag.Parse()
	args := flag.Args()
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
