package routes

import (
	"marblegame/lobby"
	"marblegame/views"
	"strconv"

	"github.com/labstack/echo/v4"
)

func LobbyRoutes(e *echo.Echo) {
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
			lobbies[lobbyId] = lobby.NewLobby(lobbyId, strconv.Itoa(lobbyId))
			myLobby, _ = lobby.GetLobby(lobbyId)
		}

		// add user to lobby
		myLobby.AddPlayerToLobby(userToken.Value)

		return views.Lobby(myLobby).Render(c.Request().Context(), c.Response().Writer)
	})
	e.GET("/listOfLobbies", func(c echo.Context) error {
		return views.ListOfLobbies(lobby.GetLobbies()).Render(c.Request().Context(), c.Response().Writer)
	})
}
