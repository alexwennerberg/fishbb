package main

import (
	"fmt"
	"time"
)

type Thread struct {
	ID      int
	Title   string
	Author  User
	Created time.Time
	Pinned  bool
	Locked  bool
	Latest Post
	Replies int
}

// TODO paginate
func getThreads(forumID, limit, offset int) []Thread {
	var threads []Thread
	rows, _ := stmtGetThreads.Query(forumID)
	for rows.Next() {
		var t Thread
		err := rows.Scan(
			&t.ID, &t.Author.ID, &t.Author.Username, &t.Title, 
			&t.Created, &t.Pinned, &t.Locked,
			&t.Latest.ID, &t.Latest.Author.ID, &t.Latest.Author.Username, 
			&t.Latest.Created, &t.Replies)
		logIfErr(err)
		fmt.Print(t.Replies)
		threads = append(threads, t)
	}
	return threads
}

func getThread(threadid int) Thread {
	row := stmtGetThread.QueryRow(threadid)
	var t Thread
	err := row.Scan(&t.ID, &t.Title, &t.Author.ID, &t.Author.Username, &t.Created, &t.Pinned, &t.Locked)
	logIfErr(err)
	return t
}

// returns inserted thread ID
func createThread(authorid, forumid int, title string) (int64, error) {
	res, err := stmtCreateThread.Exec(authorid, forumid, title)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}
