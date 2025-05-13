package routes

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"marblegame/views"
	"marblegame/websockets"

	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
)

func ChatHub(e *echo.Echo) {
	chatHub := websockets.NewHub(
		func(c *websockets.Client) {
			buffer := bytes.Buffer{}
			views.ChatboxResponse(c.UserToken[:4]+" joined chat", "Server").Render(context.Background(), &buffer)
			c.Hub.Broadcast <- buffer.Bytes()
		},
		nil,
		chatReadPumpHandler,
		0, // instant chat, no debouncing
		chatWritePumpHandler,
	)
	chatHub.UnregisterHandler = func(c *websockets.Client) {
		buffer := bytes.Buffer{}
		views.ChatboxResponse(c.UserToken[:4]+" left chat", "Server").Render(context.Background(), &buffer)
		chatHub.Broadcast <- buffer.Bytes()
	}
	go chatHub.Run()

	e.GET("/ws/chat", func(c echo.Context) error {
		return websockets.ServeWS(chatHub, c)
	})
}

func chatReadPumpHandler(c *websockets.Client, message []byte) {
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

	buffer := bytes.Buffer{}
	views.ChatboxResponse(r.Message, c.UserToken).Render(context.Background(), &buffer)

	c.Hub.Broadcast <- buffer.Bytes()
}

func chatWritePumpHandler(c *websockets.Client, message []byte) error {
	w, err := c.Conn.NextWriter(websocket.TextMessage)
	if err != nil {
		return err
	}

	n := len(c.Send)
	for i := 0; i < n; i++ {
		message = append(message, []byte{'\n'}...)
		message = append(message, <-c.Send...)
	}

	w.Write(message)

	if err := w.Close(); err != nil {
		return err
	}

	return nil
}
