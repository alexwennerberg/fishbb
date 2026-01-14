package main

import (
	"fmt"
	"time"
)

type Forum struct {
	ID          int
	Name        string
	Description string
	Slug        string
	// lowest level that can view this for
	ReadPermissions  Role
	WritePermissions Role
	LastPost         Post
	ThreadCount      int
	UniqueUsers      int
	Board            Board
}

func createForum(name, description string, boardid int) error {
	_, err := db.Exec("insert into forum (name, description, slug, boardid) values (?, ?, ?, ?)", name, description, slugify(name), boardid)
	return err
}

func getForum(id int) (Forum, error) {
	row := db.QueryRow(`
		select forum.id, forum.name, forum.description, forum.slug, forum.read_permissions, forum.write_permissions,
		board.id, board.name, board.slug
		from forum
		join board on forum.boardid = board.id
		where forum.id = ?`, id)
	var f Forum
	var u string
	err := row.Scan(&f.ID, &f.Name, &f.Description, &u, &f.ReadPermissions, &f.WritePermissions,
		&f.Board.ID, &f.Board.Name, &f.Board.Slug)
	f.Slug = fmt.Sprintf("/%s/f/%s", f.Board.Slug, u)
	return f, err
}

// forum name should be invariant -- it messes with the URL
func updateForum(id int, description string, readRole Role, writeRole Role) error {
	_, err := db.Exec("update forum set description = ?, read_permissions = ?, write_permissions = ? where id = ?", description, readRole, writeRole, id)
	return err
}

func getForumBySlug(slug string) (Forum, error) {
	row := db.QueryRow(`
		select forum.id, forum.name, forum.description, forum.slug, forum.read_permissions, forum.write_permissions,
		board.id, board.name, board.slug
		from forum
		join board on forum.boardid = board.id
		where forum.slug = ?`, slug)
	var f Forum
	var u string
	err := row.Scan(&f.ID, &f.Name, &f.Description, &u, &f.ReadPermissions, &f.WritePermissions,
		&f.Board.ID, &f.Board.Name, &f.Board.Slug)
	f.Slug = fmt.Sprintf("/%s/f/%s", f.Board.Slug, u)
	return f, err
}

func getForumID(forumSlug string) int {
	row := db.QueryRow("select id from forum where slug = ?", forumSlug)
	var id int
	err := row.Scan(&id)
	logIfErr(err)
	return id
}

func getForums() ([]Forum, error) {
	var forums []Forum
	rows, err := db.Query(`
		select forum.id, forum.name, forum.description, forum.slug, forum.read_permissions, forum.write_permissions,
		coalesce(threadid, 0), coalesce(latest.title, ''), coalesce(latest.id, 0), coalesce(latest.authorid, 0),
		coalesce(latest.username, ''), coalesce(latest.created, ''),
		count(thread.id),
		coalesce(unique_users.user_count, 0),
		board.id, board.name, board.slug
		from forum
		join board on forum.boardid = board.id
		left join (
			select threadid, thread.title, post.id, thread.authorid,
			user.username, max(post.created) as created, forumid
			from post
			join user on user.id = post.authorid
			join thread on post.threadid = thread.id
			group by forumid
		) latest on latest.forumid = forum.id
		left join thread on thread.forumid = forum.id
		left join (
			select thread.forumid, count(distinct post.authorid) as user_count
			from post
			join thread on post.threadid = thread.id
			group by thread.forumid
		) unique_users on unique_users.forumid = forum.id
		group by 1

`)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		var f Forum
		var created string
		var forumSlug string
		err := rows.Scan(&f.ID, &f.Name, &f.Description, &forumSlug, &f.ReadPermissions, &f.WritePermissions,
			&f.LastPost.ThreadID, &f.LastPost.ThreadTitle,
			&f.LastPost.ID,
			&f.LastPost.Author.ID, &f.LastPost.Author.Username, &created, &f.ThreadCount, &f.UniqueUsers,
			&f.Board.ID, &f.Board.Name, &f.Board.Slug)
		if err != nil {
			return nil, err
		}
		if created != "" {
			f.LastPost.Created, err = time.Parse(timeISO8601, created)
			logIfErr(err)
		}
		f.Slug = fmt.Sprintf("/%s/f/%s", f.Board.Slug, forumSlug)
		forums = append(forums, f)
	}
	return forums, nil
}
