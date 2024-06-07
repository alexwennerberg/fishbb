package main

import (
	"fishbb/login"
	"humungus.tedunangst.com/r/webs/templates"
	"log"
	"net/http"
	"os"
	"strings"
)

var views *templates.Template

func serveHTML(w http.ResponseWriter, r *http.Request, name string, info map[string]any) {
	u := login.GetUserInfo(r)
	// TODO add other context?
	if u == nil && !devMode {
		w.Header().Set("Cache-control", "max-age=60")
	}
	info["User"] = u
	err := views.Execute(w, name, info)
	if err != nil {
		log.Printf("foo") // TODO logging story
		// TODO server error
	}
}

func indexPage(w http.ResponseWriter, r *http.Request) {
	tmpl := make(map[string]any)
	tmpl["Forums"] = []Forum{}
	views.Execute(w, "index.html", tmpl)
}

func mePage(w http.ResponseWriter, r *http.Request) {
	// if path == "me" special case
}

func forumPage(w http.ResponseWriter, r *http.Request) {
}

func threadPage(w http.ResponseWriter, r *http.Request) {
}

func loginPage(w http.ResponseWriter, r *http.Request) {
}

func serveAsset(w http.ResponseWriter, r *http.Request) {
	if !devMode {
		w.Header().Set("Cache-Control", "max-age=604800")
	}
	http.ServeFile(w, r, config.ViewDir+r.URL.Path)
}

func loadTemplates() *templates.Template {
	var toload []string
	viewDir := config.ViewDir
	temps, err := os.ReadDir(viewDir)
	if err != nil {
		panic(err)
	}
	for _, temp := range temps {
		name := temp.Name()
		if strings.HasSuffix(name, ".html") {
			toload = append(toload, viewDir+name)
		}
	}
	views := templates.Load(devMode, toload...)
	return views
}

func authMidddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		next.ServeHTTP(w, r)
	})
}

// placeholder
func dummy(w http.ResponseWriter, r *http.Request) {
}

func serve() {
	db = opendb()
	views = loadTemplates()
	prepareStatements(db)

	// Setup Templates
	mux := http.NewServeMux()
	mux.HandleFunc("/", indexPage)
	mux.HandleFunc("GET /{forumid}", forumPage)
	mux.HandleFunc("GET /{forumid}/{threadid}", threadPage)
	mux.HandleFunc("GET /user/{id}", mePage)
	mux.HandleFunc("GET /login", loginPage)
	mux.HandleFunc("GET /reset-password", dummy)
	mux.HandleFunc("GET /search", dummy)
	mux.HandleFunc("GET /style.css", serveAsset)

	// TODO rss

	mux.HandleFunc("POST /login", login.LoginFunc)
	mux.HandleFunc("POST /logout", login.LogoutFunc)

	mux.HandleFunc("POST /new-thread", dummy)
	mux.HandleFunc("POST /new-post", dummy)
	mux.HandleFunc("POST /delete-post", dummy)
	mux.HandleFunc("POST /edit-post", dummy)
	mux.HandleFunc("POST /update-thread-meta", dummy)
	mux.HandleFunc("POST /update-user", dummy)
	mux.HandleFunc("POST /reset-password", dummy)

	// admin functions
	mux.HandleFunc("POST /ban-user", dummy)
	mux.HandleFunc("POST /set-user-role", dummy)

	err := http.ListenAndServe(config.Port, login.Checker(mux))
	if err != nil {
		panic(err)
	}
}
