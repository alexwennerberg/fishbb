package main

import (
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"strconv"
	"strings"
	"text/template"
	"time"

	"fishbb/login"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/httprate"

	// I don't love this
	"github.com/go-chi/httplog/v2"
)

var views *template.Template

func serveHTML(w http.ResponseWriter, r *http.Request, name string, info map[string]any) {
	l := httplog.LogEntry(r.Context())
	u := login.GetUserInfo(r)
	if u == nil && !devMode {
		w.Header().Set("Cache-control", "max-age=60")
	}
	info["User"] = u
	info["Config"] = config
	info["Version"] = softwareVersion
	info["CSRFToken"] = login.GetCSRF(r)
	var title = config.BoardName
	if name != "index" {
		title += " > " + name
	}
	info["Title"] = title
	err := views.ExecuteTemplate(w, name+".html", info)
	if err != nil {
		*l = *l.With(httplog.ErrAttr(err))
		w.Write([]byte("<h1>TEMPLATE ERROR</h1>"))
	}
}

func errorPage(w http.ResponseWriter, r *http.Request, code int, message string) {
	tmpl := make(map[string]any)
	w.WriteHeader(code)
	tmpl["Error"] = strconv.Itoa(code) + " " + http.StatusText(code) + " " + message
	serveHTML(w, r, "error", tmpl)
}

func serverError(w http.ResponseWriter, r *http.Request, err error) {
	l := httplog.LogEntry(r.Context())
	*l = *l.With(httplog.ErrAttr(err))
	errorPage(w,r,http.StatusInternalServerError, "")
}

func notFound(w http.ResponseWriter, r *http.Request) {
	errorPage(w,r,http.StatusNotFound,"")
}

func unauthorized(w http.ResponseWriter, r *http.Request) {
	errorPage(w,r,http.StatusUnauthorized,"You are not authorized to perform this action, sorry!")
}

func indexPage(w http.ResponseWriter, r *http.Request) {
	tmpl := make(map[string]any)
	tmpl["Forums"] = getForums()
	serveHTML(w, r, "index", tmpl)
}

func forumPage(w http.ResponseWriter, r *http.Request) {
	tmpl := make(map[string]any)
	fid := getForumID(r.PathValue("forum"))
	page := page(r)
	threads, err := getThreads(fid, page) 
	if err != nil {
		serverError(w,r,err)
	}
	count, err := getThreadCount(fid)
	if err != nil {
		serverError(w,r,err)
	}	
	tmpl["ForumID"] = fid
	tmpl["Threads"] = threads
	// pagination
	tmpl["Page"] = page
	tmpl["ItemCount"] = count
	tmpl["Threads"] = threads
	serveHTML(w, r, "forum", tmpl)
}

func threadPage(w http.ResponseWriter, r *http.Request) {
	tmpl := make(map[string]any)
	threadID, err := strconv.Atoi(r.PathValue("threadid"))
	if err != nil {
		notFound(w,r)
		return
	}
	fid := getForumID(r.PathValue("forum"))
	forum := getForum(fid)
	page := page(r)
	thread, err := getThread(threadID)
	if err != nil {
		serverError(w,r,err)
		return
	}
	tmpl["Thread"] = thread
	// pagination
	tmpl["Page"] = page
	tmpl["ItemCount"] = thread.Replies + 1
	tmpl["Forum"] = forum
	tmpl["Posts"] = getPosts(threadID, page)
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
	thread, err := getThread(tid)
	if err != nil {
		serverError(w,r,err)
		return
	}
	tmpl["Thread"] = thread
	tmpl["Forum"] = getForum(thread.ForumID)
	serveHTML(w, r, "new-post", tmpl)
}

func createNewPost(w http.ResponseWriter, r *http.Request) {
	u := login.GetUserInfo(r)
	content := r.FormValue("content")
	if !postValid(content) {
		return // TODO 4xx
	}
	tid, _ := strconv.Atoi(r.URL.Query().Get("threadid"))
	pid, err := createPost(u.UserID, int(tid), content)
	if err != nil {
		serverError(w,r,err)
	}
	slug, err := getPostSlug(int(pid))
	if err != nil {
		serverError(w,r,err)
	}
	http.Redirect(w, r, slug, http.StatusSeeOther)
}

func doDeletePost(w http.ResponseWriter, r *http.Request) {
	u := login.GetUserInfo(r)
	pid, err := strconv.Atoi(r.PathValue("postid"))
	if err != nil {
		notFound(w,r)
		return
	}
	// TODO build abstraction around post controlling?
	post, err := getPost(pid)
	if err != nil {
		serverError(w,r,err)
		return
	}
	aid := post.Author.ID
	if u.UserID != aid {
		unauthorized(w,r)
		return
	}
	err = deletePost(pid)
	if err != nil {
		serverError(w,r,err)
	}
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func editPostPage(w http.ResponseWriter, r *http.Request) {
	tmpl := make(map[string]any)
	pid, _ := strconv.Atoi(r.PathValue("postid"))
	post, err := getPost(pid)
	if err != nil {
		panic(err)
		// TODO distinguish notfound
		return
	}
	u := login.GetUserInfo(r)
	if post.Author.ID != u.UserID {
		unauthorized(w,r)
		return 
	}
	if r.Method == "POST" {
		content := r.FormValue("content")
		if !postValid(content) {
			// TODO
		}
		err = editPost(pid, content)
		if err != nil {
			serverError(w,r,err)
			return
		}
		post.Content = content
		// TODO redirect to post
	}
	tmpl["Post"] = post
	serveHTML(w, r, "edit-post", tmpl)
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
	if r.Method == "POST" {
		username := r.FormValue("username")
		email := r.FormValue("email")
		password := r.FormValue("password")
		password2 := r.FormValue("password2")

		formErr := func(s string) {
			w.WriteHeader(http.StatusBadRequest)
			tmpl["Flash"] = s
			serveHTML(w, r, "register", tmpl)
		}

		if password != password2 {
			formErr("Passwords do not match.")
			return
		}
		if !validUsername(username) {
			formErr("Invalid username. Must contain only letters and numbers and be maximum of 25 characters.")
			return
		}
		if !validEmail(email) {
			formErr("Invalid email.")
			return
		}

		err := createUser(username, email, password, RoleUser)
		if err != nil {
			serverError(w,r,err)
		}
		// log in with new account
		login.LoginFunc(w,r)
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
	views, err :=  template.New("main").Funcs(template.FuncMap{
		"timeago": timeago,
		"pageArr": pageArray,
		"markup": markup,
		"inc": func(i int) int {
            return i + 1
		
		},
		"dec": func(i int) int {
			return i - 1
		},
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
		serverError(w,r,err)
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
		serverError(w,r,err)
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

	logger := httplog.NewLogger("fishbb", httplog.Options{
		LogLevel: slog.LevelDebug,
		Concise: true,
		RequestHeaders: false,
		ResponseHeaders: false,
	})
	logger.Logger = log
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(httplog.RequestLogger(logger)) // TODO look into other logger
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(60 * time.Second))
	r.Use(login.Checker) // TODO -- maybe not every route?

	// Setup Templates
	r.HandleFunc("/", indexPage)
	r.HandleFunc("GET /f/{forum}", forumPage)
	r.HandleFunc("GET /f/{forum}/{threadid}", threadPage)
	r.HandleFunc("GET /user/{userid}", userPage)
	r.HandleFunc("GET /login", loginPage)
	// TODO limit registration successes
	// TODO csrf wrap
	r.HandleFunc("/register", registerPage)
	r.HandleFunc("GET /search", dummy)
	r.HandleFunc("GET /style.css", serveAsset)
	r.HandleFunc("GET /a", avatarHandler)
	r.With(httprate.LimitByIP(10, 1 * time.Hour)).HandleFunc("POST /dologin", login.LoginFunc)
	r.HandleFunc("POST /logout", login.LogoutFunc)

	r.Group(func(r chi.Router) {
		r.Use(login.Required)

		r.HandleFunc("GET /post/new", newPostPage)
		r.HandleFunc("POST /post/new", createNewPost)
		r.HandleFunc("GET /thread/new", newThreadPage)
		r.With(login.CSRFWrap).HandleFunc("POST /thread/new", createNewThread)
		r.HandleFunc("GET /me", mePage)
		r.With(login.CSRFWrap).HandleFunc("POST /me", doUpdateMe)
		r.With(login.CSRFWrap).HandleFunc("POST /post/{postid}/delete", doDeletePost)
		r.HandleFunc("GET /reset-password", resetPasswordPage)
		r.HandleFunc("GET /post/{postid}/edit", editPostPage)
		r.With(login.CSRFWrap).HandleFunc("POST /post/{postid}/edit", editPostPage)
		r.With(login.CSRFWrap).HandleFunc("POST /thread/{threadid}/update-meta", dummy)
		r.HandleFunc("POST /user/{userid}/reset-password", dummy)
	})

	// admin functions
	r.HandleFunc("POST /ban-user", dummy)
	r.HandleFunc("POST /set-user-role", dummy)

	r.HandleFunc("/*", notFound)

	log.Debug("starting server")
	err := http.ListenAndServe(config.Port, r)
	if err != nil {
		panic(err)
}
}
