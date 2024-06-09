package main

import "time"

type Post struct {
	ID      int
	Content string // TODO markdown
	Author  User
	Created string // TODO
	Edited  *time.Time
}

type PostSummary struct {
	ID          int
	Author      string
	ThreadTitle string
	CreatedAt   time.Time
}

// TODO paginate
func getPosts(threadid, limit, offset int) []Post {
	var posts []Post
	rows, _ := stmtGetPosts.Query(threadid)
	for rows.Next() {
		var p Post
		err := rows.Scan(&p.ID, &p.Content, &p.Author.ID, &p.Author.Username, &p.Created, &p.Edited)
		logIfErr(err)
		posts = append(posts, p)
	}
	return posts
}

// returns post id
func createPost(authorid int, threadid int, body string) (int64, error) {
	// TODO markdown
	res, err := stmtCreatePost.Exec(authorid, threadid, body)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func editPost() {
}

func deletePost(id int) error {
	return nil
}

func reportPost(id int) error {
	return nil
}
