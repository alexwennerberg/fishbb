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
	// TODO fix null schema
	Role     Role
	Active   bool
	About    string
	Website  string
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

func getUser(id int) (User, error) {
	row := stmtGetUser.QueryRow(id)
	var u User
	var created string
	err := row.Scan(&u.ID, &u.Username, &u.Email, &u.Role, &u.Active, &u.About, &u.Website, &created, &u.Posts)
	if err != nil {
		return u, err
	}
	u.Created, err = time.Parse(timeISO8601, created)
	return u, err
}

// used for self configuration
func updateMe(id int, about, website string) (error) {
	_, err := stmtUpdateMe.Exec(about, website, id)
   return err
}
