package handlers

import (
	"fmt"
	"net/http"
	"os"
	"sync"

	"github.com/labstack/echo/v4"
	"zenoforge.com/goLiveNotif/log"
	"zenoforge.com/goLiveNotif/models"
	"zenoforge.com/goLiveNotif/templates"
	"zenoforge.com/goLiveNotif/utils"
)

func MainPageHandler(c echo.Context) error {

	ManageScheduledPosts()

	sortPosts()
	postListComp := templates.PostsTempl(Posts)

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

	return utils.Render(c, http.StatusOK, templates.BasePage(postListComp, appConfig))
}

func WebhookHandler(c echo.Context) error {
	var appConfig models.AppConfig

	err := utils.LoadDataFromFile(&appConfig, "data", "app.config.json")
	if err != nil {
		log.Error(err.Error())
		return err
	}

	appConfig.Settings.DiscordWebhook = c.FormValue("discordWebhook")

	err = utils.SaveDataToFile(&appConfig, "data", "app.config.json")
	if err != nil {
		log.Error(err.Error())
		return err
	}

	return c.NoContent(http.StatusOK)
}

// Structure to manage user channels
type UserChannels struct {
	channels []chan string
	mu       sync.Mutex
}

var userChannels = &UserChannels{
	channels: make([]chan string, 0),
}

func PublishUpdate(update string) {
	userChannels.mu.Lock()
	defer userChannels.mu.Unlock()

	// Broadcast the update to all user channels
	for _, ch := range userChannels.channels {
		ch <- update
	}
}

func EventsHandler(c echo.Context) error {
	flusher, ok := c.Response().Writer.(http.Flusher)
	if !ok {
		return c.JSON(http.StatusInternalServerError, map[string]string{"message": "Streaming unsupported"})
	}

	c.Response().Header().Add("Content-Type", "text/event-stream")
	c.Response().Header().Add("Cache-Control", "no-cache")
	c.Response().Header().Add("Connection", "keep-alive")

	// Create a new channel for this connection
	userChannel := make(chan string)
	userChannels.mu.Lock()
	userChannels.channels = append(userChannels.channels, userChannel)
	userChannels.mu.Unlock()

	defer func() {
		// Clean up the channel when the connection is closed
		userChannels.mu.Lock()
		for i, ch := range userChannels.channels {
			if ch == userChannel {
				userChannels.channels = append(userChannels.channels[:i], userChannels.channels[i+1:]...)
				close(userChannel)
				break
			}
		}
		userChannels.mu.Unlock()
	}()

	for {
		select {
		case <-c.Request().Context().Done():
			// Client disconnected; exit handler
			return nil
		case update := <-userChannel:
			// Send updates to the client
			_, err := fmt.Fprintf(c.Response().Writer, "data: %s\n\n", update)
			if err != nil {
				// Handle the error, e.g., log it or perform cleanup
				log.Error(err.Error())
				return nil
			}
			flusher.Flush()
		}
	}
}
