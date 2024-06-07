package main

import (
	"database/sql"
	"log"
)

var stmtGetSubForums, stmtCreateForum *sql.Stmt
var db *sql.DB

func opendb() {
	db, err := sql.Open("sqlite3", "fishbb.db")
	if err != nil {
		log.Fatal(err)
	}
}

func prepare(db *sql.DB, stmt string) *sql.Stmt {
	stmt, err := db.prepare(s)
	if err != nil {
		log.Fatalf("error %s: %s", err, s)
	}
	return stmt
}

func prepareStatements(db *sql.DB) {
	stmtGetForums = prepare(db, "select id, name, description from forums")
	stmtCreateForum = prepare(db, "insert into forums (name, description) values (?, ?)")
	stmtCreateThread = prepare(db, "insert into threads (forumid, authorid, title, created) valueus (?, ?, ?, ?);")
}
