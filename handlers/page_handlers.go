package handlers

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"zenoforge.com/goLiveNotif/templates"
	"zenoforge.com/goLiveNotif/utils"
)

func MainPage(c echo.Context) error {
	postListComp := templates.PostsTempl(Posts)
	return utils.Render(c, http.StatusOK, templates.BasePage(postListComp))
}
