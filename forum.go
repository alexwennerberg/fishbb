package main

import (
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
}

func createForum(name, description string) error {
	_, err := db.Exec("insert into forums (name, description, slug) values (?, ?, ?)", name, description, slugify(name))
	return err
}

func getForum(id int) (Forum, error) {
	row := db.QueryRow("select id, name, description, slug, read_permissions, write_permissions from forums where id = ?", id)
	var f Forum
	err := row.Scan(&f.ID, &f.Name, &f.Description, &f.Slug, &f.ReadPermissions, &f.WritePermissions)
	return f, err
}

// forum name should be invariant -- it messes with the URL
func updateForum(id int, description string, readRole Role, writeRole Role) error {
	_, err := db.Exec("update forums set description = ?, read_permissions = ?, write_permissions = ? where id = ?", description, readRole, writeRole, id)
	return err
}

func getForumBySlug(slug string) (Forum, error) {
	row := db.QueryRow("select id, name, description, slug, read_permissions, write_permissions from forums where slug = ?", slug)
	var f Forum
	err := row.Scan(&f.ID, &f.Name, &f.Description, &f.Slug, &f.ReadPermissions, &f.WritePermissions)
	return f, err
}

func getForumID(forumSlug string) int {
	row := db.QueryRow("select id from forums where slug = ?", forumSlug)
	var id int
	err := row.Scan(&id)
	logIfErr(err)
	return id
}

func getForums() ([]Forum, error) {
	var forums []Forum
	rows, err := db.Query(`
		select forums.id, name, description, read_permissions, write_permissions,
		coalesce(threadid, 0), coalesce(latest.title, ''), coalesce(latest.id, 0), coalesce(latest.authorid, 0),
		coalesce(latest.username, ''), coalesce(latest.created, ''),
		count(threads.id),
		coalesce(unique_users.user_count, 0)
		from forums
		left join (
			select threadid, threads.title, posts.id, threads.authorid,
			users.username, max(posts.created) as created, forumid
			from posts
			join users on users.id = posts.authorid
			join threads on posts.threadid = threads.id
			group by forumid
		) latest on latest.forumid = forums.id
		left join threads on threads.forumid = forums.id
		left join (
			select threads.forumid, count(distinct posts.authorid) as user_count
			from posts
			join threads on posts.threadid = threads.id
			group by threads.forumid
		) unique_users on unique_users.forumid = forums.id
		group by 1

`)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		var f Forum
		var created string
		err := rows.Scan(&f.ID, &f.Name, &f.Description, &f.ReadPermissions, &f.WritePermissions,
			&f.LastPost.ThreadID, &f.LastPost.ThreadTitle,
			&f.LastPost.ID,
			&f.LastPost.Author.ID, &f.LastPost.Author.Username, &created, &f.ThreadCount, &f.UniqueUsers)
		if err != nil {
			return nil, err
		}
		if created != "" {
			f.LastPost.Created, err = time.Parse(timeISO8601, created)
			logIfErr(err)
		}
		f.Slug = slugify(f.Name)
		forums = append(forums, f)
	}
	return forums, nil
}
