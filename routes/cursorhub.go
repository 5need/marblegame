package routes

import (
	"encoding/json"
	"errors"
	"fmt"
	"marblegame/websockets"
	"time"

	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
)

func CursorHub(e *echo.Echo) {
	cursorHub := websockets.NewHub(
		nil,
		nil,
		cursorReadPumpHandler,
		50*time.Millisecond,
		cursorWritePumpHandler,
	)
	go cursorHub.Run()

	e.GET("/ws/cursor", func(c echo.Context) error {
		return websockets.ServeWS(cursorHub, c)
	})
}

func cursorReadPumpHandler(c *websockets.Client, message []byte) {
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

	c.Hub.Broadcast <- byteSlice
}

func cursorWritePumpHandler(c *websockets.Client, message []byte) error {
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
