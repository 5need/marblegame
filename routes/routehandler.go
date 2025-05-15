package routes

import (
	"marblegame/engine"
	"marblegame/lobby"
	"marblegame/views"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

var marbleGame = engine.NewMarbleGame()

func MarbleGameRouteHandler(e *echo.Echo) {
	cursorHub := CursorHub{}
	go cursorHub.Run()

	e.GET("/ws/cursor", func(c echo.Context) error {
		return cursorHub.ServeWS(c)
	})

	gameHub := GameHub{}
	go gameHub.Run()
	e.GET("/ws/game", func(c echo.Context) error {
		return gameHub.ServeWS(c)
	})

	lobby.LobbyRoutes(e)

	e.GET("/", func(c echo.Context) error {
		userToken, err := c.Cookie("userToken")
		if err != nil {
			// client does not have a userToken cookie, make one for them
			// TODO: refresh or re-add cookie regardless
			cookie := new(http.Cookie)
			cookie.Name = "userToken"
			cookie.Value = uuid.New().String()
			cookie.Expires = time.Now().Add(1000 * 24 * time.Hour)
			c.SetCookie(cookie)
			return views.MarbleGame(cookie.Value).Render(c.Request().Context(), c.Response().Writer)
		}
		return views.MarbleGame(userToken.Value).Render(c.Request().Context(), c.Response().Writer)
	})
}
