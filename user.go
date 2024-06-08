package main

import (
	"time"

	"golang.org/x/crypto/bcrypt"
)

type Role string

type User struct {
	ID       int
	Username string
	Email    string
	Role     *Role
	Active   *bool
	About    *string
	Website  *string
	Created  time.Time
}

var Admin Role = "admin"
var Mod Role = "mod"
var Regular Role = ""

func createUser(username, email, password string, role Role) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	if err != nil {
		return err
	}
	_, err = stmtCreateUser.Exec(username, email, hash)
	return err
}

func getUser(id int) User {
	row := stmtGetUser.QueryRow(id)
	var u User
	err := row.Scan(&u.ID, &u.Username, &u.Email, &u.Role, &u.Active, &u.About, &u.Website, &u.Created)
	logIfErr(err)
	return u
}
