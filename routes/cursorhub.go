package routes

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
)

func CursorHub(e *echo.Echo) {
	cursorHub := newHub(
		nil,
		nil,
		cursorReadPumpHandler,
		50*time.Millisecond,
		cursorWritePumpHandler,
	)
	go cursorHub.run()

	e.GET("/ws/cursor", func(c echo.Context) error {
		return serveWS(cursorHub, c)
	})
}

func cursorReadPumpHandler(c *Client, message []byte) {
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
	r.UserToken = c.userToken

	byteSlice, err := json.Marshal(r)
	if err != nil {
		fmt.Println("couldn't marshal :-(")
		return
	}

	c.hub.broadcast <- byteSlice
}

func cursorWritePumpHandler(c *Client, message []byte) error {
	n := len(c.send)
	for i := 0; i < n; i++ {
		// only grab the last message lmao
		message = <-c.send
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

	if r.UserToken != c.userToken {
		w, err := c.conn.NextWriter(websocket.TextMessage)
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
