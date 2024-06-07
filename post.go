package main

import "time"

type Post struct {
}

type PostSummary struct {
	ID          int
	Author      string
	ThreadTitle string
	CreatedAt   time.Time
}

// TODO paginate
func getPosts(limit, offset int) []Post {
	return nil
}

func createPost() {
}

func updatePost() {
}

func deletePost(id int) error {
	return nil
}

func reportPost(id int) error {
	return nil
}
