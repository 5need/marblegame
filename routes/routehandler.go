package routes

import (
	"marblegame/engine"
	"marblegame/lobby"
	"marblegame/views"
	"net/http"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

var marbleGame = engine.NewMarbleGame()
var lobbies = map[int]*lobby.Lobby{}

func MarbleGameRouteHandler(e *echo.Echo) {
	ChatHub(e)
	CursorHub(e)
	GameHub(e)

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

	e.GET("/lobby", func(c echo.Context) error {
		userToken, _ := c.Cookie("userToken")
		// TODO: we're assuming a userToken here
		return views.PregameLobby(userToken.Value).Render(c.Request().Context(), c.Response().Writer)
	})

	e.GET("/lobby/:lobbyId", func(c echo.Context) error {
		userToken, _ := c.Cookie("userToken")
		lobbyId, _ := strconv.Atoi(c.Param("lobbyId"))

		myLobby, err := lobby.GetLobby(lobbyId)
		if err != nil {
			// lobby not created yet, add it
			lobbies[lobbyId] = lobby.NewLobby(lobbyId)
			myLobby, _ = lobby.GetLobby(lobbyId)
		}

		// add user to lobby
		myLobby.AddPlayer(userToken.Value)

		return views.Lobby(myLobby).Render(c.Request().Context(), c.Response().Writer)
	})
	e.GET("/listOfLobbies", func(c echo.Context) error {
		return views.ListOfLobbies(lobby.GetLobbies()).Render(c.Request().Context(), c.Response().Writer)
	})
}
