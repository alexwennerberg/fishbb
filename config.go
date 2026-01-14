package main

import "strconv"

// non user-configurable config
var Port = ":8081"
var DBPath = "fishbb.db"

// Changing this will break existing URLs
const PageSize int = 50

// most of these don't work yet
type Config struct {
	// Whether new signups require admin approval before users can post
	RequiresApproval bool
	// The title of the bulletin board (NOT CONFIGURABLE)
	BoardName string
	// The description of the bulletin board
	BoardDescription string
}

// in multi-instance, config values that are shared by the cluster
// TODO
type SharedConfig struct {
}

func DefaultConfig() Config {
	return Config{
		BoardName:        "fishbb",
		BoardDescription: "A discussion board",
		RequiresApproval: false,
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
	var err error
	c.BoardName, err = GetConfigValue("board-name")
	if err != nil {
		return Config{}, err
	}
	c.BoardDescription, _ = GetConfigValue("board-description")
	r, _ := GetConfigValue("requires-approval")
	c.RequiresApproval, _ = strconv.ParseBool(r)
	return c, nil
}

func GetConfigValue(key string) (string, error) {
	row := db.QueryRow("select value from config where key = ?", key)
	var val string
	err := row.Scan(&val)
	if err != nil {
		return "", err
	}
	return val, nil
}

func UpdateConfig(key string, value any) error {
	_, err := db.Exec("insert into config(key,value) values(?1,?2) on conflict(key) do update set value = ?2", key, value)
	return err
}
