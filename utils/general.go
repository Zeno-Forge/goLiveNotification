package utils

import (
	"os"

	"zenoforge.com/goLiveNotif/models"
)

func FindPostIndexByID(id string, posts []models.Post) (int, bool) {
	for index, post := range posts {
		if post.ID == id {
			return index, true
		}
	}
	return -1, false
}

func GetEnv(key string) (string, bool) {
	val, ok := os.LookupEnv(key)
	return val, ok
}
