package main

type Config struct {
	// allow signups without approval
	openSignups bool
	name string
	description string
	requireEmailVerification bool
	// private instances are only visible to logged in users
	private string

	// port to run the server on
	Port int

	// WARNING: you probably don't want this set to true! only do so if you're
	// on a small, secure network (ie, not the public internet) 
	allowAnonymousPosts bool
}

func NewConfig() {
	return Config {
		Port: 8080
	}
}
