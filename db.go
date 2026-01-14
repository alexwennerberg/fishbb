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
	var err error
	db, err = sql.Open("sqlite3", DBPath) // probably not safe variable
	if err != nil {
		panic(err)
	}
	return db
}

//go:embed schema.sql
var sqlSchema string

func initdb() {
	dbname := DBPath
	var err error
	db, err = sql.Open("sqlite3", dbname+"?_txlock=immediate")
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
	// squash errors for idempotence
	createForum("General", "General discussion")
	// create admin / admin
	createUser("admin", "webmaster@foo", "admin", RoleAdmin)

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
	LoginInit(LoginInitArgs{
		Db: db,
	})
	db.Close()
}
