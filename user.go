package main

import (
	"net/mail"
	"regexp"
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

var RoleAdmin Role = "admin"
var RoleMod Role = "mod"
var RoleUser Role = "user"

var unameRegex, _ = regexp.Compile("^[a-zA-Z0-9]{1,25}$")

func validUsername(u string) bool {
	return unameRegex.MatchString(u)
}

func validEmail(e string) bool {
	_, err := mail.ParseAddress(e)
	return err == nil
}

func validPassword(p string) bool {
	return len(p) >= 8
}

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
	err := row.Scan(&u.ID, &u.Username, &u.Email, &u.Role, &u.Active, &u.About, &u.Website, &u.Created, &u.Posts)
	if err != nil {
		return u, err
	}
	return u, err
}

// used for self configuration
func updateMe(id int, about, website string) (error) {
	_, err := stmtUpdateMe.Exec(about, website, id)
   return err
}
