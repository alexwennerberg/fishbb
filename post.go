package main

import (
	"fmt"
	"time"
)

type Post struct {
	ID      int
	Content string // TODO markdown
	Author  User
	Created time.Time 
	Edited  *time.Time
}

type PostSummary struct {
	ID          int
	Author      User
	ThreadID string
	ThreadTitle string
	Created   time.Time
}

func postValid(body string) bool {
	if len(body) > 10000 {
		return false
	}	
	return true
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

func getPost(postid int) (Post, error) {
	var p Post
	row := stmtGetPost.QueryRow(postid)
	err := row.Scan(&p.ID, &p.Content, &p.Author.ID, &p.Author.Username, &p.Created, &p.Edited)
	return p, err
}

// returns post id
func createPost(authorid int, threadid int, body string) (int64, error) {
	res, err := stmtCreatePost.Exec(threadid, authorid, body)

	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func editPost(postid int, content string) error {
	res, err := stmtEditPost.Exec(content, postid)
	aff, err :=res.RowsAffected()
	if err != nil {
		return err
	}
	if aff != 1 {
		return fmt.Errorf("unexpected number of rows affected, should be 1, was %d", aff)
	}
	return err
}

func deletePost(postid int) error {
	_, err := stmtDeletePost.Exec(postid)
	return err
}
