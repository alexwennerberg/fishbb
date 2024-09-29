package main

import (
	fishbb "git.sr.ht/~aw/fishbb/server"
)

// Run a single instance of fishBB
func main() {
	fishbb.SingleInstance = true
	fishbb.Serve()
}
