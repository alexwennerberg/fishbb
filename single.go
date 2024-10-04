package main

import (
	"os"

	fishbb "git.sr.ht/~aw/fishbb/server"
)

// Run a single instance of fishBB
func main() {
	fishbb.SingleInstance = true
	// temp hack
	if os.Getenv("USER") != "alex" {
		fishbb.DBPath = "/var/lib/fishbb/fishbb.db"
	}
	fishbb.Serve()
}
