package server

import (
	"bufio"
	"fmt"
	"strings"
	"time"
)

type Post struct {
	ID      int
	Content string
	Author  User
	// TODO less ad-hoc
	ThreadID        string
	ThreadTitle     string
	ThreadPostCount int
	Created         time.Time
	Edited          *time.Time
}

// This does an inefficient db call for now
// TODO make it work with joins
func (p Post) Slug() string {
	s, _ := getPostSlug(p.ID)
	return s
}

const previewLength = 150

// unused atm
func (p Post) Preview() string {
	var text string
	scanner := bufio.NewScanner(strings.NewReader(p.Content))
	for scanner.Scan() {
		t := scanner.Text()
		// TODO better stripping of gmi directives maybe
		if !(strings.HasPrefix(t, ">") || strings.HasSuffix(t, "wrote:") || strings.HasPrefix(t, "=>") || strings.HasPrefix(t, "```")) {
			text += t
		}
	}
	if len(text) > previewLength-3 {
		return text[:previewLength-3] + "..."
	}
	if text == "" {

	}
	return text
}

func postValid(body string) bool {
	if len(body) > 10000 {
		return false
	}
	return true
}

// requires a lot of stuff
func getPostSlug(postid int) (string, error) {
	var threadid int
	var forumname string
	var count int
	row := stmtGetPostSlug.QueryRow(postid)
	err := row.Scan(&threadid, &forumname, &count)
	if err != nil {
		return "", err
	}
	postPage := ((count) / PageSize) + 1
	var url string
	// TODO url builder
	if postPage != 1 {
		url = fmt.Sprintf("/f/%s/%d?p=%d#%d", forumname, threadid, postPage, postid)
	} else {
		url = fmt.Sprintf("/f/%s/%d#%d", forumname, threadid, postid)
	}
	return url, nil
}

// TODO maybe consolidate with query builder
func searchPosts(q string) ([]Post, error) {
	var posts []Post
	rows, err := stmtSearchPosts.Query("%" + q + "%")
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		var p Post
		err := rows.Scan(&p.ID, &p.Content, &p.Author.ID, &p.Author.Username, &p.Created, &p.Edited)
		if err != nil {
			return nil, err
		}
		posts = append(posts, p)
	}
	return posts, nil
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

func getPostsByUser(uid int, page int) ([]Post, error) {
	var posts []Post
	limit, offset := paginate(page)
	rows, _ := stmtGetPostsByUser.Query(uid, limit, offset)
	for rows.Next() {
		var p Post
		err := rows.Scan(&p.ID, &p.Content, &p.Author.ID, &p.Author.Username, &p.Created, &p.Edited)
		if err != nil {
			return nil, err
		}
		posts = append(posts, p)
	}
	return posts, nil
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
	aff, err := res.RowsAffected()
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
