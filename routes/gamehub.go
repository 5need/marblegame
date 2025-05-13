package routes

import (
	"encoding/json"
	"fmt"
	"marblegame/engine"
	"math/rand"
	"time"

	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
)

func GameHub(e *echo.Echo) {
	gameHub := newHub(
		gameRegisterHandler,
		nil,
		gameReadPumpHandler,
		0,
		gameWritePumpHandler,
	)
	go gameHub.run()

	e.GET("/ws/game", func(c echo.Context) error {
		return serveWS(gameHub, c)
	})
}

func gameRegisterHandler(c *Client) {
	time.AfterFunc(500*time.Millisecond, // TODO: this is jank sauce
		func() {
			if _, exists := marbleGame.Players[c.userToken]; !exists {
				joiningPlayer := engine.Player{
					UserToken:         c.userToken,
					DisplayName:       c.userToken[:4],
					Score:             0,
					Hue:               int(rand.Int31n(256)),
					ShouldSkipMyTurns: false,
					TurnsTaken:        0,
					Inventory: []engine.MarbleType{
						engine.MarbleTypes[0],
						engine.MarbleTypes[0],
						engine.MarbleTypes[1],
						engine.MarbleTypes[0],
						engine.MarbleTypes[2],
						engine.MarbleTypes[1],
						engine.MarbleTypes[0],
						engine.MarbleTypes[0],
						engine.MarbleTypes[2],
						engine.MarbleTypes[0],
						engine.MarbleTypes[0],
					},
				}
				marbleGame.Players[c.userToken] = &joiningPlayer
				marbleGame.TurnOrder = append(marbleGame.TurnOrder, &joiningPlayer)
			}

			sendMarbleGameToClient(c, marbleGame)
		},
	)
}

type ActionRequest struct {
	ActionString string `json:"action"` // stringified input cause lazy
}

func gameReadPumpHandler(c *Client, message []byte) {
	// so when we read this from the ws we need to do some things

	// 1. check if it's the player's turn

	// 2. process their action
	// we first have to extract out the stringified action cause of how the frontend is
	var r ActionRequest
	err := json.Unmarshal(message, &r)
	if err != nil {
		fmt.Println(err)
		return
	}
	var a engine.Action
	err = json.Unmarshal([]byte(r.ActionString), &a)
	fmt.Println(a)
	if err != nil {
		fmt.Println(err)
		return
	}
	a.UserToken = c.userToken

	// 3. calculate their hit into a new game state
	latestFrame := marbleGame.Frames[len(marbleGame.Frames)-1]
	successfullyValidateFrame, err := marbleGame.ValidateGameAction(a, latestFrame)

	if err != nil {
		// send an error or something
		fmt.Println(err)
	} else {
		newGameFrames := marbleGame.GenerateNewGameFrames(&a, &successfullyValidateFrame)
		marbleGame.Frames = newGameFrames
		// add to the played turns on player
		marbleGame.TurnOrder[marbleGame.ActivePlayerIndex].TurnsTaken++

		// 4. next in turn order
		marbleGame.ActivePlayerIndex++
		if marbleGame.ActivePlayerIndex >= len(marbleGame.TurnOrder) {
			marbleGame.ActivePlayerIndex = 0
		}

		// 5. send the new game state to all the clients
		sendMarbleGameToClients(c.hub, marbleGame)
	}
}

func sendMarbleGameToClients(h *Hub, marbleGame *engine.MarbleGame) {
	marshalledMarbleGame, _ := json.Marshal(marbleGame)
	h.broadcast <- marshalledMarbleGame
}

func sendMarbleGameToClient(c *Client, marbleGame *engine.MarbleGame) {
	marshalledMarbleGame, _ := json.Marshal(marbleGame)
	c.send <- marshalledMarbleGame
}

func gameWritePumpHandler(c *Client, message []byte) error {
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
