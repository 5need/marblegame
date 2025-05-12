package routes

import (
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024, // maybe change this to be larger since going to send tons of game frame data
}

func serveWS(hub *Hub, c echo.Context) error {
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
		hub:       hub,
		userToken: userToken,
		conn:      conn,
		send:      make(chan []byte, 256),
	}

	client.hub.register <- client

	go client.writePump()
	go client.readPump()

	if client.hub.registerHandler != nil {
		client.hub.registerHandler(client)
	}

	return nil
}

// TODO: make the readPumpHandler and writePumpHandler nillable, and use the chat thing as the defaults if they are nil
// A Hub handles multiple Clients
type Hub struct {
	clients                  map[*Client]bool
	broadcast                chan []byte
	register                 chan *Client
	unregister               chan *Client
	registerHandler          func(c *Client)
	unregisterHandler        func(c *Client)
	readPumpHandler          func(c *Client, message []byte)
	readPumpDebounceDuration time.Duration
	writePumpHandler         func(c *Client, message []byte) error
}

func newHub(
	registerHandler func(c *Client),
	unregisterHandler func(c *Client),
	readPumpHandler func(c *Client, message []byte),
	readPumpDebounceDuration time.Duration,
	writePumpHandler func(c *Client, message []byte) error,
) *Hub {
	return &Hub{
		clients:                  make(map[*Client]bool),
		broadcast:                make(chan []byte),
		register:                 make(chan *Client),
		unregister:               make(chan *Client),
		registerHandler:          registerHandler,
		unregisterHandler:        unregisterHandler,
		readPumpHandler:          readPumpHandler,
		readPumpDebounceDuration: readPumpDebounceDuration,
		writePumpHandler:         writePumpHandler,
	}
}

func (h *Hub) run() {
	for {
		select {
		case client := <-h.register:
			// triggers whenever register channel gets something
			h.clients[client] = true
			// registerHandler is in serveWS() for reasons. Something to do with starting the goroutines writePump and readPump before executing registerHandler
		case client := <-h.unregister:
			// triggers whenever unregister channel gets something
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
				if h.unregisterHandler != nil {
					deceasedClient := &Client{
						userToken: client.userToken,
					}
					time.AfterFunc(
						100*time.Millisecond,
						func() { h.unregisterHandler(deceasedClient) },
					)
				}
			}
		case message := <-h.broadcast:
			// triggers whenever broadcast channel gets something
			// fmt.Println("broadcasting from Hub")
			for client := range h.clients {
				select {
				case client.send <- message:
				// successfully put message into client send channel
				default:
					// failed to put message into client send channel
					fmt.Println("failed to put message into client send channel")
					close(client.send)
					delete(h.clients, client)
				}
			}
		}
	}
}

const (
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = (pongWait * 9) / 10
	maxMessageSize = 512
)

var (
	newline = []byte{'\n'}
	space   = []byte{' '}
)

// A Client connects a ws connection to its Hub
type Client struct {
	hub       *Hub
	userToken string
	conn      *websocket.Conn
	send      chan []byte
}

// readPump pumps messages from the websocket connection to the hub
func (c *Client) readPump() {
	fmt.Println("starting readPump goroutine")
	defer func() {
		fmt.Println("exiting readPump goroutine")
		c.hub.unregister <- c
		c.conn.Close()
	}()

	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(appData string) error { c.conn.SetReadDeadline(time.Now().Add(pongWait)); return nil })

	messageChan := make(chan []byte)

	// this continuously gets every message from the ws connection and throws it into the messageChan channel to be processed by the debouncer, if needed
	go func() {
		for {
			_, message, err := c.conn.ReadMessage()
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

	shouldDebounce := c.hub.readPumpDebounceDuration != 0

	if shouldDebounce {
		debounceTicker := time.NewTicker(c.hub.readPumpDebounceDuration)
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
					c.hub.readPumpHandler(c, lastMessage)
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
			c.hub.readPumpHandler(c, message)
		}
	}

}

// writePump pumps messages from the hub to the websocket connection(s)
func (c *Client) writePump() {
	fmt.Println("starting writePump goroutine")
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		fmt.Println("exiting writePump goroutine")
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
			}

			err := c.hub.writePumpHandler(c, message)

			if err != nil {
				return
			}
		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}
