package main

import (
	"bytes"
	"crypto/rand"
	"crypto/sha512"
	"fmt"
	"html"
	"html/template"
	"image"
	"image/png"
	"io"
	"math/big"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"strings"
)

var timeISO8601 = "2006-01-02 15:04:05"

// stolen from git snippet

func init() {
	assertAvailablePRNG()
}

func assertAvailablePRNG() {
	// Assert that a cryptographically secure PRNG is available.
	// Panic otherwise.
	buf := make([]byte, 1)

	_, err := io.ReadFull(rand.Reader, buf)
	if err != nil {
		panic(fmt.Sprintf("crypto/rand is unavailable: Read() failed with %#v", err))
	}
}

// GenerateRandomString returns a securely generated random string.
// It will return an error if the system's secure random
// number generator fails to function correctly, in which
// case the caller should not continue.
func GenerateRandomString(n int) (string, error) {
	const letters = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz-"
	ret := make([]byte, n)
	for i := 0; i < n; i++ {
		num, err := rand.Int(rand.Reader, big.NewInt(int64(len(letters))))
		if err != nil {
			return "", err
		}
		ret[i] = letters[num.Int64()]
	}

	return string(ret), nil
}

// returns page list
func pageArray(n int) []int {
	c := ((n - 1) / config.PageSize) + 1
	p := make([]int, c)
	for i := range c {
		p[i] = i + 1
	}
	return p
}

// returns limit, offset
// 1-indexed
func paginate(page int) (int, int) {
	limit := config.PageSize
	offset := (page - 1) * limit
	return limit, offset
}

// pulls page from url param. 1-indexed
func page(r *http.Request) int {
	page, _ := strconv.Atoi(r.URL.Query().Get("p"))
	if page == 0 { // 1-indexing
		page = 1
	}
	return page
}

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

// Generate a simple avatar based on a hash of your name
//
// Derived from Ted Unangst's Honk
// https://humungus.tedunangst.com/r/honk/v/tip/f/avatar.go
//
// Copyright (c) 2019 Alex Wennerberg, Ted Unangst <tedu@tedunangst.com>
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

// Modify the four colors used
// TODO align with css
var avatarcolors = [4][4]byte{
	{0, 0, 0, 255},
	{255, 255, 255, 255},
	{172, 172, 172, 255},
	{96, 96, 96, 255},
}

// x and y dimensions in pixels
const size = 400

// generate PNG avatar from string
func genAvatar(name string) []byte {
	h := sha512.New()
	h.Write([]byte(name))
	s := h.Sum(nil)
	// Mess with these numbers to change the size
	img := image.NewNRGBA(image.Rect(0, 0, size, size))
	for i := 0; i < size; i++ {
		for j := 0; j < size; j++ {
			p := i*img.Stride + j*4
			xx := i/(size/4)*16 + j/(size/4)
			x := s[xx]
			for n := 0; n < 4; n++ {
				img.Pix[p+n] = avatarcolors[x/64][n]
			}
		}
	}
	var buf bytes.Buffer
	png.Encode(&buf, img)
	return buf.Bytes()
}

// turns a url into a string, but only if it's http/https
func addLinkTags(link string) template.HTML {
	u, err := url.Parse(link)
	if err != nil || (u.Scheme != "http" && u.Scheme != "https") {
		return template.HTML(html.EscapeString(link))
	}
	return template.HTML(fmt.Sprintf("<a href='%[1]s'>%[1]s</a>", html.EscapeString(link)))
}

const solarYearSecs = 31556926

// timeago if < 1 year, else yyyy-mm-dd
func timeago(t *time.Time) string {
	d := time.Since(*t)
	var metric string
	var amount int
	if d.Seconds() < 60 {
		amount = int(d.Seconds())
		metric = "second"
	} else if d.Minutes() < 60 {
		amount = int(d.Minutes())
		metric = "minute"
	} else if d.Hours() < 24 {
		amount = int(d.Hours())
		metric = "hour"
	} else if d.Seconds() < solarYearSecs {
		amount = int(d.Hours()) / 24
		metric = "day"
	} else {
		return t.Format("2006-02-01")
	}
	if amount == 1 {
		return fmt.Sprintf("%d %s ago", amount, metric)
	} else {
		return fmt.Sprintf("%d %ss ago", amount, metric)
	}
}
