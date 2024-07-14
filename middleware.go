package main

import "net/http"

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
