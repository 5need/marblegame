package lobby

import (
	"errors"
	"fmt"
	"marblegame/websockets"
	"net/http"
	"slices"
	"strconv"

	"github.com/labstack/echo/v4"
)

var rooms = make(map[int]*Room, 0)

func GetRooms() map[int]*Room {
	return rooms
}

func GetRoom(roomId int) (*Room, error) {
	_, ok := rooms[roomId]
	if !ok {
		return nil, errors.New("Room does not exist")
	}
	return rooms[roomId], nil
}

func NewRoom(roomId int, name string) *Room {
	h := websockets.NewHub()
	newRoom := &Room{
		Hub:         h,
		Id:          strconv.Itoa(roomId),
		Name:        name,
		MaxPlayers:  2,
		PartyLeader: "",
		Players:     []string{},
	}

	rooms[roomId] = newRoom

	go newRoom.Run()

	return newRoom
}

func RoomRoutes(e *echo.Echo) {
	// Shows a list of rooms, and can join by clicking on any available ones
	e.GET("/lobby", func(c echo.Context) error {
		userToken, _ := c.Cookie("userToken")
		// TODO: we're assuming a userToken here
		return Lobby(userToken.Value).Render(c.Request().Context(), c.Response().Writer)
	})

	// Returns a list of rooms
	e.GET("/listOfRooms", func(c echo.Context) error {
		userToken, _ := c.Cookie("userToken")
		return ListOfRooms(GetRooms(), userToken.Value).Render(c.Request().Context(), c.Response().Writer)
	})

	// Brings user to a room, where they can start a game
	e.GET("/room/:roomId", func(c echo.Context) error {
		userToken, _ := c.Cookie("userToken")
		roomId, _ := strconv.Atoi(c.Param("roomId"))

		myLobby, err := GetRoom(roomId)
		if err != nil {
			// lobby not created yet, add it
			rooms[roomId] = NewRoom(roomId, strconv.Itoa(roomId))
			myLobby, _ = GetRoom(roomId)
		}

		// add user to lobby
		if err = myLobby.AddPlayerToRoom(userToken.Value); err != nil {
			fmt.Println(err)
			return c.String(http.StatusUnauthorized, "Room is full or you're not allowed in")
		}

		return RoomView(myLobby, userToken.Value).Render(c.Request().Context(), c.Response().Writer)
	})

	// WebSocket to keep connected to the room
	e.GET("/ws/room/:roomId", func(c echo.Context) error {
		// find the Room, then serve it there
		roomId, _ := strconv.Atoi(c.Param("roomId"))
		myRoom, _ := GetRoom(roomId)

		if ok := slices.Contains(myRoom.Players, c.QueryParam("userToken")); !ok {
			return c.String(http.StatusUnauthorized, "You're not allowed in this room")
		}

		return myRoom.ServeWS(c)
	})
}
