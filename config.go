package main

// most of these don't work yet
type Config struct {
	// Signups require admin approval
	RequiresApproval         bool
	BoardName                string
	BoardDescription         string
	RequireEmailVerification bool
	// private instances are only visible to logged in users
	Private string

	// smaller forum settings
	PageSize int

	// Internal Config (not exposed to forum admins for security reasons)
	// Directory where views (templates and static data) live.
	// include trailing slash
	ViewDir string
	// Path to backend database
	DBPath string
	// port to run the server on
	Port string
}

func NewConfig() Config {
	return Config{
		Port:             ":8080",
		BoardName:        "fishbb",
		BoardDescription: "A discussion board",
		PageSize:         5,
		ViewDir:          "./views/",
		DBPath:           "fishbb.db",
		RequiresApproval: true,
	}
}

func GetConfig() (Config, error) {
	return Config{}, nil
}

func SaveConfig() error {
	// update global config value as well
	return nil
}
