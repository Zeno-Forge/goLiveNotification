package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"zenoforge.com/goLiveNotif/log"
	"zenoforge.com/goLiveNotif/models"
	"zenoforge.com/goLiveNotif/templates"
	"zenoforge.com/goLiveNotif/utils"
)

var Posts []models.Post

func CreatePostHandler(c echo.Context) error {
	var appConfig models.AppConfig
	err := utils.LoadDataFromFile(&appConfig, "data", "app.config.json")

	if err != nil && os.IsNotExist(err) {
		appConfig = createDefConf()
		if err := utils.SaveDataToFile(appConfig, "data", "app.config.json"); err != nil {
			log.Error(err.Error())
			return err
		}

	} else if err != nil {
		log.Error(err.Error())
		return err
	}

	var newPost = models.Post{
		ScheduleAt: time.Now(),
		Message:    appConfig.Settings.PostTemplate.Message,
	}

	// Assign a new unique ID
	newPost.ID = uuid.New().String()

	postModal := templates.PostModal(newPost, "Create Post")

	return utils.Render(c, http.StatusOK, postModal)
}

func GetPostList(c echo.Context) error {
	sortPosts()

	postListComp := templates.PostsTempl(Posts)

	return utils.Render(c, http.StatusOK, postListComp)
}

func GetPostHandler(c echo.Context) error {
	id := c.Param("id")

	index, found := utils.FindPostIndexByID(id, Posts)
	if !found {
		return c.JSON(http.StatusNotFound, map[string]string{"message": "Post not found"})
	}

	postModal := templates.PostModal(Posts[index], "Edit Post")

	return utils.Render(c, http.StatusOK, postModal)
}

func EditPostHandler(c echo.Context) error {
	id := c.Param("id")

	index, found := utils.FindPostIndexByID(id, Posts)
	if !found {

		var appConfig models.AppConfig

		err := utils.LoadDataFromFile(&appConfig, "data", "app.config.json")
		if err != nil && os.IsNotExist(err) {
			appConfig = createDefConf()
			if err := utils.SaveDataToFile(appConfig, "data", "app.config.json"); err != nil {
				log.Error(err.Error())
				return err
			}
		}

		var newPost models.Post
		newPost.ID = id
		newPost.Message = appConfig.Settings.PostTemplate.Message

		err = updatePost(c, &newPost)
		if err != nil {
			log.Error(err.Error())
			return err
		}
		Posts = append(Posts, newPost)
	} else {
		updateP := Posts[index]
		err := updatePost(c, &updateP)
		if err != nil {
			log.Error(err.Error())
			return err
		}

		Posts[index] = updateP
	}

	if err := utils.SaveDataToFile(Posts, "data", "postStorage.json"); err != nil {
		return err
	}

	ManageScheduledPosts()

	sortPosts()

	postListComp := templates.PostsTempl(Posts)

	return utils.Render(c, http.StatusOK, postListComp)
}

func GetTemplateHandler(c echo.Context) error {

	var appConfig models.AppConfig

	err := utils.LoadDataFromFile(&appConfig, "data", "app.config.json")
	if err != nil && os.IsNotExist(err) {
		appConfig = createDefConf()
		if err := utils.SaveDataToFile(appConfig, "data", "app.config.json"); err != nil {
			log.Error(err.Error())
			return err
		}

	} else if err != nil {
		log.Error(err.Error())
		return err
	}

	var postTemp = models.Post{
		Template: true,
		Message:  appConfig.Settings.PostTemplate.Message,
	}

	postModal := templates.PostModal(postTemp, "Edit Template")

	return utils.Render(c, http.StatusOK, postModal)
}

func EditTemplateHandler(c echo.Context) error {

	var appConfig models.AppConfig

	err := utils.LoadDataFromFile(&appConfig, "data", "app.config.json")
	if err != nil && os.IsNotExist(err) {
		appConfig = createDefConf()
		if err := utils.SaveDataToFile(appConfig, "data", "app.config.json"); err != nil {
			log.Error(err.Error())
			return err
		}

	} else if err != nil {
		log.Error(err.Error())
		return err
	}

	var postTemplate models.Post

	postTemplate.Message = appConfig.Settings.PostTemplate.Message

	err = updatePost(c, &postTemplate)
	if err != nil {
		log.Error(err.Error())
		return err
	}

	appConfig.Settings.PostTemplate.Message = postTemplate.Message

	if err := utils.SaveDataToFile(appConfig, "data", "app.config.json"); err != nil {
		return err
	}

	return c.NoContent(http.StatusOK)
}

func DeletePostHandler(c echo.Context) error {
	id := c.Param("id")

	err := DeletePost(id)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"message": "Unable to delete post"})
	}

	ManageScheduledPosts()

	sortPosts()

	postListComp := templates.PostsTempl(Posts)

	return utils.Render(c, http.StatusOK, postListComp)
}

func DeletePost(id string) error {

	index, found := utils.FindPostIndexByID(id, Posts)
	if !found {
		return fmt.Errorf("could not find post for post id:%s", id)
	}

	// Remove the post from the slice
	Posts = append(Posts[:index], Posts[index+1:]...)

	if err := utils.SaveDataToFile(Posts, "data", "postStorage.json"); err != nil {
		return fmt.Errorf("unable to save to postStorage.json after delete operation.\nError:\n%s", err)
	}

	// cleanup file uploads
	err := os.RemoveAll(fmt.Sprintf("./data/uploads/%s", id))
	if err != nil {
		log.Error("Failed to remove directory from removed post: %v", err)
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
		log.Info("Scheduler Started")

	} else if len(Posts) == 0 && checkerGoroutineRunning {
		stopSignal <- "stop" // Send stop signal
		checkerGoroutineRunning = false
		stopSignal = nil
		log.Info("Scheduler Stopped")
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
						log.Error(fmt.Sprintf("Error in deleting post after schedule send:\n%s", err.Error()))
						checkerGoroutineRunning = false
						log.Info("Scheduler Stopped")
						return
					}
					err = DeletePost(post.ID)
					if err != nil {
						log.Error(fmt.Sprintf("Error in deleting post after schedule send:\n%s", err.Error()))
						checkerGoroutineRunning = false
						log.Info("Scheduler Stopped")
						return
					}
					PublishUpdate("post deleted")
					if len(Posts) == 0 {
						checkerGoroutineRunning = false
						log.Info("Scheduler Stopped")
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
	var buffer bytes.Buffer
	multipartWriter := multipart.NewWriter(&buffer)

	// Create a part for the image file
	imagePath := post.Message.Embed[0].Image.URL
	if imagePath != "" {
		file, err := os.Open(imagePath)
		if err != nil {
			log.Error(fmt.Sprintf("Error in opening image for discord message:\n%s", err.Error()))
			return err
		}
		defer file.Close()

		filePart, err := multipartWriter.CreateFormFile("files", filepath.Base(file.Name()))
		if err != nil {
			log.Error(fmt.Sprintf("Error in creating file part for discord request:\n%s", err.Error()))
			return err
		}
		_, err = io.Copy(filePart, file)
		if err != nil {
			log.Error(fmt.Sprintf("Error in copying image file to discord message:\n%s", err.Error()))
			return err
		}

		post.Message.Embed[0].Image.URL = fmt.Sprintf("attachment://%s", filepath.Base(file.Name()))
	}

	// Create a part for the thumbnail file
	thumbPath := post.Message.Embed[0].Thumbnail.URL
	if thumbPath != "" {
		thumbFile, err := os.Open(thumbPath)
		if err != nil {
			log.Error(fmt.Sprintf("Error in opening image for discord message:\n%s", err.Error()))
			return err
		}
		defer thumbFile.Close()

		thumbPart, err := multipartWriter.CreateFormFile("files", filepath.Base(thumbFile.Name()))
		if err != nil {
			log.Error(fmt.Sprintf("Error in creating file part for discord request:\n%s", err.Error()))
			return err
		}
		_, err = io.Copy(thumbPart, thumbFile)
		if err != nil {
			log.Error(fmt.Sprintf("Error in copying image file to discord message:\n%s", err.Error()))
			return err
		}

		post.Message.Embed[0].Thumbnail.URL = fmt.Sprintf("attachment://%s", filepath.Base(thumbFile.Name()))
	}

	// Create a part for the JSON payload
	part, err := multipartWriter.CreatePart(
		textproto.MIMEHeader{
			"Content-Disposition": []string{`form-data; name="payload_json"`},
			"Content-Type":        []string{"application/json"},
		},
	)
	if err != nil {
		log.Error(fmt.Sprintf("Error in creating part in multipart request for discord message:\n%s", err.Error()))
		return err
	}

	// Encode and write the JSON payload to the part
	err = json.NewEncoder(part).Encode(post.Message)
	if err != nil {
		log.Error(fmt.Sprintf("Error in encoding discord message before sending:\n%s", err.Error()))
		return err
	}

	jsonPayload, _ := json.Marshal(post.Message)
	log.Info(string(jsonPayload))

	multipartWriter.Close()

	var appConfig models.AppConfig
	err = utils.LoadDataFromFile(&appConfig, "data", "app.config.json")
	if err != nil && os.IsNotExist(err) {
		appConfig = createDefConf()
		if err := utils.SaveDataToFile(appConfig, "data", "app.config.json"); err != nil {
			log.Error(err.Error())
			return err
		}

	} else if err != nil {
		log.Error(err.Error())
		return err
	}

	webhookURL := appConfig.Settings.DiscordWebhook

	req, err := http.NewRequest("POST", webhookURL, &buffer)
	if err != nil {
		log.Error(fmt.Sprintf("Error in creating discord post request:\n%s", err.Error()))
		return err
	}
	req.Header.Set("Content-Type", multipartWriter.FormDataContentType())

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Error(fmt.Sprintf("Error in sending discord post request:\n%s", err.Error()))
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		bodyBytes, _ := io.ReadAll(resp.Body)
		log.Error(fmt.Sprintf("discord webhook error: %s", string(bodyBytes)))
		return fmt.Errorf("discord webhook error: %s", string(bodyBytes))
	}

	return nil
}

func sortPosts() {
	slices.SortFunc(Posts, func(a, b models.Post) int {
		if a.ScheduleAt.Before(b.ScheduleAt) {
			return -1
		} else if a.ScheduleAt.After(b.ScheduleAt) {
			return 1
		}
		return 0
	})
}

func updatePost(c echo.Context, post *models.Post) error {
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

	if c.Param("id") != "" {
		post.ID = c.Param("id")
	}

	oldFileURL := ""
	if post.Message.Embed != nil && len(post.Message.Embed) > 0 {
		oldFileURL = post.Message.Embed[0].Image.URL
	}
	newFileURL := c.FormValue("imageURLInput")
	if oldFileURL != newFileURL && oldFileURL != "" {
		err := os.Remove(oldFileURL)
		if err != nil {
			log.Error("Failed to remove file old file: %v", err)
		}
	}

	oldThumbURL := ""
	if post.Message.Embed != nil && len(post.Message.Embed) > 0 {
		oldFileURL = post.Message.Embed[0].Thumbnail.URL
	}
	newThumbURL := c.FormValue("thumbnailInput")
	if oldThumbURL != newThumbURL && oldThumbURL != "" {
		err := os.Remove(oldFileURL)
		if err != nil {
			log.Error("Failed to remove file old file: %v", err)
		}
	}

	post.ScheduleAt = schedule
	post.Message = models.DiscordMessage{
		Content: c.FormValue("contentInput"),
		Embed: []models.Embed{
			{
				Title:       c.FormValue("titleInput"),
				Description: c.FormValue("descriptionInput"),
				Image:       models.URL{URL: post.Message.Embed[0].Image.URL},
				URL:         c.FormValue("urlInput"),
				Color:       color,
				Thumbnail:   models.URL{URL: post.Message.Embed[0].Thumbnail.URL},
				Footer: models.Footer{
					IconURL: c.FormValue("footerIconInput"),
					Text:    c.FormValue("footerTextInput"),
				},
			},
		},
	}

	dirPath := "./data/uploads/" + c.Param("id")
	if c.Param("id") == "" {
		dirPath = "./data/uploads/template"
	}

	file, err := c.FormFile("imageUpload")
	if err == nil {
		updateFile(file, dirPath)
		post.Message.Embed[0].Image.URL = dirPath + "/" + file.Filename
	}

	thumbImg, err := c.FormFile("thumbImgUpload")
	if err == nil {
		updateFile(thumbImg, dirPath)
		post.Message.Embed[0].Thumbnail.URL = dirPath + "/" + thumbImg.Filename
	}

	return nil
}

func updateFile(file *multipart.FileHeader, dirPath string) error {
	src, err := file.Open()

	if err != nil {
		return err
	}

	defer src.Close()

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

	return nil
}

func createDefConf() models.AppConfig {
	var appConfig = models.AppConfig{
		Name:    "goLiveNotif",
		Version: "0.1.2",
		Settings: models.Settings{
			Theme:          "Light",
			DiscordWebhook: "Update Me!",
			PostTemplate: models.PostTemplate{
				Message: models.DiscordMessage{
					Embed: []models.Embed{
						{
							Title:       "",
							Description: "",
							Image:       models.URL{URL: ""},
							URL:         "",
							Color:       0,
							Thumbnail:   models.URL{URL: ""},
							Footer: models.Footer{
								IconURL: "",
								Text:    "",
							},
						},
					},
				},
			},
		},
	}

	return appConfig
}
