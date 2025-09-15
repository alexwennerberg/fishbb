package main

import (
	"database/sql"
	"errors"
	"fmt"
	"regexp"
	"slices"
	"time"

	"github.com/alexedwards/argon2id"
)

type Role string

func (r Role) Can(req Role) bool {
	return slices.Index(hierarchy, r) >= slices.Index(hierarchy, req)
}

// lowest to highest, so that -1 = no role
var hierarchy = []Role{RoleNone, RoleInactive, RoleUser, RoleMod, RoleAdmin}

// Roles are hierarchical, admins can do everything mods can, and so on
var RoleAdmin Role = "admin"
var RoleMod Role = "mod"
var RoleUser Role = "user"
var RoleInactive Role = "inactive"

// Logged out, non-user
var RoleNone Role = ""

type User struct {
	ID       int
	Username string
	Email    string
	// Whether the user wishes to display email publicly
	EmailPublic bool
	// TODO fix null schema
	Role           Role
	About          string
	Website        string
	Created        time.Time
	Posts          int
}

var unameRegex, _ = regexp.Compile("^[a-zA-Z0-9]{1,25}$")

func validUsername(u string) bool {
	return unameRegex.MatchString(u)
}


func updatePassword(id int, password string) error {
	hash, err := argon2id.CreateHash(password, argon2id.DefaultParams)
	if err != nil {
		return err
	}
	_, err = stmtUpdatePassword.Exec(hash, id)
	return err
}

func createUser(username, password string, role Role) error {
	hash, err := argon2id.CreateHash(password, argon2id.DefaultParams)
	if err != nil {
		return err
	}
	_, err = stmtCreateUser.Exec(username, hash, role)
	return err
}


func getUser(username string) (*User, error) {
	row := stmtGetUser.QueryRow(username)
	var u User
	err := row.Scan(&u.ID, &u.Username, &u.Role, &u.Created, &u.Posts)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to scan row: %w", err)
	}
	return &u, err
}

func getUsers() ([]User, error) {
	var users []User
	rows, err := stmtGetUsers.Query()
	if err != nil {
		return nil, fmt.Errorf("could not execute query: %w", err)
	}
	for rows.Next() {
		var u User
		err := rows.Scan(&u.ID, &u.Username, &u.Email, &u.Role, &u.About, &u.Website, &u.Created, &u.Posts)
		if err != nil {
			return nil, fmt.Errorf("failed to scan rows: %w", err)
		}
		users = append(users, u)
	}
	return users, nil
}


func updateUserRole(id int, role Role) error {
	_, err := stmtUpdateUserRole.Exec(role, id)
	return err
}
