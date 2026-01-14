package main

import (
	"os"
)

func main() {
	// temp hack
	if os.Getenv("USER") != "alex" {
		DBPath = "/var/lib/fishbb/fishbb.db"
	}
	Serve()
}
