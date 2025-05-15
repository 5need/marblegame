package lobby

import (
	"errors"
	"fmt"
	"marblegame/websockets"
	"net/http"
	"slices"
	"strconv"

	"github.com/labstack/echo/v4"
)

var lobbies = make(map[int]*LobbyHub, 0)

func GetLobbies() map[int]*LobbyHub {
	return lobbies
}

func GetLobby(lobbyId int) (*LobbyHub, error) {
	_, ok := lobbies[lobbyId]
	if !ok {
		return nil, errors.New("Lobby does not exist")
	}
	return lobbies[lobbyId], nil
}

func NewLobby(lobbyId int, name string) *LobbyHub {
	fmt.Println("NEW LOBBY")
	h := websockets.NewHub()
	newLobby := &LobbyHub{
		Hub:         h,
		Id:          strconv.Itoa(lobbyId),
		Name:        name,
		MaxPlayers:  2,
		PartyLeader: "",
		Players:     []string{},
	}

	lobbies[lobbyId] = newLobby

	go newLobby.Run()

	return newLobby
}

func LobbyRoutes(e *echo.Echo) {
	e.GET("/lobby", func(c echo.Context) error {
		userToken, _ := c.Cookie("userToken")
		// TODO: we're assuming a userToken here
		return PregameLobby(userToken.Value).Render(c.Request().Context(), c.Response().Writer)
	})

	e.GET("/lobby/:lobbyId", func(c echo.Context) error {
		userToken, _ := c.Cookie("userToken")
		lobbyId, _ := strconv.Atoi(c.Param("lobbyId"))

		myLobby, err := GetLobby(lobbyId)
		if err != nil {
			// lobby not created yet, add it
			lobbies[lobbyId] = NewLobby(lobbyId, strconv.Itoa(lobbyId))
			myLobby, _ = GetLobby(lobbyId)
		}

		// add user to lobby
		if err = myLobby.AddPlayerToLobby(userToken.Value); err != nil {
			fmt.Println(err)
			return c.String(http.StatusUnauthorized, "Lobby is full or you're not allowed in")
		}

		return LobbyView(myLobby, userToken.Value).Render(c.Request().Context(), c.Response().Writer)
	})
	e.GET("/listOfLobbies", func(c echo.Context) error {
		userToken, _ := c.Cookie("userToken")
		return ListOfLobbies(GetLobbies(), userToken.Value).Render(c.Request().Context(), c.Response().Writer)
	})
	e.GET("/ws/lobby/:lobbyId", func(c echo.Context) error {
		// find the lobbyHub, then serve it there
		fmt.Println(lobbies)
		lobbyId, _ := strconv.Atoi(c.Param("lobbyId"))
		myLobby, _ := GetLobby(lobbyId)

		fmt.Println(myLobby.Players)

		if ok := slices.Contains(myLobby.Players, c.QueryParam("userToken")); !ok {
			return c.String(http.StatusUnauthorized, "You're not allowed in this lobby")
		}
		fmt.Println("TRYING TO WHATEVER")
		fmt.Println(myLobby)

		return myLobby.ServeWS(c)
	})
}
