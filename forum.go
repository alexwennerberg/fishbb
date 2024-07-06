package main

import (
	"time"
)

type Forum struct {
	ID          int
	Name        string
	Description string
	Slug        string
	LastPost    PostSummary
}

func createForum(name, description string) error {
	_, err := stmtCreateForum.Exec(name, description, slugify(name))
	return err
}

// TODO error handling
func getForum(fid int) (Forum, error) {
	row := stmtGetForum.QueryRow(fid)
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

func getForums() []Forum {
	var forums []Forum
	rows, err := stmtGetForums.Query()
	logIfErr(err)
	for rows.Next() {
		var f Forum
		var created string
		err := rows.Scan(&f.ID, &f.Name, &f.Description,
			&f.LastPost.ThreadID, &f.LastPost.ThreadTitle,
			&f.LastPost.ID,
			&f.LastPost.Author.ID, &f.LastPost.Author.Username, &created)
		logIfErr(err)
		f.LastPost.Created, err = time.Parse(timeISO8601, created)
		logIfErr(err)
		f.Slug = slugify(f.Name)
		forums = append(forums, f)
	}
	return forums
}
