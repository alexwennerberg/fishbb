package main

// Capabilities stores the ability to perform certain actions.
// capabilities must be attached to a role. In the future TODO
// administrators can create roles

// TODO consistent verbiage
const (
	// Can edit any post
	editPosts = 1 << iota
	// Can delete any post
	deletePosts
	// Can update thread metadata (pin or lock threads)
	updateThreadMeta
	// Can ban users
	banUser
	// can update user roles
	setUserRole
	// access the admin panel (only a view role without other perms)
	viewAdminPanel
)

const AdminPerms = (iota << 1) - 1
const ModPerms = editPosts & deletePosts & updateThreadMeta
const UserPerms = 0

// order doesnt matter
func can(a, b int) bool {
	return a&b > 0
}
