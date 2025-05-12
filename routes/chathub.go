package routes

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"marblegame/views"

	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
)

func ChatHub(e *echo.Echo) {
	chatHub := newHub(
		func(c *Client) {
			buffer := bytes.Buffer{}
			views.ChatboxResponse(c.userToken[:4]+" joined chat", "Server").Render(context.Background(), &buffer)
			c.hub.broadcast <- buffer.Bytes()
		},
		nil,
		chatReadPumpHandler,
		0, // instant chat, no debouncing
		chatWritePumpHandler,
	)
	chatHub.unregisterHandler = func(c *Client) {
		buffer := bytes.Buffer{}
		views.ChatboxResponse(c.userToken[:4]+" left chat", "Server").Render(context.Background(), &buffer)
		chatHub.broadcast <- buffer.Bytes()
	}
	go chatHub.run()

	e.GET("/marblegame/ws/chat", func(c echo.Context) error {
		return serveWS(chatHub, c)
	})
}

func chatReadPumpHandler(c *Client, message []byte) {
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
	views.ChatboxResponse(r.Message, c.userToken).Render(context.Background(), &buffer)

	c.hub.broadcast <- buffer.Bytes()
}

func chatWritePumpHandler(c *Client, message []byte) error {
	w, err := c.conn.NextWriter(websocket.TextMessage)
	if err != nil {
		return err
	}

	n := len(c.send)
	for i := 0; i < n; i++ {
		message = append(message, newline...)
		message = append(message, <-c.send...)
	}

	w.Write(message)

	if err := w.Close(); err != nil {
		return err
	}

	return nil
}
