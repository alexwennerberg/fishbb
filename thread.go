package main

import (
	"fmt"
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
func (t Thread) Pages() []int {
	// + 1 + 1
	c := ((t.Replies + 2) / config.PageSize)
	p := make([]int, c)
	for i := range c {
		p[i] = i+1
	}
	return p
}

// TODO fix case when no threads
func getThreads(forumID, page int) []Thread {
	var threads []Thread
	limit, offset := paginate(page)
	rows, _ := stmtGetThreads.Query(forumID, limit, offset)
	for rows.Next() {
		var t Thread
		var created string
		var tcreated string
		err := rows.Scan(
			&t.ID, &t.ForumID, &t.Author.ID, &t.Author.Username, &t.Title, 
			&tcreated, &t.Pinned, &t.Locked,
			&t.Latest.ID, &t.Latest.Author.ID, &t.Latest.Author.Username, 
			&created, &t.Replies)
		logIfErr(err)
		t.Latest.Created, err = time.Parse(timeISO8601, created)
		logIfErr(err)
		t.Created, err = time.Parse(timeISO8601, tcreated)
		fmt.Print(t.Replies)
		threads = append(threads, t)
	}
	return threads
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
