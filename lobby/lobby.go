package lobby

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"marblegame/websockets"
	"net/http"
	"slices"
	"strconv"
	"strings"

	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
)

type Lobby struct {
	Id          string
	Name        string
	MaxPlayers  int
	PartyLeader string
	Players     []string
	Hub         *LobbyHub
}
type LobbyHub struct {
	websockets.Hub
	Lobby *Lobby
}

var lobbies = make(map[int]*Lobby, 0)

func GetLobbies() map[int]*Lobby {
	return lobbies
}

func GetLobby(lobbyId int) (*Lobby, error) {
	_, ok := lobbies[lobbyId]
	if !ok {
		return nil, errors.New("Lobby does not exist")
	}
	return lobbies[lobbyId], nil
}

func NewLobby(lobbyId int, name string) *Lobby {
	newLobby := &Lobby{
		Id:          strconv.Itoa(lobbyId),
		Name:        name,
		MaxPlayers:  2,
		PartyLeader: "",
		Players:     []string{},
		Hub:         nil,
	}
	newLobby.Hub = NewLobbyHub(newLobby)

	lobbies[lobbyId] = newLobby

	go newLobby.Hub.Run()

	return newLobby
}

func NewLobbyHub(lobby *Lobby) *LobbyHub {
	baseHub := websockets.NewHub2()

	lobbyHub := LobbyHub{
		Hub:   *baseHub,
		Lobby: lobby,
	}

	lobbyHub.RegisterHandler = func(c *websockets.Client) {
		// buffer := bytes.Buffer{}
		// CurrentLobby(lobbyHub.Lobby).Render(context.Background(), &buffer)
		fmt.Println("ADDED USER")

		// c.Hub.Broadcast <- buffer.Bytes()
	}

	lobbyHub.UnregisterHandler = func(c *websockets.Client) {
		// buffer := bytes.Buffer{}
		// CurrentLobby(lobbyHub.Lobby).Render(context.Background(), &buffer)
		// lobbyHub.Broadcast <- buffer.Bytes()
	}

	lobbyHub.ReadPumpHandler = func(c *websockets.Client, message []byte) {
		var r struct {
			Message string `json:"message"`
		}

		err := json.Unmarshal(message, &r)
		if err != nil {
			fmt.Println("couldn't unmarshal :-(")
			return
		}
		if len(r.Message) == 0 {
			return
		}
		fmt.Println(r)

		isCommand := strings.HasPrefix(r.Message, "/")
		if !isCommand {
			buffer := bytes.Buffer{}
			ChatboxResponse(r.Message, c.UserToken).Render(context.Background(), &buffer)
			c.Hub.Broadcast <- buffer.Bytes()
		} else {
			switch r.Message {
			case "/disband":
				buffer := bytes.Buffer{}
				ChatboxResponse("Lobby Leader disbanded the lobby", c.UserToken).Render(context.Background(), &buffer)
				ReturnToLobbyViewerResponse().Render(context.Background(), &buffer)
				lobbyHub.Broadcast <- buffer.Bytes()
				lobbyHub.CloseAllConnections()
				lobbyId, _ := strconv.Atoi(lobby.Id)
				delete(lobbies, lobbyId)
			}
		}
	}

	lobbyHub.WritePumpHandler = func(c *websockets.Client, message []byte) error {
		w, err := c.Conn.NextWriter(websocket.TextMessage)
		if err != nil {
			return err
		}

		n := len(c.Send)
		for range n {
			message = append(message, []byte{'\n'}...)
			message = append(message, <-c.Send...)
		}

		w.Write(message)

		if err := w.Close(); err != nil {
			return err
		}

		return nil
	}

	lobbyHub.ReadPumpDebounceDuration = 0

	return &lobbyHub
}

func (lobby *Lobby) AddPlayerToLobby(userToken string) error {
	alreadyInLobby := slices.Contains(lobby.Players, userToken)
	if alreadyInLobby {
		return nil
	}

	atMaxPlayers := len(lobby.Players) >= lobby.MaxPlayers

	if atMaxPlayers {
		return errors.New("Player count already at max")
	} else {
		lobby.Players = append(lobby.Players, userToken)

		buffer := bytes.Buffer{}
		CurrentLobby(lobby).Render(context.Background(), &buffer)
		lobby.Hub.Broadcast <- buffer.Bytes()
		return nil
	}
}

func (lobby *Lobby) RemovePlayerFromLobby(userToken string) error {
	inLobby := slices.Contains(lobby.Players, userToken)
	if !inLobby {
		return errors.New("Player already removed or not in lobby")
	}

	updatedPlayerList := []string{}
	for _, player := range lobby.Players {
		if player != userToken {
			updatedPlayerList = append(updatedPlayerList, player)
		}
	}

	lobby.Players = updatedPlayerList

	buffer := bytes.Buffer{}
	CurrentLobby(lobby).Render(context.Background(), &buffer)
	lobby.Hub.Broadcast <- buffer.Bytes()

	return nil
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

		return websockets.ServeWS(&myLobby.Hub.Hub, c)
	})
}
