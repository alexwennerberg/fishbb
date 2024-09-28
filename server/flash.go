package server

import (
	"encoding/base64"
	"net/http"
	"time"
)

// simple flash messages
// from alex edwards

// SetCookieValue is used to store some arbitrary value in a cookie
func SetCookieValue(w http.ResponseWriter, name string, value string) {
	c := &http.Cookie{Name: name, Value: encode([]byte(value))}
	http.SetCookie(w, c)
}

// GetCookieValue gets a velue stored in a cookie
func GetCookieValue(r *http.Request, name string) (string, error) {
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
	return string(value), nil
}

// GetFlash gets a cookie value and resets it
func GetFlash(w http.ResponseWriter, r *http.Request, name string) (string, error) {
	value, err := GetCookieValue(r, name)
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
