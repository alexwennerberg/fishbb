package main

import (
	"time"

	"github.com/alexedwards/argon2id"
)

type Role string

type User struct {
	ID       int
	Username string
	Email    string
	Capabilities int
	// TODO fix null schema
	Role     *Role
	Active   *bool
	About    *string
	Website  *string
	Created  time.Time
	Posts int
}

var Admin Role = "admin"
var Mod Role = "mod"
var Regular Role = "user"

func createUser(username, email, password string, role Role) error {
	hash, err := argon2id.CreateHash(password, argon2id.DefaultParams)
	if err != nil {
		return err
	}
	_, err = stmtCreateUser.Exec(username, email, hash, role)
	return err
}

func getUser(id int) User {
	row := stmtGetUser.QueryRow(id)
	var u User
	err := row.Scan(&u.ID, &u.Username, &u.Email, &u.Capabilities, &u.Role, &u.Active, &u.About, &u.Website, &u.Created, &u.Posts)
	logIfErr(err)
	return u
}
