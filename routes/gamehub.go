package routes

import (
	"encoding/json"
	"fmt"
	"marblegame/engine"
	"marblegame/websockets"
	"math/rand"
	"time"

	"github.com/gorilla/websocket"
)

type GameHub struct {
	websockets.Hub
}

var _ websockets.HubInterface = (*GameHub)(nil)

func (gh *GameHub) RegisterHandler(c *websockets.Client) {
	time.AfterFunc(500*time.Millisecond, // TODO: this is jank sauce
		func() {
			if _, exists := marbleGame.Players[c.UserToken]; !exists {
				joiningPlayer := engine.Player{
					UserToken:         c.UserToken,
					DisplayName:       c.UserToken[:4],
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
				marbleGame.Players[c.UserToken] = &joiningPlayer
				marbleGame.TurnOrder = append(marbleGame.TurnOrder, &joiningPlayer)
			}

			gh.sendMarbleGameToClient(c, marbleGame)
		},
	)
}

type ActionRequest struct {
	ActionString string `json:"action"` // stringified input cause lazy
}

func (gh *GameHub) ReadPumpHandler(c *websockets.Client, message []byte) {
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
	a.UserToken = c.UserToken

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
		gh.sendMarbleGameToClients(marbleGame)
	}
}

func (gh *GameHub) WritePumpHandler(c *websockets.Client, message []byte) error {
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

func (gh *GameHub) sendMarbleGameToClients(marbleGame *engine.MarbleGame) {
	marshalledMarbleGame, _ := json.Marshal(marbleGame)
	gh.Broadcast <- marshalledMarbleGame
}

func (gh *GameHub) sendMarbleGameToClient(c *websockets.Client, marbleGame *engine.MarbleGame) {
	marshalledMarbleGame, _ := json.Marshal(marbleGame)
	c.Send <- marshalledMarbleGame
}
