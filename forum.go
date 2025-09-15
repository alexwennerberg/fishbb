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
	LastPost         Post
	ThreadCount      int
	UniqueUsers      int
}

func createForum(name, description string) error {
	_, err := stmtCreateForum.Exec(name, description, slugify(name))
	return err
}

func getForum(id int) (Forum, error) {
	row := stmtGetForum.QueryRow(id)
	var f Forum
	err := row.Scan(&f.ID, &f.Name, &f.Description, &f.Slug)
	return f, err
}

// forum name should be invariant -- it messes with the URL
func updateForum(id int, description string) error {
	_, err := stmtUpdateForum.Exec(description, id)
	return err
}

func getForumBySlug(slug string) (Forum, error) {
	row := stmtGetForumBySlug.QueryRow(slug)
	var f Forum
	err := row.Scan(&f.ID, &f.Name, &f.Description, &f.Slug)
	return f, err
}

func getForumID(forumSlug string) int {
	row := stmtGetForumID.QueryRow(forumSlug)
	var id int
	err := row.Scan(&id)
	logIfErr(err)
	return id
}

func getForums() ([]Forum, error) {
	var forums []Forum
	rows, err := stmtGetForums.Query()
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		var f Forum
		var created string
		err := rows.Scan(&f.ID, &f.Name, &f.Description, 
			&f.LastPost.ThreadID, &f.LastPost.ThreadTitle,
			&f.LastPost.ID,
			&f.LastPost.Author.ID, &f.LastPost.Author.Username, &created, &f.ThreadCount, &f.UniqueUsers)
		if err != nil {
			return nil, err
		}
		f.LastPost.Created, err = time.Parse(timeISO8601, created)
		logIfErr(err)
		f.Slug = slugify(f.Name)
		forums = append(forums, f)
	}
	return forums, nil
}
