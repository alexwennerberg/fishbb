package main

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

func getForum(fid int) Forum {
	row := stmtGetForum.QueryRow(fid)
	var f Forum
	err := row.Scan(&f.ID, &f.Name, &f.Description, &f.Slug)
	logIfErr(err)
	return f
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
	rows, _ := stmtGetForums.Query()
	for rows.Next() {
		var f Forum
		err := rows.Scan(&f.ID, &f.Name, &f.Description)
		logIfErr(err)
		f.Slug = slugify(f.Name)
		forums = append(forums, f)
	}
	return forums
}
