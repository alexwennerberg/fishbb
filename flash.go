package main

import (
	"encoding/base64"
	"net/http"
	"time"
)

// simple flash messages
// from alex edwards

func SetFlash(w http.ResponseWriter, name string, value string) {
	c := &http.Cookie{Name: name, Value: encode([]byte(value))}
	http.SetCookie(w, c)
}

func GetFlash(w http.ResponseWriter, r *http.Request, name string) (string, error) {
	c, err := r.Cookie(name)
	if err != nil {
		switch err {
		case http.ErrNoCookie:
			return "", nil
		default:
			return "", err
		}
	}
	value, err := decode(c.Value)
	if err != nil {
		return "", err
	}
	dc := &http.Cookie{Name: name, MaxAge: -1, Expires: time.Unix(1, 0)}
	http.SetCookie(w, dc)
	return string(value), nil
}

func encode(src []byte) string {
	return base64.URLEncoding.EncodeToString(src)
}

func decode(src string) ([]byte, error) {
	return base64.URLEncoding.DecodeString(src)
}

