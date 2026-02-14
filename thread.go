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

func getThreadCount(forumID int) (int, error) {
	var c int
	row := db.QueryRow("select count(1) from thread where forumid = ?", forumID)
	err := row.Scan(&c)
	return c, err
}

func getThreads(forumID, page int) ([]Thread, error) {
	var threads []Thread
	limit, offset := paginate(page)
	rows, err := db.Query(`
		select threadid, forumid, thread.authorid, user.username, title,
		thread.created, thread.pinned, thread.locked,
		latest.id, latest.authorid, latest.username, latest.content,
		latest.created, latest.replies - 1
		from thread
		join user on user.id = thread.authorid
		join (
			select threadid, post.id, authorid, user.username, post.content, max(post.created) as created, count(1) as replies
			from post
			join user on user.id = post.authorid
			group by threadid
		) latest
		on latest.threadid = thread.id
		where forumid = ?
		order by pinned desc, latest.created desc limit ? offset ?
	`, forumID, limit, offset)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		var t Thread
		var created string
		err := rows.Scan(
			&t.ID, &t.ForumID, &t.Author.ID, &t.Author.Username, &t.Title,
			&t.Created, &t.Pinned, &t.Locked,
			&t.Latest.ID, &t.Latest.Author.ID, &t.Latest.Author.Username,
			&t.Latest.Content,
			&created, &t.Replies)
		if err != nil {
			return nil, err
		}
		// TODO -- wonder if I can get away from this
		if created != "" {
			t.Latest.Created, err = time.Parse(timeISO8601, created)
			logIfErr(err)
		}
		threads = append(threads, t)
	}
	return threads, nil
}

func setThreadPin(threadid int, pinned bool) error {
	_, err := db.Exec("update thread set pinned = ? where id = ?", pinned, threadid)
	return err
}

func setThreadLock(threadid int, locked bool) error {
	_, err := db.Exec("update thread set locked = ? where id = ?", locked, threadid)
	return err
}

func getThread(threadid int) (Thread, error) {
	row := db.QueryRow(`
		select thread.id, forumid, title, thread.authorid, user.username,
		thread.created, thread.pinned, thread.locked,
		(select count(1) - 1 from post where threadid = thread.id) as replies
		from thread
		join user on user.id = thread.authorid
		where thread.id = ?`, threadid)
	var t Thread
	err := row.Scan(&t.ID, &t.ForumID, &t.Title, &t.Author.ID, &t.Author.Username, &t.Latest.Created, &t.Pinned, &t.Locked, &t.Replies)
	return t, err
}

// returns inserted thread ID
func createThread(authorid, forumid int, title string) (int64, error) {
	res, err := db.Exec("insert into thread (authorid, forumid, title) values (?, ?, ?);", authorid, forumid, title)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}
