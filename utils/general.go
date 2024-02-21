package utils

import "zenoforge.com/goLiveNotif/models"

func FindPostIndexByID(id int, posts []models.Post) (int, bool) {
	for index, post := range posts {
		if post.ID == id {
			return index, true
		}
	}
	return -1, false
}
