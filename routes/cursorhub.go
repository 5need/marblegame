package routes

import (
	"encoding/json"
	"errors"
	"fmt"
	"marblegame/websockets"

	"github.com/gorilla/websocket"
)

type CursorHub struct {
	websockets.Hub
	Id          string
	Name        string
	MaxPlayers  int
	PartyLeader string
	Players     []string
}

var _ websockets.HubInterface = (*CursorHub)(nil)

func (ch *CursorHub) ReadPumpHandler(c *websockets.Client, message []byte) {
	var r struct {
		UserToken string `json:"userToken"`
		MouseX    string `json:"mouseX"`
		MouseY    string `json:"mouseY"`
	}

	err := json.Unmarshal(message, &r)
	if err != nil {
		fmt.Println("couldn't unmarshal :-(")
		return
	}
	// fmt.Println("cursor:", r)
	r.UserToken = c.UserToken

	byteSlice, err := json.Marshal(r)
	if err != nil {
		fmt.Println("couldn't marshal :-(")
		return
	}

	ch.Broadcast <- byteSlice
}

func (ch *CursorHub) WritePumpHandler(c *websockets.Client, message []byte) error {
	n := len(c.Send)
	for i := 0; i < n; i++ {
		// only grab the last message lmao
		message = <-c.Send
	}

	var r struct {
		UserToken string `json:"userToken"`
		MouseX    string `json:"mouseX"`
		MouseY    string `json:"mouseY"`
	}

	err := json.Unmarshal(message, &r)
	if err != nil {
		return errors.New("couldn't unmarshal :-(")
	}

	if r.UserToken != c.UserToken {
		w, err := c.Conn.NextWriter(websocket.TextMessage)
		if err != nil {
			return err
		}
		w.Write(message)
		if err := w.Close(); err != nil {
			return err
		}
	}

	return nil
}
