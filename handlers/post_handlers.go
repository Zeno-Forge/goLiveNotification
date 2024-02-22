package handlers

import (
	"fmt"
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
	var discordConf models.DiscordConfig
	err := utils.LoadDataFromFile(&discordConf, "data", "config.json")
	if err != nil {
		return err
	}

	contentString := fmt.Sprintf("<@&%s>", discordConf.Discord.RoleID)

	var newPost = models.Post{
		ScheduleAt: time.Now(),
		Message: models.DiscordMessage{
			Content: contentString,
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

	postModal := templates.PostModal(newPost)

	return utils.Render(c, http.StatusOK, postModal)
}

func GetPostList(c echo.Context) error {
	postListComp := templates.PostsTempl(Posts)

	return utils.Render(c, http.StatusOK, postListComp)
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

	color, _ := strconv.Atoi(c.FormValue("colorInput"))

	timezone := c.FormValue("timezone")
	loc, err := time.LoadLocation(timezone)
	if err != nil {
		return err
	}
	schedule, err := time.ParseInLocation("2006-01-02T15:04", c.FormValue("scheduleInput"), loc)
	if err != nil {
		return err
	}

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

	index, found := utils.FindPostIndexByID(id, Posts)
	if !found {
		Posts = append(Posts, updatedPost)
	} else {
		Posts[index] = updatedPost
	}

	if err := utils.SaveDataToFile(Posts, "data", "postStorage.json"); err != nil {
		return err
	}

	ManageScheduledPosts()

	postListComp := templates.PostsTempl(Posts)

	return utils.Render(c, http.StatusOK, postListComp)
}

func DeletePostHandler(c echo.Context) error {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"message": "Invalid post ID"})
	}

	err = DeletePost(id)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"message": "Unable to delete post"})
	}

	ManageScheduledPosts()

	postListComp := templates.PostsTempl(Posts)

	return utils.Render(c, http.StatusOK, postListComp)
}

func DeletePost(id int) error {

	index, found := utils.FindPostIndexByID(id, Posts)
	if !found {
		return fmt.Errorf("could not find post for post id:%d", id)
	}

	// Remove the post from the slice
	Posts = append(Posts[:index], Posts[index+1:]...)

	if err := utils.SaveDataToFile(Posts, "data", "postStorage.json"); err != nil {
		return fmt.Errorf("unable to save to postStorage.json after delete operation.\nError:\n%s", err)
	}

	return nil
}

var checkerGoroutineRunning bool
var stopSignal chan string

func ManageScheduledPosts() {
	if stopSignal == nil {
		stopSignal = make(chan string)
	}
	if len(Posts) > 0 && !checkerGoroutineRunning {
		go scheduledPostsChecker(stopSignal)
		checkerGoroutineRunning = true

	} else if len(Posts) == 0 && checkerGoroutineRunning {
		stopSignal <- "stop" // Send stop signal
		checkerGoroutineRunning = false
		stopSignal = nil
	}
}

func scheduledPostsChecker(stop <-chan string) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			now := time.Now()
			for i := 0; i < len(Posts); {
				post := Posts[i]
				if now.After(post.ScheduleAt) {
					err := sendPostNotif(post)
					if err != nil {
						return
					}
					err = DeletePost(post.ID)
					if err != nil {
						return
					}
					PublishUpdate("post deleted")
					if len(Posts) == 0 {
						checkerGoroutineRunning = false
						return
					}
				} else {
					i++
				}
			}
		case <-stop: // Received a stop signal
			return // Exit the goroutine
		}
	}
}

func sendPostNotif(post models.Post) error {
	// Implement the functionality you want to execute for each post
	println("Executing post:", post.ID)
	return nil
}
