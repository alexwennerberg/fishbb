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

// TODO paginate
// TODO fix case when no threads
func getThreads(forumID, limit, offset int) []Thread {
	var threads []Thread
	rows, _ := stmtGetThreads.Query(forumID)
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

func getThread(threadid int) Thread {
	row := stmtGetThread.QueryRow(threadid)
	var t Thread
	var created string
	err := row.Scan(&t.ID, &t.ForumID, &t.Title, &t.Author.ID, &t.Author.Username, &created, &t.Pinned, &t.Locked)
	logIfErr(err)
	t.Latest.Created, err = time.Parse(timeISO8601, created)
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
