package server

import (
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/httprate"
)

// stolen from https://github.com/go-chi/chi/pull/386

// ChangePostToHiddenMethod looks for the _hidden attribute of forms so that we
// can use DELETE and PUT in <form> submissions. This is, of course, a
// non-standard "hack"
func ChangePostToHiddenMethod(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			next.ServeHTTP(w, r)
			return
		}
		method := r.FormValue("_method")
		if method == "POST" || method == "DELETE" {
			r.Method = method
		}
		next.ServeHTTP(w, r)
	})
}

func LimitByUser(requestLimit int, windowLength time.Duration) func(next http.Handler) http.Handler {
	return httprate.Limit(requestLimit, windowLength, httprate.WithLimitHandler(tooManyRequests), httprate.WithKeyFuncs(KeyByUserID))
}

func LimitByRealIP(requestLimit int, windowLength time.Duration) func(next http.Handler) http.Handler {
	return httprate.Limit(requestLimit, windowLength, httprate.WithLimitHandler(tooManyRequests), httprate.WithKeyFuncs(httprate.KeyByRealIP))
}

func KeyByUserID(r *http.Request) (string, error) {
	return fmt.Sprintf("%d", GetUserInfo(r).UserID), nil
}
