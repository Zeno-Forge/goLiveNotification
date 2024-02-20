package main

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/a-h/templ"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"zenoforge.com/goLiveNotif/post"
	"zenoforge.com/goLiveNotif/templates"
)

var posts []post.Post

func loadPostsFromFile(filename string) error {
	file, err := os.Open(filename)
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	if os.IsNotExist(err) {
		return nil
	}

	defer file.Close()

	decoder := json.NewDecoder(file)
	err = decoder.Decode(&posts)
	if err != nil {
		return err
	}

	return nil
}

func savePostsToFile(filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	err = encoder.Encode(posts)
	if err != nil {
		return err
	}

	return nil
}

func NewServer(logger *slog.Logger) *echo.Echo {
	e := echo.New()

	// Middleware
	e.Use(middleware.RequestLoggerWithConfig(middleware.RequestLoggerConfig{
		LogStatus:   true,
		LogURI:      true,
		LogError:    true,
		LogMethod:   true,
		HandleError: true, // forwards error to the global error handler, so it can decide appropriate status code
		LogValuesFunc: func(c echo.Context, v middleware.RequestLoggerValues) error {
			if v.Error == nil {
				logger.LogAttrs(context.Background(), slog.LevelInfo, "REQUEST",
					slog.Group(
						v.Method,
						slog.String("uri", v.URI),
						slog.Int("status", v.Status),
					),
				)
			} else {
				logger.LogAttrs(context.Background(), slog.LevelError, "REQUEST_ERROR",
					slog.Group(
						v.Method,
						slog.String("uri", v.URI),
						slog.Int("status", v.Status),
						slog.String("err", v.Error.Error()),
					),
				)
			}
			return nil
		},
	}))
	e.Use(middleware.Recover())

	// Static files
	e.Static("/uploads", "./uploads")

	// Routes
	e.GET("/", mainPage)
	e.GET("/post/:id", getPost)
	e.GET("/post/create", createPost)
	e.PUT("/post/:id", editPost)
	e.DELETE("/post/:id", deletePost)

	return e
}

// This custom Render replaces Echo's echo.Context.Render() with templ's templ.Component.Render().
func Render(ctx echo.Context, statusCode int, t templ.Component) error {
	ctx.Response().Writer.WriteHeader(statusCode)
	ctx.Response().Header().Set(echo.HeaderContentType, echo.MIMETextHTML)
	return t.Render(ctx.Request().Context(), ctx.Response().Writer)
}

func mainPage(c echo.Context) error {
	postListComp := templates.PostsTempl(posts)
	return Render(c, http.StatusOK, templates.BasePage(postListComp))
}

func createPost(c echo.Context) error {
	var newPost = post.Post{
		Message: post.DiscordMessage{
			Embed: []post.Embed{
				{
					Thumbnail: post.URL{URL: "https://static-cdn.jtvnw.net/jtv_user_pictures/f77022c4-2d3e-45ff-a7a9-22ada2688c50-profile_image-300x300.png"},
					URL:       "https://twitch.tv/marshievt",
					Color:     6504867,
					Footer: post.Footer{
						IconURL: "https://www.freepnglogos.com/uploads/twitch-logo-transparent-png-20.png",
						Text:    "Twitch",
					},
				},
			},
		},
	}

	// Assign a new unique ID
	newID := 1
	if len(posts) > 0 {
		newID = posts[len(posts)-1].ID + 1
	}
	newPost.ID = newID

	posts = append(posts, newPost)

	if err := savePostsToFile("posts.json"); err != nil {
		return err
	}

	postModal := templates.PostModal(newPost)

	return Render(c, http.StatusOK, postModal)
}

func getPost(c echo.Context) error {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"message": "Invalid post ID"})
	}

	index, found := findPostIndexByID(id)
	if !found {
		return c.JSON(http.StatusNotFound, map[string]string{"message": "Post not found"})
	}

	postModal := templates.PostModal(posts[index])

	return Render(c, http.StatusOK, postModal)
}

func editPost(c echo.Context) error {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"message": "Invalid post ID"})
	}

	index, found := findPostIndexByID(id)
	if !found {
		return c.JSON(http.StatusNotFound, map[string]string{"message": "Post not found"})
	}

	color, _ := strconv.Atoi(c.FormValue("colorInput"))
	schedule, _ := time.Parse("2006-01-02T15:04", c.FormValue("scheduleInput"))

	var updatedPost = post.Post{
		ID:         id,
		ScheduleAt: schedule,
		Message: post.DiscordMessage{
			Content: c.FormValue("contentInput"),
			Embed: []post.Embed{
				{
					Title:       c.FormValue("titleInput"),
					Description: c.FormValue("descriptionInput"),
					URL:         c.FormValue("urlInput"),
					Color:       color,
					Thumbnail:   post.URL{URL: c.FormValue("thumbnailInput")},
					Footer: post.Footer{
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
		dirPath := "./uploads/" + c.Param("id")
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

	posts[index] = updatedPost

	if err := savePostsToFile("posts.json"); err != nil {
		return err
	}

	postListComp := templates.PostsTempl(posts)

	return Render(c, http.StatusOK, postListComp)
}

func deletePost(c echo.Context) error {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"message": "Invalid post ID"})
	}

	index, found := findPostIndexByID(id)
	if !found {
		return c.JSON(http.StatusNotFound, map[string]string{"message": "Post not found"})
	}

	// Remove the post from the slice
	posts = append(posts[:index], posts[index+1:]...)

	if err := savePostsToFile("posts.json"); err != nil {
		return err
	}

	postListComp := templates.PostsTempl(posts)

	return Render(c, http.StatusOK, postListComp)
}

func findPostIndexByID(id int) (int, bool) {
	for index, post := range posts {
		if post.ID == id {
			return index, true
		}
	}
	return -1, false
}

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	// Load posts from file
	if err := loadPostsFromFile("posts.json"); err != nil && !os.IsNotExist(err) {
		logger.Log(context.Background(), slog.LevelError, "Failed to load posts", slog.String("Error:", err.Error()))
	}

	server := NewServer(logger)

	// Start server
	server.Logger.Fatal(server.Start(":8989"))
}
