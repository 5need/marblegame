package lobby

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"marblegame/websockets"
	"slices"
	"strconv"
	"strings"

	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
)

type LobbyHub struct {
	*websockets.Hub
	Id          string
	Name        string
	MaxPlayers  int
	PartyLeader string
	Players     []string
}

var _ websockets.HubInterface = (*LobbyHub)(nil)

func (lh *LobbyHub) RegisterHandler(c *websockets.Client) {
	// buffer := bytes.Buffer{}
	// CurrentLobby(lobbyHub.Lobby).Render(context.Background(), &buffer)
	fmt.Println("ADDED USER")

	// c.Hub.Broadcast <- buffer.Bytes()
}

func (lh *LobbyHub) UnregisterHandler(c *websockets.Client) {
	// buffer := bytes.Buffer{}
	// CurrentLobby(lobbyHub.Lobby).Render(context.Background(), &buffer)
	// lobbyHub.Broadcast <- buffer.Bytes()
}

func (lh *LobbyHub) ReadPumpHandler(c *websockets.Client, message []byte) {
	fmt.Println(string(message))
	var msg struct {
		Message string `json:"message"`
	}

	err := json.Unmarshal(message, &msg)
	if err != nil {
		fmt.Println("couldn't unmarshal :-(")
		return
	}
	if len(msg.Message) == 0 {
		return
	}
	fmt.Println(msg)

	isCommand := strings.HasPrefix(msg.Message, "/")
	if !isCommand {
		buffer := bytes.Buffer{}
		ChatboxResponse(msg.Message, c.UserToken).Render(context.Background(), &buffer)
		c.Hub.BroadcastChan() <- buffer.Bytes()
	} else {
		switch msg.Message {
		case "/disband":
			buffer := bytes.Buffer{}
			ChatboxResponse("Lobby Leader disbanded the lobby", c.UserToken).Render(context.Background(), &buffer)
			ReturnToLobbyViewerResponse().Render(context.Background(), &buffer)
			lh.Broadcast <- buffer.Bytes()
			lh.CloseAllConnections()
			lobbyId, _ := strconv.Atoi(lh.Id)
			delete(lobbies, lobbyId)
		}
	}
}

func (lh *LobbyHub) WritePumpHandler(c *websockets.Client, message []byte) error {
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

func (lobby *LobbyHub) AddPlayerToLobby(userToken string) error {
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

func (lobby *LobbyHub) RemovePlayerFromLobby(userToken string) error {
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

func (l *LobbyHub) ServeWS(c echo.Context) error {
	upgrader := websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024, // maybe change this to be larger since going to send tons of game frame data
	}

	userToken := c.QueryParam("userToken")
	if userToken == "" {
		return errors.New("no userToken")
	}

	conn, err := upgrader.Upgrade(c.Response(), c.Request(), nil)
	if err != nil {
		log.Println(err)
		return err
	}
	client := &websockets.Client{
		Hub:       l,
		UserToken: userToken,
		Conn:      conn,
		Send:      make(chan []byte, 256),
	}

	l.Register <- client

	go client.WritePump()
	go client.ReadPump()

	l.RegisterHandler(client)

	return nil
}
