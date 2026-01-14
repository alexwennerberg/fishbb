package main

import (
	"database/sql"
	"errors"
	"fmt"
	"net/mail"
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
	MentionsUnread int
}

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

func updatePassword(id int, password string) error {
	hash, err := argon2id.CreateHash(password, argon2id.DefaultParams)
	if err != nil {
		return err
	}
	_, err = db.Exec("update users set hash = ? where id = ?", hash, id)
	return err
}

func createUser(username, email, password string, role Role) error {
	hash, err := argon2id.CreateHash(password, argon2id.DefaultParams)
	if err != nil {
		return err
	}
	_, err = db.Exec("insert into users (username, email, hash, role) values (?, ?, ?, ?)", username, email, hash, role)
	return err
}

func getUserIDByEmail(email string) (*int, error) {
	row := db.QueryRow("select users.id from users where users.email = ?", email)
	var id int
	err := row.Scan(&id)
	if errors.Is(err, sql.ErrNoRows) {
		// return no value
		return nil, nil
	}
	return &id, err
}

func getUser(username string) (*User, error) {
	row := db.QueryRow(`
		select users.id,username,email,email_public,role,about,website,users.created,count(posts.id),mentions_checked
		from users
		left join posts on users.id = posts.authorid
		where users.username = ?
		group by users.id
		`, username)
	var mentionsChecked time.Time
	var u User
	err := row.Scan(&u.ID, &u.Username, &u.Email, &u.EmailPublic, &u.Role, &u.About, &u.Website, &u.Created, &u.Posts, &mentionsChecked)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to scan row: %w", err)
	}
	// probably super inefficient
	row = db.QueryRow(`select count(1) from posts where content like ? and created > ?`, fmt.Sprintf("%%@%s%%", u.Username), mentionsChecked)
	fmt.Println(u.MentionsUnread)
	err = row.Scan(&u.MentionsUnread)
	if err != nil {
		return nil, err
	}
	return &u, err
}

func getUsers() ([]User, error) {
	var users []User
	rows, err := db.Query(`
		select users.id,username,email,role,about,website,users.created, count(1)
		from users
		left join posts on users.id = posts.authorid
		group by users.id
		order by role, users.created desc
		`)
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

// unused
func getAllUsernames() ([]string, error) {
	var usernames []string
	rows, err := db.Query("select username from users;")
	if err != nil {
		return nil, fmt.Errorf("could not execute query: %w", err)
	}
	for rows.Next() {
		var username string
		err := rows.Scan(&username)
		if err != nil {
			return nil, fmt.Errorf("failed to scan rows: %w", err)
		}
		usernames = append(usernames, username)
	}
	return usernames, nil
}

func activateUser(id int) error {
	_, err := db.Exec("update users set role = 'user' where id = ?", id)
	return err
}

func updateUserBanStatus(id int, banned bool) error {
	role := "inactive"
	if !banned {
		role = "user"
	}
	_, err := db.Exec("update users set role = ? where id = ?", role, id)
	return err
}

func deleteUser(id int) error {
	_, err := db.Exec("delete from users where id = ?", id)
	return err
}

func updateUserRole(id int, role Role) error {
	_, err := db.Exec("update users set role = ? where id = ?", role, id)
	return err
}

func setNotificationsRead(id int) error {
	_, err := db.Exec("update users set mentions_checked = ? where id = ?", time.Now().UTC(), id)
	return err
}

// doesn't include all fields
func updateMe(u User) error {
	_, err := db.Exec("update users set username = ?, email = ?, email_public = ?, about = ?, website = ? where id = ?", u.Username, u.Email, u.EmailPublic, u.About, u.Website, u.ID)
	return err
}
