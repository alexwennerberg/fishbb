package main

import (
	"encoding/xml"
	"fmt"
	"net/http"
	"time"
)

type AtomFeed struct {
	XMLName xml.Name    `xml:"feed"`
	XMLNS   string      `xml:"xmlns,attr"`
	Title   string      `xml:"title"`
	ID      string      `xml:"id"`
	Links   []AtomLink  `xml:"link"`
	Updated string      `xml:"updated"`
	Entries []AtomEntry `xml:"entry"`
}

type AtomLink struct {
	Href string `xml:"href,attr"`
	Rel  string `xml:"rel,attr,omitempty"`
	Type string `xml:"type,attr,omitempty"`
}

type AtomEntry struct {
	Title   string      `xml:"title"`
	ID      string      `xml:"id"`
	Link    AtomLink    `xml:"link"`
	Updated string      `xml:"updated"`
	Content AtomContent `xml:"content"`
	Author  AtomAuthor  `xml:"author"`
}

type AtomContent struct {
	Type string `xml:"type,attr"`
	Text string `xml:",chardata"`
}

type AtomAuthor struct {
	Name string `xml:"name"`
}

func baseURL(r *http.Request) string {
	scheme := "https"
	if r.TLS == nil && r.Header.Get("X-Forwarded-Proto") == "" {
		scheme = "http"
	}
	return fmt.Sprintf("%s://%s", scheme, r.Host)
}

func getRecentForumPosts(forumID int, limit int) ([]Post, error) {
	var posts []Post
	rows, err := db.Query(`
		select post.id, post.content, user.id, user.username, post.created, post.edited,
		       thread.title
		from post
		join user on post.authorid = user.id
		join thread on post.threadid = thread.id
		where thread.forumid = ?
		order by post.created desc
		limit ?`, forumID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var p Post
		err := rows.Scan(&p.ID, &p.Content, &p.Author.ID, &p.Author.Username,
			&p.Created, &p.Edited, &p.ThreadTitle)
		if err != nil {
			return nil, err
		}
		posts = append(posts, p)
	}
	return posts, nil
}

func forumFeedPage(w http.ResponseWriter, r *http.Request) {
	forum, err := getForumBySlug(r.PathValue("forum"))
	if err != nil {
		notFound(w, r)
		return
	}

	posts, err := getRecentForumPosts(forum.ID, 50)
	if err != nil {
		serverError(w, r, err)
		return
	}

	base := baseURL(r)
	feedURL := fmt.Sprintf("%s%s/feed", base, forum.Slug)

	var updated string
	if len(posts) > 0 {
		updated = posts[0].Created.Format(time.RFC3339)
	} else {
		updated = time.Now().UTC().Format(time.RFC3339)
	}

	feed := AtomFeed{
		XMLNS: "http://www.w3.org/2005/Atom",
		Title: fmt.Sprintf("%s - %s", forum.Name, config.BoardName),
		ID:    feedURL,
		Links: []AtomLink{
			{Href: feedURL, Rel: "self", Type: "application/atom+xml"},
			{Href: base + forum.Slug, Rel: "alternate", Type: "text/html"},
		},
		Updated: updated,
	}

	for _, p := range posts {
		slug := p.Slug()
		if slug == "" {
			continue
		}
		postURL := base + slug

		entryUpdated := p.Created.Format(time.RFC3339)
		if p.Edited != nil {
			entryUpdated = p.Edited.Format(time.RFC3339)
		}

		entry := AtomEntry{
			Title:   p.ThreadTitle,
			ID:      postURL,
			Link:    AtomLink{Href: postURL},
			Updated: entryUpdated,
			Content: AtomContent{Type: "html", Text: string(p.Render())},
			Author:  AtomAuthor{Name: p.Author.Username},
		}
		feed.Entries = append(feed.Entries, entry)
	}

	w.Header().Set("Content-Type", "application/atom+xml; charset=utf-8")
	w.Write([]byte(xml.Header))
	enc := xml.NewEncoder(w)
	enc.Indent("", "  ")
	enc.Encode(feed)
}
