package main

import "time"

type Notification struct {
	ID      int
	Message string
	Created time.Time
}

// Don't keep notifications beyond this
const maxNotifications = 128

func getNotifications(uid int) ([]Notification, error) {
	return nil, nil
}

func storeNotification(uid int, message string) ([]Notification, error) {
	return nil, nil
}
