// Copyright (c) 2019 Ted Unangst <tedu@tedunangst.com>
//
// Permission to use, copy, modify, and distribute this software for any
// purpose with or without fee is hereby granted, provided that the above
// copyright notice and this permission notice appear in all copies.
//
// THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES
// WITH REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF
// MERCHANTABILITY AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR
// ANY SPECIAL, DIRECT, INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES
// WHATSOEVER RESULTING FROM LOSS OF USE, DATA OR PROFITS, WHETHER IN AN
// ACTION OF CONTRACT, NEGLIGENCE OR OTHER TORTIOUS ACTION, ARISING OUT OF
// OR IN CONNECTION WITH THE USE OR PERFORMANCE OF THIS SOFTWARE.

// Simple cookie and password based logins.
// See Init for required schema.
package main

import (
	"context"
	"crypto/rand"
	"crypto/sha512"
	"crypto/subtle"
	"database/sql"
	"fmt"
	"hash"
	"io"
	"net/http"
	"regexp"
	"slices"
	"strings"
	"time"

	"github.com/alexedwards/argon2id"

	// Dependencies for cache and gate
	"errors"
	mathrand "math/rand"
	"reflect"
	"sync"
)

// Begin gate package
// Copyright (c) 2019 Ted Unangst <tedu@tedunangst.com>
//
// Permission to use, copy, modify, and distribute this software for any
// purpose with or without fee is hereby granted, provided that the above
// copyright notice and this permission notice appear in all copies.
//
// THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES
// WITH REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF
// MERCHANTABILITY AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR
// ANY SPECIAL, DIRECT, INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES
// WHATSOEVER RESULTING FROM LOSS OF USE, DATA OR PROFITS, WHETHER IN AN
// ACTION OF CONTRACT, NEGLIGENCE OR OTHER TORTIOUS ACTION, ARISING OUT OF
// OR IN CONNECTION WITH THE USE OR PERFORMANCE OF THIS SOFTWARE.
// The gate package provides rate limiters and serializers.

func init() {
	mathrand.Seed(time.Now().Unix())
}

// Limiter limits the number of concurrent outstanding operations.
// Typical usage: limiter.Start(); defer limiter.Finish()
type Limiter struct {
	maxout  int
	numout  int
	waiting int
	lock    sync.Mutex
	bell    *sync.Cond
	busy    map[interface{}]bool
}

// Create a new Limiter with maxout operations
func NewLimiter(maxout int) *Limiter {
	l := new(Limiter)
	l.maxout = maxout
	l.bell = sync.NewCond(&l.lock)
	l.busy = make(map[interface{}]bool)
	return l
}

// Wait for an opening, then return when ready.
func (l *Limiter) Start() {
	l.lock.Lock()
	for l.numout >= l.maxout {
		l.waiting++
		l.bell.Wait()
		l.waiting--
	}
	l.numout++
	l.lock.Unlock()
}

// Wait for an opening, then return when ready.
func (l *Limiter) StartKey(key interface{}) {
	l.lock.Lock()
	for l.numout >= l.maxout || l.busy[key] {
		l.waiting++
		l.bell.Wait()
		l.waiting--
	}
	l.busy[key] = true
	l.numout++
	l.lock.Unlock()
}

// Free an opening after finishing.
func (l *Limiter) Finish() {
	l.lock.Lock()
	l.numout--
	l.bell.Broadcast()
	l.lock.Unlock()
}

// Free an opening after finishing.
func (l *Limiter) FinishKey(key interface{}) {
	l.lock.Lock()
	delete(l.busy, key)
	l.numout--
	l.bell.Broadcast()
	l.lock.Unlock()
}

// Return current outstanding count
func (l *Limiter) Outstanding() int {
	l.lock.Lock()
	defer l.lock.Unlock()
	return l.numout
}

// Return current waiting count
func (l *Limiter) Waiting() int {
	l.lock.Lock()
	defer l.lock.Unlock()
	return l.waiting
}

type result struct {
	res interface{}
	err error
}

// Serializer restricts function calls to one at a time per key.
// Saved results from the first call are returned.
// (To only download a resource a single time.)
type Serializer struct {
	gates   map[interface{}][]chan<- result
	serials map[interface{}]*bool
	cancels map[interface{}]context.CancelFunc
	lock    sync.Mutex
}

// Create a new Serializer
func NewSerializer() *Serializer {
	g := new(Serializer)
	g.gates = make(map[interface{}][]chan<- result)
	g.serials = make(map[interface{}]*bool)
	g.cancels = make(map[interface{}]context.CancelFunc)
	return g
}

// Cancelled. Try again. Maybe.
var Cancelled = errors.New("cancelled")

// Call fn, gated by key.
// Subsequent calls with the same key will wait until the first returns,
// then all functions return the same result.
func (g *Serializer) Call(key interface{}, fn func() (interface{}, error)) (interface{}, error) {
	ctxfn := func(context.Context) (interface{}, error) {
		return fn()
	}
	return g.CallWithContext(key, context.Background(), ctxfn)
}

func (g *Serializer) makethecall(key interface{}, ctx context.Context, fn func(context.Context) (interface{}, error)) {
	var dead bool
	g.serials[key] = &dead
	ctx, cancel := context.WithCancel(ctx)
	g.cancels[key] = cancel
	g.lock.Unlock()

	res, err := fn(ctx)

	g.lock.Lock()
	defer g.lock.Unlock()
	cancel()
	// serial check, we may not know why ctx is cancelled
	if dead {
		return
	}
	// we won, clear space for next call and send results
	inflight := g.gates[key]
	delete(g.gates, key)
	delete(g.serials, key)
	delete(g.cancels, key)
	sendresults(res, err, inflight)
}

func (g *Serializer) CallWithContext(key interface{}, ctx context.Context, fn func(context.Context) (interface{}, error)) (interface{}, error) {
	g.lock.Lock()
	inflight, ok := g.gates[key]
	c := make(chan result)
	g.gates[key] = append(inflight, c)
	if !ok {
		// nobody going, start one
		go g.makethecall(key, ctx, fn)
	} else {
		g.lock.Unlock()
	}
	r := <-c
	return r.res, r.err
}

func sendresults(res interface{}, err error, chans []chan<- result) {
	r := result{res: res, err: err}
	for _, c := range chans {
		c <- r
		close(c)
	}
}

func (g *Serializer) cancel(key interface{}) {
	dead, ok := g.serials[key]
	if !ok {
		return
	}
	*dead = true
	delete(g.serials, key)
	cancel := g.cancels[key]
	cancel()
	delete(g.cancels, key)
	inflight := g.gates[key]
	sendresults(nil, Cancelled, inflight)
	delete(g.gates, key)
}

// Cancel any operations in progress.
// The calling function may block, but waiters will return immediately.
func (g *Serializer) Cancel(key interface{}) {
	g.lock.Lock()
	g.cancel(key)
	g.lock.Unlock()
}

// Cancel everything.
func (g *Serializer) CancelAll() {
	g.lock.Lock()
	for key := range g.serials {
		g.cancel(key)
	}
	g.lock.Unlock()
}

// End gate package

// Begin cache package
// Copyright (c) 2019 Ted Unangst <tedu@tedunangst.com>
//
// Permission to use, copy, modify, and distribute this software for any
// purpose with or without fee is hereby granted, provided that the above
// copyright notice and this permission notice appear in all copies.
//
// THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES
// WITH REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF
// MERCHANTABILITY AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR
// ANY SPECIAL, DIRECT, INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES
// WHATSOEVER RESULTING FROM LOSS OF USE, DATA OR PROFITS, WHETHER IN AN
// ACTION OF CONTRACT, NEGLIGENCE OR OTHER TORTIOUS ACTION, ARISING OUT OF
// OR IN CONNECTION WITH THE USE OR PERFORMANCE OF THIS SOFTWARE.

// A simple in memory, in process cache

// Fill functions should be roughtly compatible with this type.
// They may use stronger types, however.
// It will be called after a cache miss.
// It should return a value and bool indicating success.
type Filler func(key interface{}) (interface{}, bool)

// A function which returns the size of an element in the cache.
// It may be stronger typed.
type Sizer func(res interface{}) int

// A function which reduces a complex key into one suitable for a map.
// It may be stronger typed.
type Reducer func(key interface{}) interface{}

// Arguments to creating a new cache.
// Filler is required. See Filler type documentation.
// Entries will expire after Duration if set.
// Invalidator allows invalidating multiple dependent caches.
// Limit is max entries.
// SizeLimit is max size of all elements, combined with Sizer.
// Reducer allows for the use of complex keys.
type CacheOptions struct {
	Filler      interface{}
	Duration    time.Duration
	Invalidator *Invalidator
	Limit       int
	SizeLimit   int
	Sizer       interface{}
	Reducer     interface{}
	Singleton   bool
}

type entry struct {
	value interface{}
	size  int
	stale time.Time
}

type entrymap map[interface{}]entry

// The cache object
type Cache struct {
	cache      entrymap
	filler     Filler
	sizer      Sizer
	reducer    Reducer
	lock       sync.Mutex
	duration   time.Duration
	serializer *Serializer
	serialno   int
	limit      int
	size       int
	sizelimit  int
	singleton  bool
}

// An Invalidator is a collection of caches to be cleared or flushed together.
// It is created, then its address passed to cache creation.
type Invalidator struct {
	caches []*Cache
}

// Create a new Cache. Arguments are provided via Options.
func NewCache(options CacheOptions) *Cache {
	c := new(Cache)
	c.cache = make(entrymap)
	if fillfn := options.Filler; fillfn != nil {
		ftype := reflect.TypeOf(fillfn)
		if ftype.Kind() != reflect.Func {
			panic("cache filler is not function")
		}
		if ftype.NumIn() != 1 || ftype.NumOut() != 2 {
			panic("cache filler has wrong argument count")
		}
		vfn := reflect.ValueOf(fillfn)
		c.filler = func(key interface{}) (interface{}, bool) {
			args := []reflect.Value{reflect.ValueOf(key)}
			rv := vfn.Call(args)
			return rv[0].Interface(), rv[1].Bool()
		}
	}
	if sizefn := options.Sizer; sizefn != nil {
		ftype := reflect.TypeOf(sizefn)
		if ftype.Kind() != reflect.Func {
			panic("cache sizer is not function")
		}
		if ftype.NumIn() != 1 || ftype.NumOut() != 1 {
			panic("cache sizer has wrong argument count")
		}
		vfn := reflect.ValueOf(sizefn)
		c.sizer = func(res interface{}) int {
			args := []reflect.Value{reflect.ValueOf(res)}
			rv := vfn.Call(args)
			return int(rv[0].Int())
		}
	}
	if reducefn := options.Reducer; reducefn != nil {
		ftype := reflect.TypeOf(reducefn)
		if ftype.Kind() != reflect.Func {
			panic("cache sizer is not function")
		}
		if ftype.NumIn() != 1 || ftype.NumOut() != 1 {
			panic("cache sizer has wrong argument count")
		}
		vfn := reflect.ValueOf(reducefn)
		c.reducer = func(res interface{}) interface{} {
			args := []reflect.Value{reflect.ValueOf(res)}
			rv := vfn.Call(args)
			return rv[0].Interface()
		}
	}
	if options.Duration != 0 {
		c.duration = options.Duration
	}
	if options.Invalidator != nil {
		options.Invalidator.caches = append(options.Invalidator.caches, c)
	}
	c.serializer = NewSerializer()
	c.limit = options.Limit
	c.sizelimit = options.SizeLimit
	c.singleton = options.Singleton
	return c
}

// Get a value for a key. Returns true for success.
// Will automatically fill the cache.
// Returns holding the cache lock. Useful when the cached value can mutate.
func (c *Cache) GetAndLock(key interface{}, value interface{}) bool {
	origkey := key
	if c.reducer != nil {
		key = c.reducer(key)
	}
	c.lock.Lock()
recheck:
	ent, ok := c.cache[key]
	if ok {
		if !ent.stale.IsZero() && ent.stale.Before(time.Now()) {
			c.remove(key, ent)
			ok = false
		}
	}
	if !ok {
		if c.filler == nil {
			return false
		}
		serial := c.serialno
		c.lock.Unlock()
		r, err := c.serializer.Call(key, func() (interface{}, error) {
			v, ok := c.filler(origkey)
			if !ok {
				return nil, errors.New("no fill")
			}
			return v, nil
		})
		c.lock.Lock()
		if err == Cancelled || serial != c.serialno {
			goto recheck
		}
		if err == nil {
			c.set(key, r)
			ent.value, ok = r, true
		}
	}
	if ok {
		ptr := reflect.ValueOf(ent.value)
		reflect.ValueOf(value).Elem().Set(ptr)
	}
	return ok
}

// Get a value for a key. Returns true for success.
// Will automatically fill the cache.
func (c *Cache) Get(key interface{}, value interface{}) bool {
	rv := c.GetAndLock(key, value)
	c.lock.Unlock()
	return rv
}

func (c *Cache) set(key interface{}, value interface{}) {
	var stale time.Time
	if c.duration != 0 {
		stale = time.Now().Add(c.duration)
	}
	size := 0
	if c.sizer != nil {
		size = c.sizer(value)
	}
	if c.limit > 0 && len(c.cache) >= c.limit {
		tries := 0
		var now time.Time
		if c.duration != 0 {
			now = time.Now()
		} else {
			tries = 5
		}
		for key, ent := range c.cache {
			if tries < 5 && ent.stale.After(now) {
				tries++
				continue
			}
			c.remove(key, ent)
			break
		}
	}
	if c.sizelimit > 0 {
		if size > c.sizelimit/4 {
			return
		}
		if size+c.size > c.sizelimit {
			for key, ent := range c.cache {
				c.remove(key, ent)
				if size+c.size <= c.sizelimit {
					break
				}
			}
		}
	}
	c.size += size
	c.cache[key] = entry{
		value: value,
		stale: stale,
		size:  size,
	}
}

func (c *Cache) remove(key interface{}, ent entry) {
	c.size -= ent.size
	delete(c.cache, key)
}

// Manually set a cached value.
func (c *Cache) Set(key interface{}, value interface{}) {
	if c.reducer != nil {
		key = c.reducer(key)
	}
	c.lock.Lock()
	c.set(key, value)
	c.lock.Unlock()
}

// Unlock the c, iff lock is held.
func (c *Cache) Unlock() {
	c.lock.Unlock()
}

// Clear one key from the cache
func (c *Cache) Clear(key interface{}) {
	if c.singleton {
		c.Flush()
		return
	}
	if c.reducer != nil {
		key = c.reducer(key)
	}
	c.lock.Lock()
	if ent, ok := c.cache[key]; ok {
		c.serialno++
		c.remove(key, ent)
	}
	c.serializer.Cancel(key)
	c.lock.Unlock()
}

// Flush all values from the cache
func (c *Cache) Flush() {
	c.lock.Lock()
	c.serialno++
	c.cache = make(entrymap)
	c.serializer.CancelAll()
	c.lock.Unlock()
}

// Clear one key from associated caches
func (inv Invalidator) Clear(key interface{}) {
	for _, c := range inv.caches {
		c.Clear(key)
	}
}

// Flush all values from associated caches
func (inv Invalidator) Flush() {
	for _, c := range inv.caches {
		c.Flush()
	}
}

type Counter struct {
	cache *Cache
}

func NewCounter(options CacheOptions) Counter {
	var c Counter
	c.cache = NewCache(CacheOptions{Filler: func(name string) (int64, bool) {
		return 0, true
	}, Duration: options.Duration, Limit: options.Limit})
	return c
}

func (cnt Counter) Get(name string) int64 {
	c := cnt.cache
	var val int64
	c.Get(name, &val)
	return val
}

func (cnt Counter) Inc(name string) int64 {
	c := cnt.cache
	var val int64
	c.GetAndLock(name, &val)
	val++
	c.set(name, val)
	c.Unlock()
	return val
}

func (cnt Counter) Dec(name string) int64 {
	c := cnt.cache
	var val int64
	c.GetAndLock(name, &val)
	val--
	c.set(name, val)
	c.Unlock()
	return val
}

// End cache package

// TODO rework this

type UserInfo struct {
	UserID   int
	Username string
	Role     Role
}

type keytype struct{}

var thekey keytype

var dbtimeformat = "2006-01-02 15:04:05"

// Check for auth cookie. Allows failure.
func Checker(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userinfo, ok := checkauthcookie(r)
		if ok {
			ctx := context.WithValue(r.Context(), thekey, userinfo)
			r = r.WithContext(ctx)
		}
		handler.ServeHTTP(w, r)
	})
}

// Check for auth cookie. On failure redirects to /login.
// Must already be wrapped in Checker.
func Required(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ok := GetUserInfo(r) != nil
		if !ok {
			loginredirect(w, r)
			return
		}
		handler.ServeHTTP(w, r)
	})
}

// Only accessible to certain roles
func Roles(handler http.Handler, roles []Role) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		u := GetUserInfo(r)
		if u == nil || !slices.Contains(roles, u.Role) {
			loginredirect(w, r) // TODO unauth?
			return
		}
		handler.ServeHTTP(w, r)
	})
}

// May need to rework these if I think permissions need extention

// Minimum level of mod (mod or admin)
func Mod(handler http.Handler) http.Handler {
	return Roles(handler, []Role{RoleAdmin, RoleMod})
}

func Admin(handler http.Handler) http.Handler {
	return Roles(handler, []Role{RoleAdmin})
}

// Get UserInfo for this request, if any.
func GetUserInfo(r *http.Request) *UserInfo {
	userinfo, ok := r.Context().Value(thekey).(*UserInfo)
	if !ok {
		return nil
	}
	return userinfo
}

// TODO define a full get user?

func calculateCSRF(salt, auth string) string {
	hasher := sha512.New512_256()
	zero := []byte{0}
	hasher.Write(zero)
	hasher.Write([]byte(auth))
	hasher.Write(zero)
	hasher.Write([]byte(csrfkey))
	hasher.Write(zero)
	hasher.Write([]byte(salt))
	hasher.Write(zero)
	hash := hexsum(hasher)

	return salt + hash
}

// Get a CSRF token
func GetCSRF(r *http.Request) string {
	_, ok := checkauthcookie(r)
	if !ok {
		return ""
	}
	auth := getauthcookie(r)
	if auth == "" {
		return ""
	}
	hasher := sha512.New512_256()
	io.CopyN(hasher, rand.Reader, 32)
	salt := hexsum(hasher)

	return calculateCSRF(salt, auth)
}

// Checks that CSRF value is correct.
func CheckCSRF(r *http.Request) bool {
	auth := getauthcookie(r)
	if auth == "" {
		return false
	}
	csrf := r.FormValue("CSRF")
	if len(csrf) != authlen*2 {
		return false
	}
	salt := csrf[0:authlen]
	rv := calculateCSRF(salt, auth)
	ok := subtle.ConstantTimeCompare([]byte(rv), []byte(csrf)) == 1
	return ok
}

// Wrap a handler with CSRF checking.
func CSRFWrap(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ok := CheckCSRF(r)
		if !ok {
			http.Error(w, "invalid csrf", 403)
			return
		}
		handler.ServeHTTP(w, r)
	})
}

func loginredirect(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{
		Name:     "auth",
		Value:    "",
		MaxAge:   -1,
		Secure:   securecookies,
		HttpOnly: true,
		SameSite: getsamesite(r),
	})
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

var authregex = regexp.MustCompile("^[[:alnum:]]+$")
var authlen = 32

var stmtUserName, stmtUserAuth, stmtUpdateUser, stmtSaveAuth, stmtDeleteAuth *sql.Stmt
var stmtUpdateExpiry, stmtDeleteOneAuth *sql.Stmt
var csrfkey string
var securecookies bool
var samesitecookie http.SameSite
var safariworks bool

func getconfig(db *sql.DB, key string, value interface{}) error {
	row := db.QueryRow("select value from config where key = ?", key)
	err := row.Scan(value)
	if err == sql.ErrNoRows {
		err = nil
	}
	return err
}

type LoginInitArgs struct {
	Db             *sql.DB
	Insecure       bool
	SameSiteStrict bool
	SafariWorks    bool
}

// Init. Must be called with the database.
// Requires a users table with (id, username, hash) columns and a
// auth table with (userid, hash, expiry) columns.
// Requires a config table with (key, value) ('csrfkey', some secret).
func LoginInit(args LoginInitArgs) {
	db := args.Db
	var err error
	stmtUserName, err = db.Prepare("select id, hash from users where username = ? and id > 0")
	if err != nil {
		panic(err)
	}
	stmtUserAuth, err = db.Prepare("select users.id, username, role, expiry from users join auth on users.id= auth.userid where auth.hash = ? and expiry > ?")
	if err != nil {
		panic(err)
	}
	stmtUpdateUser, err = db.Prepare("update users set hash = ? where id = ?")
	if err != nil {
		panic(err)
	}
	stmtSaveAuth, err = db.Prepare("insert into auth (userid, hash, expiry) values (?, ?, ?)")
	if err != nil {
		panic(err)
	}
	stmtDeleteAuth, err = db.Prepare("delete from auth where userid = ?")
	if err != nil {
		panic(err)
	}
	stmtUpdateExpiry, err = db.Prepare("update auth set expiry = ? where hash = ?")
	if err != nil {
		panic(err)
	}
	stmtDeleteOneAuth, err = db.Prepare("delete from auth where hash = ?")
	if err != nil {
		panic(err)
	}
	securecookies = !args.Insecure
	if args.SameSiteStrict {
		samesitecookie = http.SameSiteStrictMode
	}
	safariworks = args.SafariWorks
	getconfig(db, "csrfkey", &csrfkey)
}

func getauthcookie(r *http.Request) string {
	cookie, err := r.Cookie("auth")
	if err != nil {
		return ""
	}
	auth := cookie.Value
	if !(len(auth) == authlen && authregex.MatchString(auth)) {
		return ""
	}
	return auth
}

func getsamesite(r *http.Request) http.SameSite {
	var samesite http.SameSite
	if safariworks || !strings.Contains(r.UserAgent(), "iPhone") {
		samesite = samesitecookie
	}
	return samesite
}

func getformtoken(r *http.Request) string {
	token := r.FormValue("token")
	if token == "" {
		token = r.Header.Get("Authorization")
	}
	if strings.HasPrefix(token, "Bearer ") {
		token = token[7:]
	}
	if token == "" {
		return ""
	}
	if !(len(token) == authlen && authregex.MatchString(token)) {
		log.Info("login: bad token: %s", token)
		return ""
	}
	return token
}

var validcookies = NewCache(CacheOptions{Filler: func(cookie string) (*UserInfo, bool) {
	hasher := sha512.New512_256()
	hasher.Write([]byte(cookie))
	authhash := hexsum(hasher)
	now := time.Now().UTC()
	row := stmtUserAuth.QueryRow(authhash, now.Format(dbtimeformat))
	var userinfo UserInfo
	var stamp string
	err := row.Scan(&userinfo.UserID, &userinfo.Username, &userinfo.Role, &stamp)
	if err != nil {
		if err == sql.ErrNoRows {
			log.Info("login: no auth found")
		} else {
			log.Info("login: error scanning auth row: %s", err)
		}
		return nil, false
	}
	expiry, _ := time.Parse(dbtimeformat, stamp)
	if expiry.Before(now.Add(3 * 24 * time.Hour)) {
		stmtUpdateExpiry.Exec(now.Add(7*24*time.Hour).Format(dbtimeformat), authhash)
	}

	return &userinfo, true
}, Duration: 1 * time.Minute})

func checkauthcookie(r *http.Request) (*UserInfo, bool) {
	cookie := getauthcookie(r)
	if cookie == "" {
		return nil, false
	}
	var userinfo *UserInfo
	ok := validcookies.Get(cookie, &userinfo)
	return userinfo, ok
}

func checkformtoken(r *http.Request) (*UserInfo, bool) {
	token := getformtoken(r)
	if token == "" {
		return nil, false
	}
	var userinfo *UserInfo
	ok := validcookies.Get(token, &userinfo)
	return userinfo, ok
}

func loaduser(username string) (int, string, bool) {
	row := stmtUserName.QueryRow(username) // TODO rename
	var userid int
	var hash string
	err := row.Scan(&userid, &hash)
	if err != nil {
		if err == sql.ErrNoRows {
			log.Info("login: no username found")
		} else {
			log.Info("login: error loading username: %s", err)
		}
		return -1, "", false
	}
	return userid, hash, true
}

var userregex = regexp.MustCompile("^[[:alnum:]]+$")
var userlen = 32
var passlen = 128

func hexsum(h hash.Hash) string {
	return fmt.Sprintf("%x", h.Sum(nil))[0:authlen]
}

func loginSession(w http.ResponseWriter, r *http.Request, userid int) error {
	hasher := sha512.New512_256()
	io.CopyN(hasher, rand.Reader, 32)
	auth := hexsum(hasher)

	maxage := 3600 * 24 * 365
	http.SetCookie(w, &http.Cookie{
		Name:     "auth",
		Value:    auth,
		MaxAge:   maxage,
		Secure:   securecookies,
		SameSite: getsamesite(r),
		Path:     "/",
		HttpOnly: true,
	})

	hasher.Reset()
	hasher.Write([]byte(auth))
	authhash := hexsum(hasher)

	expiry := time.Now().UTC().Add(7 * 24 * time.Hour).Format(dbtimeformat)
	_, err := stmtSaveAuth.Exec(userid, authhash, expiry)
	if err != nil {
		return err
	}
	return nil
}

// Default handler for /dologin
// Requires username and password form values.
// Redirects to / on success and /login on failure.
func LoginFunc(w http.ResponseWriter, r *http.Request) {
	username := r.FormValue("username")
	password := r.FormValue("password")
	if len(username) == 0 || len(username) > userlen ||
		!userregex.MatchString(username) || len(password) == 0 ||
		len(password) > passlen {
		SetCookieValue(w, "login-err", "Invalid username or invalid password.")
		loginredirect(w, r)
		return
	}
	userid, hash, ok := loaduser(username)
	if !ok {
		SetCookieValue(w, "login-err", "Account does not exist.")
		loginredirect(w, r)
		return
	}
	match, err := argon2id.ComparePasswordAndHash(password, hash)
	// TODO better error handling
	if !match || err != nil {
		SetCookieValue(w, "login-err", "Incorrect password.")
		loginredirect(w, r)
		return
	}
	loginSession(w, r, userid)
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func deleteauth(userid int) error {
	defer validcookies.Flush()
	_, err := stmtDeleteAuth.Exec(userid)
	return err
}

func deleteoneauth(auth string) error {
	defer validcookies.Flush()
	hasher := sha512.New512_256()
	hasher.Write([]byte(auth))
	authhash := hexsum(hasher)
	_, err := stmtDeleteOneAuth.Exec(authhash)
	return err
}

// Handler for /dologout route.
func LogoutFunc(w http.ResponseWriter, r *http.Request) {
	userinfo, ok := checkauthcookie(r)
	if ok && CheckCSRF(r) {
		err := deleteauth(userinfo.UserID)
		if err != nil {
			log.Info("login: error deleting old auth: %s", err)
			http.Redirect(w, r, "/", http.StatusSeeOther)
			return
		}
		http.SetCookie(w, &http.Cookie{
			Name:     "auth",
			Value:    "",
			MaxAge:   -1,
			Secure:   securecookies,
			HttpOnly: true,
		})
	}
	_, ok = checkformtoken(r)
	if ok {
		auth := getformtoken(r)
		deleteoneauth(auth)
		return
	}
	http.Redirect(w, r, "/", http.StatusSeeOther)
}
