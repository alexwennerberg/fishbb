package main

import (
	"bytes"

	"github.com/BurntSushi/toml"
)

// non user-configurable config
var Port = ":8080"
var ViewDir = "./views/"
var DBPath = "fishbb.db"

// most of these don't work yet
type Config struct {
	// Signups require admin approval
	RequiresApproval bool
	BoardName        string
	BoardDescription string
	// TODO does nothing yet
	RequireEmailVerification bool
	// private instances are only visible to logged in users
	Private string

	// smaller forum settings
	PageSize int

	// A secret key used for generating CSRF tokens
	CSRFKey string

	// optional (for oauth)
	Domain                  string // todo not exactly
	GoogleOAuthClientID     string
	GoogleOAuthClientSecret string

	// optional (but required for email sending)
	SMTPUsername string
	SMTPPassword string
}

func (c Config) TOMLString() string {
	var b bytes.Buffer
	err := toml.NewEncoder(&b).Encode(c)
	if err != nil {
		panic(err) // TODO
	}
	return b.String()
}

func DefaultConfig() Config {
	return Config{
		BoardName:               "fishbb",
		BoardDescription:        "A discussion board",
		PageSize:                100,
		RequiresApproval:        true,
		Domain:                  "http://localhost:8080",
		GoogleOAuthClientID:     "",
		GoogleOAuthClientSecret: "",
	}
}

func GetConfig() (Config, error) {
	row := stmtGetConfig.QueryRow()
	var val string
	err := row.Scan(&val)
	if err != nil {
		return Config{}, err
	}
	var c Config
	_, err = toml.Decode(val, &c)
	if err != nil {
		return Config{}, err
	}
	return c, nil
}

// keeps a record of previous configs... TODO maybe remove
// includes cache
func UpdateConfig(c Config) error {
	_, err := stmtUpdateConfig.Exec(c.TOMLString())
	if err == nil {
		// update global config
		config = c
		// TODO find a better way to do this
		SetupGoogleOAuth()
	}
	return err
}
