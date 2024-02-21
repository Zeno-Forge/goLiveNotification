package handlers

import (
	"io"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/labstack/echo/v4"
	"zenoforge.com/goLiveNotif/models"
	"zenoforge.com/goLiveNotif/templates"
	"zenoforge.com/goLiveNotif/utils"
)

var Posts []models.Post

func CreatePost(c echo.Context) error {
	var newPost = models.Post{
		Message: models.DiscordMessage{
			Embed: []models.Embed{
				{
					Thumbnail: models.URL{URL: "https://static-cdn.jtvnw.net/jtv_user_pictures/f77022c4-2d3e-45ff-a7a9-22ada2688c50-profile_image-300x300.png"},
					URL:       "https://twitch.tv/marshievt",
					Color:     6504867,
					Footer: models.Footer{
						IconURL: "https://www.freepnglogos.com/uploads/twitch-logo-transparent-png-20.png",
						Text:    "Twitch",
					},
				},
			},
		},
	}

	// Assign a new unique ID
	newID := 1
	if len(Posts) > 0 {
		newID = Posts[len(Posts)-1].ID + 1
	}
	newPost.ID = newID

	Posts = append(Posts, newPost)

	if err := utils.SaveDataToFile(Posts, "data", "postStorage.json"); err != nil {
		return err
	}

	postModal := templates.PostModal(newPost)

	return utils.Render(c, http.StatusOK, postModal)
}

func GetPost(c echo.Context) error {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"message": "Invalid post ID"})
	}

	index, found := utils.FindPostIndexByID(id, Posts)
	if !found {
		return c.JSON(http.StatusNotFound, map[string]string{"message": "Post not found"})
	}

	postModal := templates.PostModal(Posts[index])

	return utils.Render(c, http.StatusOK, postModal)
}

func EditPost(c echo.Context) error {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"message": "Invalid post ID"})
	}

	index, found := utils.FindPostIndexByID(id, Posts)
	if !found {
		return c.JSON(http.StatusNotFound, map[string]string{"message": "Post not found"})
	}

	color, _ := strconv.Atoi(c.FormValue("colorInput"))
	schedule, _ := time.Parse("2006-01-02T15:04", c.FormValue("scheduleInput"))

	var updatedPost = models.Post{
		ID:         id,
		ScheduleAt: schedule,
		Message: models.DiscordMessage{
			Content: c.FormValue("contentInput"),
			Embed: []models.Embed{
				{
					Title:       c.FormValue("titleInput"),
					Description: c.FormValue("descriptionInput"),
					URL:         c.FormValue("urlInput"),
					Color:       color,
					Thumbnail:   models.URL{URL: c.FormValue("thumbnailInput")},
					Footer: models.Footer{
						IconURL: c.FormValue("footerIconInput"),
						Text:    c.FormValue("footerTextInput"),
					},
				},
			},
		},
	}

	file, err := c.FormFile("imageUpload")
	if err == nil {
		src, err := file.Open()

		if err != nil {
			return err
		}

		defer src.Close()
		dirPath := "./data/uploads/" + c.Param("id")
		os.MkdirAll(dirPath, 0755)
		dstPath := dirPath + "/" + file.Filename

		dst, err := os.Create(dstPath)
		if err != nil {
			return err
		}

		defer dst.Close()

		if _, err = io.Copy(dst, src); err != nil {
			return err
		}

		updatedPost.Message.Embed[0].Image.URL = dstPath
	}

	updatedPost.ID = id

	Posts[index] = updatedPost

	if err := utils.SaveDataToFile(Posts, "data", "postStorage.json"); err != nil {
		return err
	}

	postListComp := templates.PostsTempl(Posts)

	return utils.Render(c, http.StatusOK, postListComp)
}

func DeletePost(c echo.Context) error {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"message": "Invalid post ID"})
	}

	index, found := utils.FindPostIndexByID(id, Posts)
	if !found {
		return c.JSON(http.StatusNotFound, map[string]string{"message": "Post not found"})
	}

	// Remove the post from the slice
	Posts = append(Posts[:index], Posts[index+1:]...)

	if err := utils.SaveDataToFile(Posts, "data", "postStorage.json"); err != nil {
		return err
	}

	postListComp := templates.PostsTempl(Posts)

	return utils.Render(c, http.StatusOK, postListComp)
}
