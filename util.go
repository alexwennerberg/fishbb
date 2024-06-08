package main

import (
	"net/http"
	"strings"
)

// generic log err function
// non-blocking
func logIfErr(err error) {
	if err != nil {
		log.Error("unexpected error:", "error", err)
	}
}

// Some Name -> some-name
func slugify(s string) string {
	return strings.ReplaceAll(strings.ToLower(s), " ", "-")
}

func logRequest(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Info("request", "method", r.Method, "url", r.URL)
		handler.ServeHTTP(w, r)
	})
}
