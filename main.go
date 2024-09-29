package main

import (
	"context"
	"database/sql"
	"net/http"
	"strings"

	"git.sr.ht/~aw/fishbb/server"
	fishbb "git.sr.ht/~aw/fishbb/server"
	// "github.com/go-chi/chi"
)

const instanceSubdomain = ""

// TODO parameterize
const dbFolder = "dbs"

// multi-instance main
func main() {
	fishbb.Serve()
	// initDBs()
	// TODO ...
	// r := chi.NewRouter()
	// No Subdomain -> cluster routes
	// Has subdomain -> instance route
}

func indexPage(w http.ResponseWriter, r *http.Request) {
	// cache me
	// Iterate over DBs
	// execute select * query
}

func createInstancePage(w http.ResponseWriter, r *http.Request) {
	// render html page with form to create an instance
}

func createInstanceHandler() {
	// var instanceName string
	// var adminUser string
	// var adminPassword string
	// legal == alphanumeric and - (ascii domain)
	// check if file exists
	// make db
	// initialize admin user
}

// DB context middleware
func DBContext(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		subdomain, _, _ := strings.Cut(r.Host, ".")
		d, ok := databases[subdomain]
		if !ok {
			w.WriteHeader(404)
			w.Write([]byte("404 not found"))
			return
		}
		ctx = context.WithValue(ctx, "db", d)
		next.ServeHTTP(w, r.WithContext(ctx))
	}
	return http.HandlerFunc(fn)
}

var databases map[string]*sql.DB
var configs map[string]server.Config

func initDBs() {
	// items, _ := ioutil.ReadDir(dbFolder)
	// for _, _ := range items {
	// Open DB
	//...
	// fishbb.PrepareStatements(db)
	// }
}
