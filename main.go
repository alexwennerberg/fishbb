package main

import (
	"os"
)

func main() {
	// TODO remove this
	SingleInstance = true
	// temp hack
	if os.Getenv("USER") != "alex" {
		DBPath = "/var/lib/fishbb/fishbb.db"
	}
	Serve()
}
