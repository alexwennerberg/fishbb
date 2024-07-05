package main

import (
	"time"
)

type Thread struct {
	ID      int
	ForumID int
	Title   string
	Author  User
	Created time.Time
	Pinned  bool
	Locked  bool
	Latest  Post
	Replies int
}

// returns page list
func pageArray(n int) []int {
	c := ((n + 1) / config.PageSize)
	p := make([]int, c)
	for i := range c {
		p[i] = i+1
	}
	return p
}

func getThreadCount(forumID int ) (int, error) {
	var c int
	row := stmtGetThreadCount.QueryRow(forumID)
	err := row.Scan(&c)
	return c, err
}

// TODO fix case when no threads
func getThreads(forumID, page int) ([]Thread, error) {
	var threads []Thread
	limit, offset := paginate(page)
	rows, _ := stmtGetThreads.Query(forumID, limit, offset)
	for rows.Next() {
		var t Thread
		var created string
		err := rows.Scan(
			&t.ID, &t.ForumID, &t.Author.ID, &t.Author.Username, &t.Title, 
			&t.Created, &t.Pinned, &t.Locked,
			&t.Latest.ID, &t.Latest.Author.ID, &t.Latest.Author.Username, 
			&t.Latest.Content,
			&created, &t.Replies)
		logIfErr(err)
		t.Latest.Created, err = time.Parse(timeISO8601, created)
		logIfErr(err)
		threads = append(threads, t)
	}
	return threads, nil
}

func getThread(threadid int) (Thread, error) {
	row := stmtGetThread.QueryRow(threadid)
	var t Thread
	err := row.Scan(&t.ID, &t.ForumID, &t.Title, &t.Author.ID, &t.Author.Username, &t.Latest.Created, &t.Pinned, &t.Locked, &t.Replies)
	return t, err
}

// returns inserted thread ID
func createThread(authorid, forumid int, title string) (int64, error) {
	res, err := stmtCreateThread.Exec(authorid, forumid, title)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}
