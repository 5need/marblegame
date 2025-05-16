package websockets

import (
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
)

// A Hub handles multiple Clients
type Hub struct {
	Clients                  map[*Client]bool
	Broadcast                chan []byte
	Register                 chan *Client
	Unregister               chan *Client
	ReadPumpDebounceDuration time.Duration
}

var _ HubInterface = (*Hub)(nil)

func NewHub() *Hub {
	return &Hub{
		Clients:                  make(map[*Client]bool),
		Broadcast:                make(chan []byte),
		Register:                 make(chan *Client),
		Unregister:               make(chan *Client),
		ReadPumpDebounceDuration: 0,
	}
}

type HubInterface interface {
	RegisterHandler(c *Client)
	UnregisterHandler(c *Client)
	UnregisterChan() chan *Client
	BroadcastChan() chan []byte
	ReadPumpHandler(c *Client, message []byte)
	WritePumpHandler(c *Client, message []byte) error
	ShouldDebounce() (bool, time.Duration)
	ServeWS(c echo.Context) error
	Run()
	CloseAllConnections()
}

func (h *Hub) RegisterHandler(c *Client)   {}
func (h *Hub) UnregisterHandler(c *Client) {}
func (h *Hub) UnregisterChan() chan *Client {
	return h.Unregister
}
func (h *Hub) BroadcastChan() chan []byte {
	return h.Broadcast
}
func (h *Hub) ShouldDebounce() (bool, time.Duration) {
	if h.ReadPumpDebounceDuration <= 0 {
		return false, 0
	} else {
		return true, h.ReadPumpDebounceDuration
	}
}
func (h *Hub) ReadPumpHandler(c *Client, message []byte)        {}
func (h *Hub) WritePumpHandler(c *Client, message []byte) error { return nil }

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.Register:
			// triggers whenever register channel gets something
			h.Clients[client] = true
			// registerHandler is in serveWS() for reasons. Something to do with starting the goroutines writePump and readPump before executing registerHandler
		case client := <-h.Unregister:
			// triggers whenever unregister channel gets something
			if _, ok := h.Clients[client]; ok {
				delete(h.Clients, client)
				close(client.Send)
				deceasedClient := &Client{
					UserToken: client.UserToken,
				}
				time.AfterFunc(
					100*time.Millisecond,
					func() { h.UnregisterHandler(deceasedClient) },
				)
			}
		case message := <-h.Broadcast:
			// triggers whenever broadcast channel gets something
			// fmt.Println("broadcasting from Hub")
			for client := range h.Clients {
				select {
				case client.Send <- message:
				// successfully put message into client send channel
				default:
					// failed to put message into client send channel
					fmt.Println("failed to put message into client send channel")
					close(client.Send)
					delete(h.Clients, client)
				}
			}
		}
	}
}

func (h *Hub) ServeWS(c echo.Context) error {
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
	client := &Client{
		Hub:       h,
		UserToken: userToken,
		Conn:      conn,
		Send:      make(chan []byte, 256),
	}

	h.Register <- client

	go client.WritePump()
	go client.ReadPump()

	h.RegisterHandler(client)

	return nil
}

// Unregisters all clients
func (h *Hub) CloseAllConnections() {
	for client := range h.Clients {
		h.Unregister <- client
	}
}

const (
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = (pongWait * 9) / 10
	maxMessageSize = 512
)

// A Client connects a ws connection to its Hub
type Client struct {
	Hub       HubInterface
	UserToken string
	Conn      *websocket.Conn
	Send      chan []byte
}

// ReadPump pumps messages from the websocket connection to the hub
func (c *Client) ReadPump() {
	fmt.Println("starting readPump goroutine")
	defer func() {
		fmt.Println("exiting readPump goroutine")
		c.Hub.UnregisterChan() <- c
		c.Conn.Close()
	}()

	c.Conn.SetReadLimit(maxMessageSize)
	c.Conn.SetReadDeadline(time.Now().Add(pongWait))
	c.Conn.SetPongHandler(func(appData string) error { c.Conn.SetReadDeadline(time.Now().Add(pongWait)); return nil })

	messageChan := make(chan []byte)

	// this continuously gets every message from the ws connection and throws it into the messageChan channel to be processed by the debouncer, if needed
	go func() {
		for {
			_, message, err := c.Conn.ReadMessage()
			if err != nil {
				if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
					log.Printf("error: %v", err)
				}
				close(messageChan) // Close the channel when an error occurs
				return
			}
			messageChan <- message
		}
	}()

	shouldDebounce, debounceDuration := c.Hub.ShouldDebounce()

	if shouldDebounce {
		debounceTicker := time.NewTicker(debounceDuration)
		defer debounceTicker.Stop()

		var lastMessage []byte

		for {
			select {
			case message, ok := <-messageChan:
				if !ok {
					return
				}
				lastMessage = message

			case <-debounceTicker.C:
				if len(lastMessage) != 0 {
					c.Hub.ReadPumpHandler(c, lastMessage)
					lastMessage = nil // Reset after handling
				}
			}
		}
	} else if !shouldDebounce {
		for {
			message, ok := <-messageChan
			if !ok {
				return
			}
			c.Hub.ReadPumpHandler(c, message)
		}
	}

}

// WritePump pumps messages from the hub to the websocket connection(s)
func (c *Client) WritePump() {
	fmt.Println("starting writePump goroutine")
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		fmt.Println("exiting writePump goroutine")
		ticker.Stop()
		c.Conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.Send:
			c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
			}

			err := c.Hub.WritePumpHandler(c, message)

			if err != nil {
				return
			}
		case <-ticker.C:
			c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}
