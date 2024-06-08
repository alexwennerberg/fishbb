package main

import (
	"database/sql"
	"fishbb/login"
	"os"
	"strings"

	_ "embed"

	_ "github.com/mattn/go-sqlite3"
)

var stmtGetForumID,
	stmtGetForum, stmtCreateUser, stmtGetForums,
	stmtGetUser,
	stmtCreatePost, stmtGetThread, stmtGetPosts,
	stmtGetThreads, stmtCreateThread, stmtCreateForum *sql.Stmt
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
		log.Error("unexpected error:", "error", err)
		return
	}
	for _, line := range strings.Split(sqlSchema, ";") {
		_, err = db.Exec(line)
		if err != nil {
			log.Error(line, "error", err)
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
	stmtGetForum = prepare(db, "select id, name, description, slug from forums where id = ?")
	stmtGetForumID = prepare(db, "select id from forums where slug = ?")
	stmtCreateForum = prepare(db, "insert into forums (name, description, slug) values (?, ?, ?)")
	stmtCreateUser = prepare(db, "insert into users (username, email, hash) values (?, ?, ?)")
	stmtGetUser = prepare(db, "select id,username,email,role,active,about,website,created from users where id = ?  ")
	stmtCreateThread = prepare(db, "insert into threads (forumid, authorid, title) values (?, ?, ?);")
	stmtCreatePost = prepare(db, "insert into posts (threadid, authorid, content) values (?, ?, ?)")
	stmtGetThreads = prepare(db, `
		select forumid, threads.authorid, users.username, title, 
		threads.created, threads.pinned, threads.locked
		from threads 
		join users on users.id = threads.authorid
		where forumid = ?
	`)
	stmtGetThread = prepare(db, `
		select threads.authorid, title, threads.authorid, users.username, 
		threads.created, threads.pinned, threads.locked
		from threads 
		join users on users.id = threads.authorid
		where threads.id = ?`)
	stmtGetPosts = prepare(db, `
		select posts.id, content, users.id, users.username, posts.created, posts.edited 
		from posts 
		join users on posts.authorid = users.id 
		where threadid = ?`)
	login.Init(login.InitArgs{
		Db: db,
	})
}
