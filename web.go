package main

import (
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"text/template"
	"time"

	"fishbb/login"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	// I don't love this
	"github.com/go-chi/httprate"
	slogchi "github.com/samber/slog-chi"
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
	info["LogoutCSRF"] = login.GetCSRF("logout", r)
	var title = config.BoardName
	if name != "index" {
		title += " > " + name
	}
	info["Title"] = config.BoardName // TODO better
	err := views.ExecuteTemplate(w, name+".html", info)
	if err != nil {
		log.Error(err.Error())
		serverError(w,r)
	}
}

func errorPage(w http.ResponseWriter, r *http.Request, code int, message string) {
	tmpl := make(map[string]any)
	w.WriteHeader(code)
	tmpl["Error"] = strconv.Itoa(code) + " " + http.StatusText(code) + " " + message
	serveHTML(w, r, "error", tmpl)
}

func serverError(w http.ResponseWriter, r *http.Request) {
	errorPage(w,r,http.StatusInternalServerError, "")
}

func notFound(w http.ResponseWriter, r *http.Request) {
	errorPage(w,r,http.StatusNotFound,"")
}

func indexPage(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		notFound(w,r)
		return
	}
	tmpl := make(map[string]any)
	tmpl["Forums"] = getForums()
	serveHTML(w, r, "index", tmpl)
}

func forumPage(w http.ResponseWriter, r *http.Request) {
	tmpl := make(map[string]any)
	limit := config.PageSize
	offset := 0
	fid := getForumID(r.PathValue("forum"))
	tmpl["ForumID"] = fid
	tmpl["Threads"] = getThreads(fid, limit, offset)
	serveHTML(w, r, "forum", tmpl)
}

func threadPage(w http.ResponseWriter, r *http.Request) {
	tmpl := make(map[string]any)
	threadID, _ := strconv.Atoi(r.PathValue("threadid"))
	fid := getForumID(r.PathValue("forum"))
	forum := getForum(fid)
	var limit, offset = 0, 0 // TODO
	tmpl["Thread"] = getThread(threadID)
	tmpl["Forum"] = forum
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
	tmpl := make(map[string]any)
	tid, _ := strconv.Atoi(r.URL.Query().Get("threadid"))
	thread := getThread(tid)
	tmpl["Thread"] = thread
	tmpl["Forum"] = getForum(thread.ForumID)
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

// TODO Function: meOrCapability(foo) -> optimize later

func doDeletePost(w http.ResponseWriter, r *http.Request) {
	// u := login.GetUserInfo(r)
	// pid, _ := strconv.Atoi(r.URL.Query().Get("postid"))
	// // TODO build abstraction
	// aid = getPostAuthorID(pid)
	// if u.UserID == aid || can(getUser(iad).Capabilities, deletePosts) {

	// } else {
	// 	// Unauthorized
	// }
}

func editPostPage(w http.ResponseWriter, r *http.Request) {
}

func doEditPost(w http.ResponseWriter, r *http.Request) {
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
	serveHTML(w, r, "login", tmpl)
}

func registerPage(w http.ResponseWriter, r *http.Request) {
	tmpl := make(map[string]any)
	serveHTML(w, r, "register", tmpl)
}

func doRegister(w http.ResponseWriter, r *http.Request) {
	// TODO
	http.Redirect(w, r, "/", http.StatusSeeOther)
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
	views, err :=  template.New("main").Funcs(template.FuncMap{
		"timeago": timeago,
		"markup": markup,
	}).ParseFiles(toload...)
	if err != nil {
		panic(err)
	}
	return views
}

func userPage(w http.ResponseWriter, r *http.Request) {
	tmpl := make(map[string]any)

	uid, _ := strconv.Atoi(r.PathValue("userid"))
	// TODO err handling
	info, _ := getUser(uid)
	tmpl["InfoUser"] = info
	// TODO specific DNE error
	serveHTML(w, r, "user", tmpl)
}

func mePage(w http.ResponseWriter, r *http.Request) {
	tmpl := make(map[string]any)
	u := login.GetUserInfo(r)
	info, err := getUser(u.UserID)
	if err != nil {
		log.Error("error getting user", "error", err.Error())
		serverError(w,r)
		return
	}
	tmpl["UserInfo"] = info
	serveHTML(w, r, "me", tmpl)
}

func resetPasswordPage(w http.ResponseWriter, r *http.Request) {
	tmpl := make(map[string]any)
	serveHTML(w,r, "reset-password", tmpl)
}

func doUpdateMe(w http.ResponseWriter, r *http.Request) {
	u := login.GetUserInfo(r)
	about := r.FormValue("about")
	website := r.FormValue("website")
	if len(about) > 250 || len(website) > 250 {
		return // TODO err
	}
	err := updateMe(u.UserID, r.FormValue("about"), r.FormValue("website"))
	if err != nil {
		serverError(w,r)
		return 
	}
	tmpl := make(map[string]any)
	info, _ := getUser(u.UserID)
	tmpl["UserInfo"] = info	
	tmpl["Notice"] = "Updated!"
	serveHTML(w, r, "me", tmpl)
}

func doResetPassword(w http.ResponseWriter, r *http.Request) {
}


// placeholder
func dummy(w http.ResponseWriter, r *http.Request) {
}

func serve() {
	db = opendb()
	views = loadTemplates()
	prepareStatements(db)

	r := chi.NewRouter()
	r.Use(middleware.RealIP)
	r.Use(slogchi.New(log))
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(60 * time.Second))
	r.Use(login.Checker) // TODO -- maybe not every route?

	// Setup Templates
	r.HandleFunc("/", indexPage)
	r.HandleFunc("GET /f/{forum}", forumPage)
	r.HandleFunc("GET /f/{forum}/{threadid}", threadPage)
	r.HandleFunc("GET /user/{userid}", userPage)
	r.HandleFunc("GET /login", loginPage)
	r.HandleFunc("GET /register", registerPage)
	r.HandleFunc("GET /search", dummy)
	r.HandleFunc("GET /style.css", serveAsset)
	r.HandleFunc("GET /thread/new", newThreadPage)
	r.HandleFunc("GET /a", avatarHandler)
	r.HandleFunc("POST /thread/new", createNewThread)
	r.HandleFunc("GET /post/new", newPostPage)
	r.HandleFunc("POST /post/new", createNewPost)
	r.HandleFunc("GET /post/new", newPostPage)

	r.With(httprate.LimitByIP(10, 1 * time.Hour)).HandleFunc("POST /dologin", login.LoginFunc)
	r.HandleFunc("POST /logout", login.LogoutFunc)
	r.HandleFunc("POST /register", doRegister)

	r.Group(func(r chi.Router) {
		// TODO loggedin
		r.Use(login.Required)
		r.HandleFunc("GET /me", mePage)
		r.HandleFunc("POST /me", doUpdateMe)
		r.HandleFunc("POST /post/{postid}/delete", doDeletePost)
		r.HandleFunc("GET /reset-password", resetPasswordPage)
		r.HandleFunc("GET /post/{postid}/edit", editPostPage)
		r.HandleFunc("POST /post/{postid}/edit", doEditPost)
		r.HandleFunc("POST /thread/{threadid}/update-meta", dummy)
		r.HandleFunc("POST /user/{userid}/update", dummy)
		r.HandleFunc("POST /user/{userid}/reset-password", dummy)
	})

	// admin functions
	r.HandleFunc("POST /ban-user", dummy)
	r.HandleFunc("POST /set-user-role", dummy)

	log.Debug("starting server")
	err := http.ListenAndServe(config.Port, r)
	if err != nil {
		panic(err)
}
}
