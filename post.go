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

// requires a lot of stuff
// TODO -- maybe optimize into one query
func getPostSlug(postid int) (string, error) {
	var threadid int 
	var forumname string
	var count int
	row := stmtGetPostSlug.QueryRow(postid)
	err := row.Scan(&threadid, &forumname, &count)
	if err != nil {
		return "", err
	}
	// TODO fix bug here
	lastPage := (count + 1)/ config.PageSize
	var url string
	// TODO url builder
	if lastPage != 1 {
		url = fmt.Sprintf("/f/%s/%d?page=%d#%d", forumname, threadid, lastPage, postid) 
	} else {
		url = fmt.Sprintf("/f/%s/%d#%d", forumname,threadid, postid) 
	}
	return url, nil
}

// page is 1-indexed
func getPosts(threadid, page int) []Post {
	var posts []Post
	limit, offset := paginate(page)
	rows, _ := stmtGetPosts.Query(threadid, limit, offset)
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
