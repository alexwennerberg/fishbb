package main

import (
	"fishbb/login"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"text/template"
)

var views *template.Template

func serveHTML(w http.ResponseWriter, r *http.Request, name string, info map[string]any) {
	u := login.GetUserInfo(r)
	if u == nil && !devMode {
		w.Header().Set("Cache-control", "max-age=60")
	}
	info["User"] = u
	info["Config"] = config
	info["Version"] = softwareVersion
	var title = config.BoardName
	if name != "index" {
		title += " > " + name
	}
	info["Title"] = config.BoardName // TODO better
	err := views.ExecuteTemplate(w, name+".html", info)
	if err != nil {
		log.Error(err.Error())
		// TODO server error
	}
}

func indexPage(w http.ResponseWriter, r *http.Request) {
	tmpl := make(map[string]any)
	tmpl["Forums"] = getForums()
	serveHTML(w, r, "index", tmpl)
}

func userPage(w http.ResponseWriter, r *http.Request) {
	tmpl := make(map[string]any)
	uid, _ := strconv.Atoi(r.PathValue("id"))
	tmpl["InfoUser"] = getUser(uid)
	serveHTML(w, r, "user", tmpl)
}

func forumPage(w http.ResponseWriter, r *http.Request) {
	tmpl := make(map[string]any)
	limit := config.PageSize
	offset := 0
	tmpl["ForumID"] = getForumID(r.PathValue("forum"))
	tmpl["Threads"] = getThreads(1, limit, offset)
	serveHTML(w, r, "forum", tmpl)
}

func threadPage(w http.ResponseWriter, r *http.Request) {
	tmpl := make(map[string]any)
	threadID, _ := strconv.Atoi(r.PathValue("threadid"))
	var limit, offset = 0, 0 // TODO
	tmpl["Thread"] = getThread(threadID)
	tmpl["Posts"] = getPosts(threadID, limit, offset)
	serveHTML(w, r, "thread", tmpl)
}

func newThreadPage(w http.ResponseWriter, r *http.Request) {
	tmpl := make(map[string]any)
	forumID, _ := strconv.Atoi(r.URL.Query().Get("forumid"))
	tmpl["Forum"] = getForum(forumID)
	serveHTML(w, r, "new-thread", tmpl)
}

func newPostPage(w http.ResponseWriter, r *http.Request) {
	// ...
	tmpl := make(map[string]any)
	serveHTML(w, r, "new-post", tmpl)
}

func createNewPost(w http.ResponseWriter, r *http.Request) {
	u := login.GetUserInfo(r)
	content := r.FormValue("content")
	tid, _ := strconv.Atoi(r.URL.Query().Get("threadid"))
	_, err := createPost(u.UserID, int(tid), content)
	if err != nil {
		// handle
	}
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func createNewThread(w http.ResponseWriter, r *http.Request) {
	u := login.GetUserInfo(r)
	title := r.FormValue("title")
	content := r.FormValue("content")
	forumID, _ := strconv.Atoi(r.URL.Query().Get("forumid"))
	tid, err := createThread(u.UserID, forumID, title)
	if err != nil {
		// handle
	}
	_, err = createPost(u.UserID, int(tid), content)
	if err != nil {
		// handle
	}
	slug := getForum(forumID).Slug
	http.Redirect(w, r, fmt.Sprintf("/f/%s/%d", slug, tid), http.StatusSeeOther)
}

// hashes string and builds png avatar
func avatarHandler(w http.ResponseWriter, r *http.Request) {
	in := r.FormValue("a")
	w.Header().Set("Cache-Control", "max-age=604800")
	w.Write(genAvatar(in))
}

func loginPage(w http.ResponseWriter, r *http.Request) {
	tmpl := make(map[string]any)
	u := login.GetUserInfo(r)
	if u != nil {
		// Redirect home
	}
	serveHTML(w, r, "login", tmpl)
}

func registerPage(w http.ResponseWriter, r *http.Request) {
	tmpl := make(map[string]any)
	u := login.GetUserInfo(r)
	if u != nil {
		// Redirect home
	}
	serveHTML(w, r, "register", tmpl)
}
func serveAsset(w http.ResponseWriter, r *http.Request) {
	if !devMode {
		w.Header().Set("Cache-Control", "max-age=604800")
	}
	http.ServeFile(w, r, config.ViewDir+r.URL.Path)
}

func loadTemplates() *template.Template {
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
	views, _ := template.ParseFiles(toload...)
	views.Funcs(template.FuncMap{
		"timeago": timeago,
	})
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
	mux.HandleFunc("GET /f/{forum}", forumPage)
	mux.HandleFunc("GET /f/{forum}/{threadid}", threadPage)
	mux.HandleFunc("GET /user/{id}", userPage)
	mux.HandleFunc("GET /login", loginPage)
	mux.HandleFunc("GET /register", registerPage)
	mux.HandleFunc("GET /reset-password", dummy)
	mux.HandleFunc("GET /search", dummy)
	mux.HandleFunc("GET /style.css", serveAsset)
	mux.HandleFunc("GET /thread/new", newThreadPage)
	mux.HandleFunc("GET /a", avatarHandler)
	mux.HandleFunc("POST /thread/new", createNewThread)
	mux.HandleFunc("GET /post/new", newPostPage)
	mux.HandleFunc("POST /post/new", createNewPost)

	mux.HandleFunc("POST /dologin", login.LoginFunc)
	mux.HandleFunc("POST /logout", login.LogoutFunc)

	mux.HandleFunc("POST /post/{id}/delete", dummy)
	mux.HandleFunc("POST /post/{id}/edit", dummy)
	mux.HandleFunc("POST /thread/{id}/update-meta", dummy)
	mux.HandleFunc("POST /user/{id}/update", dummy)
	mux.HandleFunc("POST /user/{id}/reset-password", dummy)

	// admin functions
	mux.HandleFunc("POST /ban-user", dummy)
	mux.HandleFunc("POST /set-user-role", dummy)

	log.Debug("starting server")
	err := http.ListenAndServe(config.Port, logRequest(login.Checker(mux)))
	if err != nil {
		panic(err)
	}
}
