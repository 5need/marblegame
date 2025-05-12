package lobby

import (
	"errors"
	"slices"
	"strconv"
)

type Lobby struct {
	Id      string
	Name    string
	Players []string
}

var lobbies = make(map[int]*Lobby, 0)

func GetLobbies() map[int]*Lobby {
	return lobbies
}

func GetLobby(lobbyId int) (*Lobby, error) {
	_, ok := lobbies[lobbyId]
	if !ok {
		return nil, errors.New("Lobby does not exist")
	}
	return lobbies[lobbyId], nil
}

func NewLobby(lobbyId int) *Lobby {
	newLobby := &Lobby{
		Id:      strconv.Itoa(lobbyId),
		Name:    strconv.Itoa(lobbyId),
		Players: []string{},
	}

	lobbies[lobbyId] = newLobby

	return newLobby
}

func (lobby *Lobby) AddPlayer(userToken string) error {
	alreadyInLobby := slices.Contains(lobby.Players, userToken)

	if alreadyInLobby {
		return errors.New("Player already in lobby")
	}

	lobby.Players = append(lobby.Players, userToken)
	return nil
}

func (lobby *Lobby) RemovePlayerFromLobby(userToken string) error {
	inLobby := slices.Contains(lobby.Players, userToken)
	if !inLobby {
		return errors.New("Player not in lobby already")
	}

	updatedPlayerList := []string{}
	for _, player := range lobby.Players {
		if player != userToken {
			updatedPlayerList = append(updatedPlayerList, player)
		}
	}

	lobby.Players = updatedPlayerList
	return nil
}
