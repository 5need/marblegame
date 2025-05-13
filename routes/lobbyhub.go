package routes

import (
	"bytes"
	"context"
	"fmt"
	"marblegame/views"
	"marblegame/websockets"

	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
)

func LobbyHub(e *echo.Echo) {
	lobbyHub := websockets.NewHub2()

	lobbyHub.RegisterHandler = func(c *websockets.Client) {
		buffer := bytes.Buffer{}
		views.ChatboxResponse(c.UserToken[:4]+" joined chat", "Server").Render(context.Background(), &buffer)
		c.Hub.Broadcast <- buffer.Bytes()
	}

	lobbyHub.UnregisterHandler = func(c *websockets.Client) {
		buffer := bytes.Buffer{}
		views.ChatboxResponse(c.UserToken[:4]+" left chat", "Server").Render(context.Background(), &buffer)
		lobbyHub.Broadcast <- buffer.Bytes()
	}

	lobbyHub.ReadPumpHandler = func(c *websockets.Client, message []byte) {
		// var r struct {
		// 	Message string `json:"message"`
		// }
		//
		// err := json.Unmarshal(message, &r)
		// if err != nil {
		// 	fmt.Println("couldn't unmarshal :-(")
		// 	return
		// }
		// if len(r.Message) == 0 {
		// 	return
		// }
		// fmt.Println(r)
		//
		// buffer := bytes.Buffer{}
		// views.ChatboxResponse(r.Message, c.userToken).Render(context.Background(), &buffer)
		//
		// c.hub.broadcast <- buffer.Bytes()
	}

	lobbyHub.WritePumpHandler = func(c *websockets.Client, message []byte) error {
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

	lobbyHub.ReadPumpDebounceDuration = 0

	go lobbyHub.Run()

	e.GET("/ws/lobby/:lobbyId", func(c echo.Context) error {
		return websockets.ServeWS(lobbyHub, c)
	})
	e.GET("/ws/lobby", func(c echo.Context) error {
		return websockets.ServeWS(lobbyHub, c)
	})
}
