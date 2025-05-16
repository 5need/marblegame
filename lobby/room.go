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

type Room struct {
	*websockets.Hub
	Id          string
	Name        string
	MaxPlayers  int
	PartyLeader string
	Players     []string
}

var _ websockets.HubInterface = (*Room)(nil)

func (lh *Room) RegisterHandler(c *websockets.Client) {
}

func (lh *Room) UnregisterHandler(c *websockets.Client) {
}

func (lh *Room) ReadPumpHandler(c *websockets.Client, message []byte) {
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

	isCommand := strings.HasPrefix(msg.Message, "/")
	if !isCommand {
		buffer := bytes.Buffer{}
		ChatboxResponse(msg.Message, c.UserToken).Render(context.Background(), &buffer)

		c.Hub.BroadcastChan() <- buffer.Bytes()
	} else {
		switch msg.Message {
		case "/disband":
			buffer := bytes.Buffer{}
			ChatboxResponse("Room Leader disbanded the room", c.UserToken).Render(context.Background(), &buffer)
			ReturnToLobbyResponse().Render(context.Background(), &buffer)
			lh.Broadcast <- buffer.Bytes()
			lh.CloseAllConnections()
			roomId, _ := strconv.Atoi(lh.Id)
			delete(rooms, roomId)
		case "/disconnect":
			// send out a chat message to everyone
			buffer := bytes.Buffer{}
			ChatboxResponse("Player left", c.UserToken).Render(context.Background(), &buffer)
			lh.Broadcast <- buffer.Bytes()

			// specifically return the dc'd player to the lobby
			buffer = bytes.Buffer{}
			ReturnToLobbyResponse().Render(context.Background(), &buffer)
			c.Send <- buffer.Bytes()

			lh.RemovePlayerFromRoom(c.UserToken)
		}
	}
}

func (lh *Room) WritePumpHandler(c *websockets.Client, message []byte) error {
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

func (room *Room) AddPlayerToRoom(userToken string) error {
	alreadyInRoom := slices.Contains(room.Players, userToken)
	if alreadyInRoom {
		return nil
	}

	atMaxPlayers := len(room.Players) >= room.MaxPlayers

	if atMaxPlayers {
		return errors.New("Player count already at max")
	} else {
		room.Players = append(room.Players, userToken)

		buffer := bytes.Buffer{}
		CurrentRoom(room).Render(context.Background(), &buffer)
		room.Hub.Broadcast <- buffer.Bytes()
		return nil
	}
}

func (room *Room) RemovePlayerFromRoom(userToken string) error {
	inRoom := slices.Contains(room.Players, userToken)
	if !inRoom {
		return errors.New("Player already removed or not in room")
	}

	updatedPlayerList := []string{}
	for _, player := range room.Players {
		if player != userToken {
			updatedPlayerList = append(updatedPlayerList, player)
		}
	}

	room.Players = updatedPlayerList

	buffer := bytes.Buffer{}
	CurrentRoom(room).Render(context.Background(), &buffer)
	room.Hub.Broadcast <- buffer.Bytes()

	return nil
}

func (l *Room) ServeWS(c echo.Context) error {
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
