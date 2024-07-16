package main

import (
	"fmt"
	"net/mail"
	"regexp"
	"time"

	"github.com/alexedwards/argon2id"
)

type Role string

func (r Role) ModLevel() bool {
	return r == RoleMod || r == RoleAdmin
}

func (r Role) AdminLevel() bool {
	return r == RoleAdmin
}

type User struct {
	ID       int
	Username string
	Email    string
	// TODO fix null schema
	Role    Role
	Active  bool
	About   string
	Website string
	Created time.Time
	Posts   int // TODO perf
}

var RoleAdmin Role = "admin"
var RoleMod Role = "mod"
var RoleUser Role = "user"

// Logged out, non-user
var RoleNone Role = ""

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

func createUser(username, email, password string, role Role, active bool) error {
	hash, err := argon2id.CreateHash(password, argon2id.DefaultParams)
	if err != nil {
		return err
	}
	_, err = stmtCreateUser.Exec(username, email, hash, role, active)
	return err
}

func getUser(id int) (User, error) {
	row := stmtGetUser.QueryRow(id)
	var u User
	err := row.Scan(&u.ID, &u.Username, &u.Email, &u.Role, &u.Active, &u.About, &u.Website, &u.Created, &u.Posts)
	if err != nil {
		return u, fmt.Errorf("failed to scan row: %w", err)
	}
	return u, err
}

func getUsers() ([]User, error) {
	var users []User
	rows, err := stmtGetUsers.Query()
	if err != nil {
		return nil, fmt.Errorf("could not execute query: %w", err)
	}
	for rows.Next() {
		var u User
		err := rows.Scan(&u.ID, &u.Username, &u.Email, &u.Role, &u.Active, &u.About, &u.Website, &u.Created, &u.Posts)
		if err != nil {
			return nil, fmt.Errorf("failed to scan rows: %w", err)
		}
		users = append(users, u)
	}
	fmt.Println(users)
	return users, nil
}

func activateUser(id int) error {
	_, err := stmtActivateUser.Exec(id)
	return err
}

func updateUserBanStatus(id int, banned bool) error {
	_, err := stmtUpdateBanStatus.Exec(!banned, id)
	return err
}

func deleteUser(id int) error {
	_, err := stmtDeleteUser.Exec(id)
	return err
}

func updateUserRole(id int, role Role) error {
	_, err := stmtUpdateUserRole.Exec(role, id)
	return err
}

// used for self configuration
func updateMe(id int, about, website string) error {
	_, err := stmtUpdateMe.Exec(about, website, id)
	return err
}
