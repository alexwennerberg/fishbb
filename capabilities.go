package main

//

type Capability int

// TODO consistent verbiage
const (
	editPosts = 1 << iota 
	deletePosts
	updateThreadMeta
	modifyUser
)

const AdminPerms = (iota << 1) - 1
const ModPerms = editPosts & deletePosts & updateThreadMeta


