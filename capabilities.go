package main

//

// TODO consistent verbiage
const (
	editPosts = 1 << iota
	deletePosts
	updateThreadMeta
	modifyUser
	banUser
	setUserRole
)

const AdminPerms = (iota << 1) - 1
const ModPerms = editPosts & deletePosts & updateThreadMeta

// order doesnt matter
func can(a, b int) bool {
	return a&b > 0
}
