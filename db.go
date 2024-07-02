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
	stmtGetUser, stmtGetPostAuthorID, stmtDeletePost,
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

	db, err := sql.Open("sqlite3", dbname + "?_txlock=immediate")
	if err != nil {
		panic(err)
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

	err = createForum("General", "General discussion")
	if err != nil {
		panic(err)
	}
	if devMode { // create admin / admin
		err := createUser("admin", "webmaster@foo", "admin", Admin)
		if err != nil {
			panic(err) // TODO
		}
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
	stmtGetForums = prepare(db, `
		select forums.id, name, description,
		threadid, latest.title, latest.id, latest.authorid,
		latest.username, latest.created
		from forums
		left join (
			select threadid, threads.title, posts.id, threads.authorid,
			users.username, max(posts.created) as created, forumid	
			from posts 
			join users on users.id = posts.authorid
			join threads on posts.threadid = threads.id
			group by forumid 
		) latest on latest.forumid = forums.id

`)
	stmtGetForum = prepare(db, "select id, name, description, slug from forums where id = ?")
	stmtGetForumID = prepare(db, "select id from forums where slug = ?")
	stmtCreateForum = prepare(db, "insert into forums (name, description, slug) values (?, ?, ?)")
	stmtCreateUser = prepare(db, "insert into users (username, email, hash, role) values (?, ?, ?, ?)")
	stmtGetUser = prepare(db, `
		select users.id,username,email,role,active,about,website,users.created, count(1)
		from users 
		join posts on users.id = posts.authorid
		where users.id = ?  
		group by users.id
		`)
	stmtCreateThread = prepare(db, "insert into threads (forumid, authorid, title) values (?, ?, ?);")
	stmtCreatePost = prepare(db, "insert into posts (threadid, authorid, content) values (?, ?, ?)")
	stmtGetPostAuthorID = prepare(db, "select authorid from posts where id = ?")
	stmtDeletePost = prepare(db, "delete from posts where id = ?")
	stmtGetThreads = prepare(db, `
		select threadid, forumid, threads.authorid, users.username, title, 
		threads.created, threads.pinned, threads.locked,
		latest.id, latest.authorid, latest.username, latest.created,
		latest.replies - 1
		from threads 
		join users on users.id = threads.authorid
		join (
			select threadid, posts.id, authorid, users.username, max(posts.created) as created, count(1) as replies
			from posts 
			join users on users.id = posts.authorid
			group by threadid
		) latest
		on latest.threadid = threads.id
		where forumid = ?
		order by latest.created desc
	`)
	stmtGetThread = prepare(db, `
		select threads.authorid, forumid, title, threads.authorid, users.username, 
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
