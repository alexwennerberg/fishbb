package main

import (
	"database/sql"
	"fmt"
	"os"
	"strings"

	_ "embed"
	_ "github.com/mattn/go-sqlite3"

	"log"
)

var stmtCreateUser, stmtGetForums, stmtCreateThread, stmtCreateForum *sql.Stmt
var db *sql.DB

func opendb() *sql.DB {
	db, err := sql.Open("sqlite3", "fishbb.db")
	if err != nil {
		panic(err)
	}
	return db
}

//go:embed schema.sql
var sqlSchema string

func initdb() {
	// set csrf key
	dbname := config.DBPath
	fmt.Println(dbname)
	_, err := os.Stat(dbname)
	if err == nil {
		panic(dbname + "already exists")
	}

	db, err := sql.Open("sqlite3", dbname)
	if err != nil {
		panic(err)
	}
	_, err = db.Exec("PRAGMA journal_mode=WAL")
	if err != nil {
		log.Print(err)
		return
	}
	for _, line := range strings.Split(sqlSchema, ";") {
		_, err = db.Exec(line)
		if err != nil {
			log.Print(err)
			return
		}
	}
	prepareStatements(db)
	// set default values
	// Create admin user
	// Set csrfkey

	createForum("General", "General discussion")
	if devMode { // create admin / admin
		createUser("admin", "webmaster@foo", "admin", Admin)
	}

	db.Close()
	os.Exit(0)
}

func prepare(db *sql.DB, stmt string) *sql.Stmt {
	s, err := db.Prepare(stmt)
	if err != nil {
		panic(err.Error() + stmt)
	}
	return s
}

func prepareStatements(db *sql.DB) {
	stmtGetForums = prepare(db, "select id, name, description from forums")
	stmtCreateForum = prepare(db, "insert into forums (name, description) values (?, ?)")
	stmtCreateUser = prepare(db, "insert into users (username, email, hash) values (?, ?, ?)")
	stmtCreateThread = prepare(db, "insert into threads (forumid, authorid, title, created) values (?, ?, ?, ?);")
}
