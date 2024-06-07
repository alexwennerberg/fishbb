package main

type Thread struct {
	Title string
	// author User
	Pinned bool
	Locked bool
}

// TODO paginate
func getThreads(forumID, limit, offset int) []Thread {
	return nil
}

func createThread() {
}
