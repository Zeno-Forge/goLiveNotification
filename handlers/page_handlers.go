package handlers

import (
	"fmt"
	"net/http"
	"os"

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

var updates = make(chan string)

func PublishUpdate(update string) {
	updates <- update
}

func EventsHandler(c echo.Context) error {
	flusher, ok := c.Response().Writer.(http.Flusher)
	if !ok {
		return c.JSON(http.StatusInternalServerError, map[string]string{"message": "Streaming unsupported"})
	}

	c.Response().Header().Add("Content-Type", "text/event-stream")
	c.Response().Header().Add("Cache-Control", "no-cache")
	c.Response().Header().Add("Connection", "keep-alive")

	for {
		select {
		case <-c.Request().Context().Done():
			// Client disconnected; exit handler
			return nil
		case update := <-updates:
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
