package main

import (
	"database/sql"
	"strings"

	_ "embed"

	_ "github.com/mattn/go-sqlite3"
)

var db *sql.DB

func opendb() *sql.DB {
	initdb()
	db, err := sql.Open("sqlite3", DBPath) // probably not safe variable
	if err != nil {
		panic(err)
	}
	return db
}

//go:embed schema.sql
var sqlSchema string

func initdb() {
	dbname := DBPath
	db, err := sql.Open("sqlite3", dbname+"?_txlock=immediate")
	if err != nil {
		panic(err)
	}
	// TODO -- some better way to do exec these
	for _, line := range strings.Split(sqlSchema, ";\n") {
		_, err = db.Exec(line)
		if err != nil {
			log.Error(line, "error", err)
			return
		}
	}
	PrepareStatements(db)
	// squash errors for idempotence
	createForum("General", "General discussion")
	// create admin / admin
	createUser("admin", "admin", RoleAdmin)

	// TODO... config
	config, err = GetConfig()
	if err != nil {
		config := DefaultConfig()
		err = SaveConfig(config)
	}
	if err != nil {
		panic(err)
	}
	csrf, err := GenerateRandomString(16)
	if err != nil {
		panic(err)
	}
	err = UpdateConfig("csrfkey", csrf)
	if err != nil {
		panic(err)
	}
	db.Close()
}

func prepare(db *sql.DB, stmt string) *sql.Stmt {
	s, err := db.Prepare(stmt)
	if err != nil {
		panic(err.Error() + stmt)
	}
	return s
}

var stmtGetForumID, stmtUpdatePassword, stmtSearchPosts,
	stmtEditPost, stmtGetPost, stmtGetPostSlug, stmtGetForum, 
	stmtGetForumBySlug, stmtCreateUser, stmtGetForums, stmtUpdateForum,
	stmtGetUser, stmtGetUserIDByEmail, stmtGetUsers, stmtGetPostAuthorID, stmtDeletePost,
	stmtThreadPin, stmtThreadLock, stmtActivateUser, stmtGetAllUsernames,
	stmtCreatePost, stmtGetThread, stmtGetPosts, stmtGetPostsByUser, stmtGetThreadCount,
	stmtUpdateMentionsChecked,
	stmtDeleteUser, stmtUpdateUserRole, stmtUpdateBanStatus, stmtUpdateConfig, stmtGetConfig,
	stmtGetThreads, stmtCreateThread, stmtCreateForum *sql.Stmt

func PrepareStatements(db *sql.DB) {
	stmtGetForums = prepare(db, `
		select forums.id, name, description, 
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
	stmtGetForum = prepare(db, "select id, name, description, slug from forums where id = ?")
	stmtGetForumBySlug = prepare(db, "select id, name, description, slug from forums where slug = ?")
	stmtGetForumID = prepare(db, "select id from forums where slug = ?")
	stmtUpdateForum = prepare(db, "update forums set description = ? where id = ?")
	stmtCreateForum = prepare(db, "insert into forums (name, description, slug) values (?, ?, ?)")
	stmtCreateUser = prepare(db, "insert into users (username, hash, role) values (?, ?, ?)")
	// kinda awk
	stmtGetUser = prepare(db, `
		select users.id,username,role,users.created,count(posts.id)
		from users 
		left join posts on users.id = posts.authorid
		where users.username = ?  
		group by users.id
		`)
	stmtGetUsers = prepare(db, `
		select users.id,username,role,users.created, count(1)
		from users 
		left join posts on users.id = posts.authorid
		group by users.id
		order by role, users.created desc
		`)
	stmtCreateThread = prepare(db, "insert into threads (authorid, forumid, title) values (?, ?, ?);")
	stmtCreatePost = prepare(db, "insert into posts (threadid, authorid, content) values (?, ?, ?)")
	stmtGetPostAuthorID = prepare(db, "select authorid from posts where id = ?")
	stmtGetThreads = prepare(db, `
		select threadid, forumid, threads.authorid, users.username, title, 
		threads.created, 
		latest.id, latest.authorid, latest.username, latest.content,
		latest.created, latest.replies - 1
		from threads 
		join users on users.id = threads.authorid
		join (
			select threadid, posts.id, authorid, users.username, posts.content, max(posts.created) as created, count(1) as replies
			from posts 
			join users on users.id = posts.authorid
			group by threadid
		) latest
		on latest.threadid = threads.id
		where forumid = ?
		order by latest.created desc limit ? offset ?
	`)
	stmtGetThread = prepare(db, `
		select threads.id, forumid, title, threads.authorid, users.username, 
		threads.created, count(1) - 1 as replies 
		from threads 
		join users on users.id = threads.authorid
		join posts on threads.id = posts.threadid
		join forums on threads.forumid = forums.id
		where threads.id = ?`)
	stmtGetThreadCount = prepare(db, "select count(1) from threads where forumid = ?")
	stmtGetPosts = prepare(db, `
		select posts.id, content, users.id, users.username, posts.created, posts.edited 
		from posts 
		join users on posts.authorid = users.id 
		where threadid = ? limit ? offset ?`)
	stmtGetPostsByUser = prepare(db, `
		select posts.id, content, users.id, users.username, posts.created, posts.edited 
		from posts 
		join users on posts.authorid = users.id 
		where authorid = ? order by posts.id desc limit ? offset ?`)
	stmtSearchPosts = prepare(db, `
		select posts.id, content, users.id, users.username, posts.created, posts.edited 
		from posts 
		join users on posts.authorid = users.id 
		where content like ? order by posts.id desc limit 1000`) // TODO paginate
	stmtGetPost = prepare(db, `
		select posts.id, content, users.id, users.username, posts.created, posts.edited 
		from posts 
		join users on posts.authorid = users.id 
		where posts.id = ?`)
	stmtEditPost = prepare(db, "update posts set content = ?, edited = current_timestamp where id = ?")
	stmtDeletePost = prepare(db, "delete from posts where id = ?")
	stmtUpdatePassword = prepare(db, "update users set hash = ? where id = ?")
	stmtDeleteUser = prepare(db, "delete from users where id = ?")
	stmtUpdateUserRole = prepare(db, "update users set role = ? where id = ?")

	// get stuff we need for a post slug
	stmtGetPostSlug = prepare(db, `
		select 
		threads.id, 
		forums.slug,
		(select count(1) from posts where threads.id = posts.threadid and id < ?1) as count
		from posts 
		left join threads on posts.threadid = threads.id
		left join forums on threads.forumid = forums.id
		where posts.id = ?1
	`)
	stmtUpdateConfig = prepare(db, "insert into config(key,value) values(?1,?2) on conflict(key) do update set value = ?2")
	stmtGetConfig = prepare(db, "select value from config where key = ?")
	stmtGetAllUsernames = prepare(db, "select username from users;")
	LoginInit(LoginInitArgs{
		Db: db,
	})
}
