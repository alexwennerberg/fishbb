package main

// most of these don't work yet
type Config struct {
	// allow signups without approval
	OpenSignups              bool
	Name                     string
	Description              string
	RequireEmailVerification bool
	// private instances are only visible to logged in users
	Private string

	// WARNING: you probably don't want this set to true! only do so if you're
	// on a small, secure network (ie, not the public internet)
	AllowAnonymousPosts bool

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
		Port:     ":8080",
		PageSize: 100,
		ViewDir:  "./views/",
		DBPath:   "fishbb.db",
	}
}
