package main

type Board struct {
	ID          int
	Name        string
	Slug        string
	Description string
}

func createBoard(name string, description string, ownerid int) error {
	_, err := db.Exec("insert into board (name, slug, description, ownerid) values (?, ?, ?, ?)", name, slugify(name), description, ownerid)
	return err
}

func getBoards() ([]Board, error) {
	var boards []Board
	rows, err := db.Query("select id, name, slug from board")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var b Board
		err := rows.Scan(&b.ID, &b.Name, &b.Slug)
		if err != nil {
			return nil, err
		}
		boards = append(boards, b)
	}
	return boards, nil
}

func getBoard(slug string) (Board, error) {
	row := db.QueryRow("select id, name, slug from board where slug = ?", slug)
	var b Board
	err := row.Scan(&b.ID, &b.Name, &b.Slug)
	return b, err
}
