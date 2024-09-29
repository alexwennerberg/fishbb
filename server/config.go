package server

import "strconv"

// non user-configurable config
var Port = ":8080"
var DBPath = "fishbb.db"

// Changing this will break existing URLs
const PageSize int = 50

// TODO -- start gating features on self hosted or not
var SingleInstance = false

// most of these don't work yet
type Config struct {
	// Whether new signups require admin approval before users can post
	RequiresApproval bool
	// The title of the bulletin board (NOT CONFIGURABLE)
	BoardName string
	// The description of the bulletin board
	BoardDescription string

	// optional (for oauth)
	Domain                  string // todo not exactly
	GoogleOAuthClientID     string
	GoogleOAuthClientSecret string

	// optional (but required for email sending)
	SMTPUsername string
	SMTPPassword string
}

// in multi-instance, config values that are shared by the cluster
// TODO
type SharedConfig struct {
}

func DefaultConfig() Config {
	return Config{
		BoardName:               "fishbb",
		BoardDescription:        "A discussion board",
		RequiresApproval:        true,
		Domain:                  "http://localhost:8080",
		GoogleOAuthClientID:     "",
		GoogleOAuthClientSecret: "",
	}
}

func SaveConfig(c Config) error {
	UpdateConfig("board-name", c.BoardName)
	UpdateConfig("board-description", c.BoardDescription)
	UpdateConfig("requires-approval", c.RequiresApproval)
	return nil
}

// get all config values TODO cleanup
func GetConfig() (Config, error) {
	var c Config
	// TODO cleanup
	c.BoardDescription, _ = GetConfigValue("board-description")
	c.BoardName, _ = GetConfigValue("board-description")
	if SingleInstance {
		c.GoogleOAuthClientID, _ = GetConfigValue("google-oauth-client-id")
		c.GoogleOAuthClientSecret, _ = GetConfigValue("google-oauth-client-secret")
		c.SMTPUsername, _ = GetConfigValue("smtp-username")
		c.SMTPPassword, _ = GetConfigValue("smtp-password")
	}
	r, _ := GetConfigValue("requires-approval")
	c.RequiresApproval, _ = strconv.ParseBool(r)
	return c, nil
}

func GetConfigValue(key string) (string, error) {
	row := stmtGetConfig.QueryRow(key)
	var val string
	err := row.Scan(&val)
	if err != nil {
		return "", err
	}
	return val, nil
}

func UpdateConfig(key string, value any) error {
	_, err := stmtUpdateConfig.Exec(key, value)
	return err
}
