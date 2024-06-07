package main

import "golang.org/x/crypto/bcrypt"

type Role string

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

func updateUserDetails() {
}
