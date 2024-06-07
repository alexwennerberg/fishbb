package main

import "log"

type Forum struct {
	ID          int
	Name        string
	Description string
	LastPost    PostSummary
}

func createForum(name, description string) error {
	_, err := stmtCreateForum.Exec(name, description)
	return err
}

func getForumName(forumID int) string {
	return "TODO"
}

func getForums() []Forum {
	var forums []Forum
	rows, _ := stmtGetForums.Query()
	for rows.Next() {
		var f Forum
		err := rows.Scan(&f.ID, &f.Name, &f.Description)
		if err != nil {
			// TODO
			log.Print(err)
		}
		forums = append(forums, f)
	}
	return forums
}
