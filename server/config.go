package server

import (
	"bytes"

	"github.com/BurntSushi/toml"
)

// non user-configurable config
var Port = ":8080"
var DBPath = "fishbb.db"

// most of these don't work yet
type Config struct {
	// Whether new signups require admin approval before users can post
	RequiresApproval bool
	// The title of the bulletin board
	BoardName string
	// The description of the bulletin board
	BoardDescription string

	// The size of pages on threads and forums
	PageSize int

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

func GetConfig(key string) (Config, error) {
	row := stmtGetConfig.QueryRow(key)
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

func UpdateConfig(key string, value string) error {
	_, err := stmtUpdateConfig.Exec(key, value)
	return err
}

// TODO stop this toml nonsense
func UpdateConfigTOML(c Config) error {
	_, err := stmtUpdateConfig.Exec("config-toml", c.TOMLString())
	if err == nil {
		// update global config
		config = c
		// TODO find a better way to do this
		SetupGoogleOAuth()
	}
	return err
}
