package main

import (
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	// I don't love this
	"github.com/go-chi/httplog/v2"
)

func serveHTML(w http.ResponseWriter, r *http.Request, name string, info map[string]any) {
	l := httplog.LogEntry(r.Context())
	u := GetUserInfo(r)
	var user *User
	var err error
	if u != nil {
		user, err = getUser(u.Username)
		if err != nil {
			*l = *l.With(httplog.ErrAttr(err))
			w.Write([]byte("<h1>TEMPLATE ERROR</h1>"))
		}
	} else if u == nil {
		// TODO -- caching broken
		// w.Header().Set("Cache-control", "max-age=60")
	}
	info["User"] = user
	info["Config"] = config
	info["Version"] = SoftwareVersion
	info["CSRFToken"] = GetCSRF(r)

	// Extract board from path
	if info["Board"] == nil {
		pathParts := strings.Split(strings.TrimPrefix(r.URL.Path, "/"), "/")
		if len(pathParts) > 0 && pathParts[0] != "" {
			board, err := getBoard(pathParts[0])
			if err == nil {
				info["Board"] = board
			}
		}
	}

	var title = config.BoardName
	if info["Subtitle"] != nil {
		title = info["Subtitle"].(string) + " - " + title
	} else if name != "index" {
		title = name + " - " + title
	}
	info["Title"] = title
	err = views.ExecuteTemplate(w, name+".html", info)
	if err != nil {
		*l = *l.With(httplog.ErrAttr(err))
		w.Write([]byte("<h1>Error Rending Template</h1>"))
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
	errorPage(w, r, http.StatusInternalServerError, "")
}

func tooManyRequests(w http.ResponseWriter, r *http.Request) {
	errorPage(w, r, http.StatusTooManyRequests, "Slow down!")
}

func notFound(w http.ResponseWriter, r *http.Request) {
	errorPage(w, r, http.StatusNotFound, "")
}

func unauthorized(w http.ResponseWriter, r *http.Request) {
	errorPage(w, r, http.StatusUnauthorized, "You are not authorized to perform this action, sorry!")
}

// new index will be a list of instances
func indexPage(w http.ResponseWriter, r *http.Request) {
	tmpl := make(map[string]any)
	i, err := getBoards()
	if err != nil {
		serverError(w, r, err)
	}
	tmpl["Forums"] = i
	serveHTML(w, r, "index", tmpl)
}

func boardIndex(w http.ResponseWriter, r *http.Request) {
	u := GetUserInfo(r)
	var role Role
	if u != nil {
		role = u.Role
	}
	tmpl := make(map[string]any)
	f, err := getForums()
	if err != nil {
		serverError(w, r, err)
		return
	}
	// filter forums above your role
	// TODO move to db layer
	var forums []Forum
	for _, ff := range f {
		if role.Can(ff.ReadPermissions) {
			forums = append(forums, ff)
		}
	}
	tmpl["Forums"] = forums
	serveHTML(w, r, "board", tmpl)
}

func forumPage(w http.ResponseWriter, r *http.Request) {
	tmpl := make(map[string]any)
	f, err := getForumBySlug(r.PathValue("forum"))
	if errors.Is(err, sql.ErrNoRows) {
		notFound(w, r)
		return
	} else if err != nil {
		serverError(w, r, err)
		return
	}
	page := page(r)
	threads, err := getThreads(f.ID, page)
	if err != nil {
		serverError(w, r, err)
		return
	}
	count, err := getThreadCount(f.ID)
	if err != nil {
		serverError(w, r, err)
		return
	}
	tmpl["Forum"] = f
	tmpl["Threads"] = threads
	// pagination
	tmpl["Page"] = page
	tmpl["ItemCount"] = count
	tmpl["Threads"] = threads
	tmpl["Subtitle"] = f.Name
	serveHTML(w, r, "forum", tmpl)
}

func threadPage(w http.ResponseWriter, r *http.Request) {
	tmpl := make(map[string]any)
	threadID, err := strconv.Atoi(r.PathValue("threadid"))
	if err != nil {
		notFound(w, r)
		return
	}

	forum, err := getForumBySlug(r.PathValue("forum"))
	if err != nil {
		serverError(w, r, err)
	}
	page := page(r)
	thread, err := getThread(threadID)
	// TODO -- doesnt work?
	if errors.Is(err, sql.ErrNoRows) {
		notFound(w, r)
		return
	} else if err != nil {
		serverError(w, r, err)
		return
	}
	tmpl["Thread"] = thread
	// pagination
	tmpl["Page"] = page
	tmpl["ItemCount"] = thread.Replies + 1
	tmpl["Forum"] = forum
	tmpl["Posts"] = getPosts(threadID, page)
	tmpl["Subtitle"] = thread.Title
	serveHTML(w, r, "thread", tmpl)
}

func newThreadPage(w http.ResponseWriter, r *http.Request) {
	tmpl := make(map[string]any)
	forumID, _ := strconv.Atoi(r.URL.Query().Get("forumid"))
	var err error
	forum, err := getForum(forumID)
	if err != nil {
		serverError(w, r, err)
		return
	}
	tmpl["Forum"] = forum
	tmpl["Board"] = forum.Board
	serveHTML(w, r, "new-thread", tmpl)
}

func createForumPage(w http.ResponseWriter, r *http.Request) {
	tmpl := make(map[string]any)
	serveHTML(w, r, "new-forum", tmpl)
}

func newPostPage(w http.ResponseWriter, r *http.Request) {
	tmpl := make(map[string]any)
	tid, _ := strconv.Atoi(r.URL.Query().Get("thread"))
	inReplyTo, _ := strconv.Atoi(r.URL.Query().Get("reply"))

	thread, err := getThread(tid)
	if err != nil {
		serverError(w, r, err)
		return
	}
	if inReplyTo != 0 {
		post, err := getPost(inReplyTo)
		if err == nil {
			tmpl["Content"] = post.BuildReply()
		}
	} else {
		tmpl["Content"] = ""
	}
	tmpl["Thread"] = thread
	forum, err := getForum(thread.ForumID)
	if err != nil {
		serverError(w, r, err)
		return
	}
	tmpl["Forum"] = forum
	tmpl["Board"] = forum.Board
	serveHTML(w, r, "new-post", tmpl)
}

func createNewPost(w http.ResponseWriter, r *http.Request) {
	u := GetUserInfo(r)
	content := r.FormValue("content")
	if !postValid(content) {
		return // TODO 4xx
	}
	tid, _ := strconv.Atoi(r.URL.Query().Get("thread"))
	thread, _ := getThread(tid)
	// TODO check if forum allows you to post
	if !u.Role.Can(RoleUser) || (thread.Locked && !u.Role.Can(RoleMod)) {
		unauthorized(w, r)
		return
	}
	pid, err := createPost(u.UserID, int(tid), content)
	if err != nil {
		serverError(w, r, err)
	}
	slug, err := getPostSlug(int(pid))
	if err != nil {
		serverError(w, r, err)
	}
	http.Redirect(w, r, slug, http.StatusSeeOther)
}

func doDeletePost(w http.ResponseWriter, r *http.Request) {
	u := GetUserInfo(r)
	pid, err := strconv.Atoi(r.PathValue("postid"))
	if err != nil {
		notFound(w, r)
		return
	}
	// TODO build abstraction around post controlling?
	post, err := getPost(pid)
	if err != nil {
		serverError(w, r, err)
		return
	}
	aid := post.Author.ID
	if u.UserID != aid && !u.Role.Can(RoleMod) {
		unauthorized(w, r)
		return
	}

	// Get thread and forum info for redirect
	var threadID int
	var forumID int
	row := db.QueryRow("select threadid, thread.forumid from post join thread on post.threadid = thread.id where post.id = ?", pid)
	err = row.Scan(&threadID, &forumID)
	if err != nil {
		serverError(w, r, err)
		return
	}

	err = deletePost(pid)
	if err != nil {
		serverError(w, r, err)
		return
	}

	forum, err := getForum(forumID)
	if err != nil {
		serverError(w, r, err)
		return
	}

	http.Redirect(w, r, fmt.Sprintf("%s/%d", forum.Slug, threadID), http.StatusSeeOther)
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
	u := GetUserInfo(r)
	if post.Author.ID != u.UserID && !u.Role.Can(RoleMod) {
		unauthorized(w, r)
		return
	}
	if r.Method == "POST" {
		content := r.FormValue("content")
		if !postValid(content) {
			// TODO
		}
		err = editPost(pid, content)
		if err != nil {
			serverError(w, r, err)
			return
		}
		post.Content = content
		slug, err := getPostSlug(pid)
		if err != nil {
			serverError(w, r, err)
			return
		}
		http.Redirect(w, r, slug, http.StatusSeeOther)
	}
	tmpl["Post"] = post
	serveHTML(w, r, "edit-post", tmpl)
}

func createNewThread(w http.ResponseWriter, r *http.Request) {
	u := GetUserInfo(r)
	forumID, _ := strconv.Atoi(r.URL.Query().Get("forumid"))
	f, err := getForum(forumID)
	if err != nil {
		serverError(w, r, err)
	}
	if !u.Role.Can(f.WritePermissions) {
		unauthorized(w, r)
		return
	}
	title := r.FormValue("title")
	content := r.FormValue("content")
	tid, err := createThread(u.UserID, forumID, title)
	if err != nil {
		// handle
		serverError(w, r, err)
	}
	_, err = createPost(u.UserID, int(tid), content)
	if err != nil {
		serverError(w, r, err)
		// handle
	}
	http.Redirect(w, r, fmt.Sprintf("%s/%d", f.Slug, tid), http.StatusSeeOther)
}

// hashes string and builds png avatar
func avatarHandler(w http.ResponseWriter, r *http.Request) {
	in := r.FormValue("a")
	w.Header().Set("Cache-Control", "max-age=604800")
	w.Write(genAvatar(in))
}

func loginPage(w http.ResponseWriter, r *http.Request) {
	tmpl := make(map[string]any)
	tmpl["LoginErr"], _ = GetFlash(w, r, "login-err")
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
		if len(password) < 8 {
			formErr("Passwords must be at least 8 characters.")
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

		role := RoleUser
		if config.RequiresApproval {
			role = RoleInactive
		}
		err := createUser(username, email, password, role)
		if err != nil {
			serverError(w, r, err)
		}
		LoginFunc(w, r)
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}
	serveHTML(w, r, "register", tmpl)
}

func serveAsset(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "max-age=604800")
	http.ServeFileFS(w, r, viewBundle, "views"+r.URL.Path)
}
func userPage(w http.ResponseWriter, r *http.Request) {
	tmpl := make(map[string]any)

	uname := r.PathValue("username")
	info, err := getUser(uname)
	if err != nil {
		serverError(w, r, err)
		return
	}
	if info == nil {
		notFound(w, r)
		return
	}
	posts, err := getPostsByUser(info.ID, page(r))
	if err != nil {
		serverError(w, r, err)
		return
	}
	tmpl["InfoUser"] = info
	tmpl["Posts"] = posts
	tmpl["Subtitle"] = info.Username
	// TODO specific DNE error?
	serveHTML(w, r, "user", tmpl)
}

func mePage(w http.ResponseWriter, r *http.Request) {
	tmpl := make(map[string]any)
	u := GetUserInfo(r)
	info, err := getUser(u.Username)
	if err != nil {
		err = fmt.Errorf("failed to get user %d: %w", u.UserID, err)
		serverError(w, r, err)
		return
	}
	tmpl["UserInfo"] = info
	serveHTML(w, r, "me", tmpl)
}

func controlPanelPage(w http.ResponseWriter, r *http.Request) {
	tmpl := make(map[string]any)
	var err error
	tmpl["Users"], err = getUsers()
	if err != nil {
		err = fmt.Errorf("failed to get users: %w", err)
		serverError(w, r, err)
		return
	}
	tmpl["Forums"], err = getForums()
	if err != nil {
		err = fmt.Errorf("failed to get forums: %w", err)
		serverError(w, r, err)
		return
	}
	serveHTML(w, r, "control", tmpl)
}

func doUpdateConfig(w http.ResponseWriter, r *http.Request) {
	fields := []string{"board-description", "requires-approval"}
	var err error
	for _, key := range fields {
		val := r.PostFormValue(key)
		// hack
		if val == "on" {
			err = UpdateConfig(key, true)
		} else {
			err = UpdateConfig(key, val)
		}
		if err != nil {
			serverError(w, r, err)
			return
		}
		config, err = GetConfig()
		if err != nil {
			serverError(w, r, err)
			return
		}
	}
	http.Redirect(w, r, "/control", http.StatusSeeOther)
}

func changePasswordPage(w http.ResponseWriter, r *http.Request) {
	tmpl := make(map[string]any)
	serveHTML(w, r, "change-password", tmpl)
}

func doChangePassword(w http.ResponseWriter, r *http.Request) {
	u := GetUserInfo(r)
	password := r.FormValue("password")
	password2 := r.FormValue("password2")
	if password != password2 {
		// TODO form validation
		return
	}
	if len(password) < 8 {
		return
	}
	err := updatePassword(u.UserID, password)
	if err != nil {
		serverError(w, r, err)
		return
	}
	http.Redirect(w, r, "/me", http.StatusSeeOther)
}

func doUpdateMe(w http.ResponseWriter, r *http.Request) {
	u := GetUserInfo(r)
	about := r.FormValue("about")
	website := r.FormValue("website")
	username := r.FormValue("username")
	if len(about) > 250 || len(website) > 250 || !validUsername(username) {
		return // TODO validation err
	}
	updateUser := User{
		ID:          u.UserID,
		Username:    username,
		Email:       r.FormValue("email"),
		EmailPublic: r.FormValue("email-public") == "on",
		About:       about,
		Website:     website,
	}
	err := updateMe(updateUser)
	if err != nil {
		serverError(w, r, err)
		return
	}
	tmpl := make(map[string]any)
	info, _ := getUser(u.Username)
	tmpl["UserInfo"] = info
	tmpl["Notice"] = "Updated!"
	serveHTML(w, r, "me", tmpl)
}

func searchPage(w http.ResponseWriter, r *http.Request) {
	tmpl := make(map[string]any)
	q := r.URL.Query().Get("q")
	tmpl["Query"] = q
	posts, err := searchPosts(q)
	if err != nil {
		serverError(w, r, err)
		return
	}
	tmpl["Posts"] = posts
	serveHTML(w, r, "search", tmpl)
}

func notificationsPage(w http.ResponseWriter, r *http.Request) {
	u := GetUserInfo(r)
	tmpl := make(map[string]any)
	q := fmt.Sprintf("@%s", u.Username)
	tmpl["Query"] = q
	posts, err := searchPosts(q)
	if err != nil {
		serverError(w, r, err)
		return
	}
	err = setNotificationsRead(u.UserID)
	if err != nil {
		serverError(w, r, err)
		return
	}
	tmpl["Posts"] = posts
	serveHTML(w, r, "search", tmpl)
}

func doSetRole(w http.ResponseWriter, r *http.Request) {
	action := r.FormValue("role")
	uid, err := strconv.Atoi(r.PathValue("uid"))
	if err != nil {
		serverError(w, r, err)
		return
	}
	// TODO must be valid role
	err = updateUserRole(uid, Role(action))
	if err != nil {
		serverError(w, r, err)
		return
	}
	http.Redirect(w, r, "/control", http.StatusSeeOther)
}

func editForumPage(w http.ResponseWriter, r *http.Request) {
	tmpl := make(map[string]any)
	var err error
	fs := r.PathValue("forum")
	if err != nil {
		serverError(w, r, err)
		return
	}
	id := getForumID(fs)
	if r.Method == "POST" {
		// TODO field validation
		err := updateForum(id, r.FormValue("description"), Role(r.FormValue("read-permissions")), Role(r.FormValue("write-permissions")))
		if err != nil {
			serverError(w, r, err)
		}
		http.Redirect(w, r, "/control", http.StatusSeeOther)
	}
	tmpl["Forum"], err = getForumBySlug(fs)
	serveHTML(w, r, "edit-forum", tmpl)
}

func doCreateForum(w http.ResponseWriter, r *http.Request) {
	name := r.FormValue("name")
	err := createForum(name, "Default Description", 1)
	if err != nil {
		serverError(w, r, err)
		return
	}

	f, err := getForumBySlug(slugify(name))
	if errors.Is(err, sql.ErrNoRows) {
		notFound(w, r)
		return
	} else if err != nil {
		serverError(w, r, err)
		return
	}

	http.Redirect(w, r, f.Slug, http.StatusSeeOther)
}

func doLockThread(w http.ResponseWriter, r *http.Request) {
	threadID, err := strconv.Atoi(r.PathValue("tid"))
	if err != nil {
		serverError(w, r, err)
		return
	}
	state, _ := strconv.ParseBool(r.URL.Query().Get("s"))

	thread, err := getThread(threadID)
	if err != nil {
		serverError(w, r, err)
		return
	}
	forum, err := getForum(thread.ForumID)
	if err != nil {
		serverError(w, r, err)
		return
	}

	err = setThreadLock(threadID, state)
	if err != nil {
		serverError(w, r, err)
		return
	}

	http.Redirect(w, r, forum.Slug, http.StatusSeeOther)
}

func doPinThread(w http.ResponseWriter, r *http.Request) {
	threadID, err := strconv.Atoi(r.PathValue("tid"))
	if err != nil {
		serverError(w, r, err)
		return
	}
	state, _ := strconv.ParseBool(r.URL.Query().Get("s"))

	thread, err := getThread(threadID)
	if err != nil {
		serverError(w, r, err)
		return
	}
	forum, err := getForum(thread.ForumID)
	if err != nil {
		serverError(w, r, err)
		return
	}

	err = setThreadPin(threadID, state)
	if err != nil {
		serverError(w, r, err)
		return
	}

	http.Redirect(w, r, forum.Slug, http.StatusSeeOther)
}

// placeholder
func dummy(w http.ResponseWriter, r *http.Request) {
}

func Serve() {
	// order is important here
	db = opendb()
	var err error
	config, err = GetConfig()
	if err != nil {
		panic(err)
	}
	views = loadTemplates()

	logger := httplog.NewLogger("fishbb", httplog.Options{
		LogLevel:        slog.LevelDebug,
		Concise:         true,
		RequestHeaders:  false,
		ResponseHeaders: false,
	})
	logger.Logger = log
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(httplog.RequestLogger(logger)) // TODO look into other logger
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(60 * time.Second))
	r.Use(Checker) // TODO -- maybe not every route?
	// This allows us to use non-POST methods in our routes
	r.Use(ChangePostToHiddenMethod)

	// Setup Templates
	r.HandleFunc("/", indexPage)
	r.HandleFunc("/{board}", boardIndex)
	// TODO these paths are kinda ugly
	r.HandleFunc("GET /{board}/f/{forum}", forumPage)
	r.HandleFunc("GET /{board}/f/{forum}/{threadid}", threadPage)
	r.HandleFunc("GET /u/{username}", userPage)
	r.HandleFunc("GET /login", loginPage)
	// TODO limit registration successes
	// TODO Consider CSRF wrapping
	r.With(LimitByRealIP(25, 1*time.Hour)).HandleFunc("/register", registerPage)
	r.HandleFunc("GET /search", searchPage)
	r.HandleFunc("GET /notifications", notificationsPage)
	r.HandleFunc("GET /style.css", serveAsset)
	r.HandleFunc("GET /robots.txt", serveAsset)
	r.HandleFunc("GET /fixi.js", serveAsset)
	r.HandleFunc("GET /favicon.ico", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "max-age=604800")
		w.Header().Set("Content-Type", "image/svg+xml")
		svg := `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 100 100"><text y="75" font-size="80">🐟</text></svg>`
		w.Write([]byte(svg))
	})
	r.HandleFunc("GET /a", avatarHandler)
	// TODO Consider CSRF wrapping
	r.With(LimitByRealIP(20, 1*time.Hour)).HandleFunc("POST /dologin", LoginFunc)
	r.HandleFunc("POST /logout", LogoutFunc)

	r.Group(func(r chi.Router) {
		r.Use(Required)

		r.HandleFunc("GET /post/new", newPostPage)
		r.With(CSRFWrap).With(LimitByUser(10, 5*time.Minute)).HandleFunc("POST /post/new", createNewPost)
		r.HandleFunc("GET /thread/new", newThreadPage)
		r.With(CSRFWrap).With(LimitByUser(10, 5*time.Minute)).HandleFunc("POST /thread/new", createNewThread)

		r.HandleFunc("GET /post/{postid}/edit", editPostPage)
		r.With(CSRFWrap).HandleFunc("POST /post/{postid}/edit", editPostPage)
		r.With(CSRFWrap).HandleFunc("DELETE /post/{postid}", doDeletePost)
		r.HandleFunc("GET /me", mePage)
		// TODO POST -> PUT /user/{id} unify user updates?
		r.With(CSRFWrap).HandleFunc("POST /me", doUpdateMe)
		r.HandleFunc("GET /change-password", changePasswordPage)
		r.HandleFunc("POST /change-password", doChangePassword)

		r.With(CSRFWrap).HandleFunc("POST /thread/{threadid}/update-meta", dummy)
		r.HandleFunc("POST /user/{userid}/change-password", dummy)
		r.HandleFunc("GET /forum/new", createForumPage)
		r.HandleFunc("POST /forum/new", doCreateForum)
	})

	r.Group(func(r chi.Router) {
		r.Use(Mod)
		r.HandleFunc("/control", controlPanelPage)
		r.HandleFunc("POST /thread/{tid}/lock", doLockThread)
		r.HandleFunc("POST /thread/{tid}/pin", doPinThread)
	})
	// admin functions
	r.Group(func(r chi.Router) {
		r.Use(Admin)
		r.HandleFunc("POST /update-config", doUpdateConfig)
		r.HandleFunc("POST /user/{uid}/set-role", doSetRole)
	})

	r.HandleFunc("/*", notFound)

	log.Info(fmt.Sprintf("starting server on port %s", Port))
	err = http.ListenAndServe(Port, r)
	if err != nil {
		panic(err)
	}
}
