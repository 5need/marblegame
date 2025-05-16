package lobby_test

import (
	"marblegame/lobby"
	"reflect"
	"testing"
)

func TestPlayerAdding(t *testing.T) {
	testCases := []struct {
		desc  string
		input []string
		want  []string
	}{
		{
			desc:  "No players",
			input: []string{},
			want:  []string{},
		},
		{
			desc:  "Add one player multiple times",
			input: []string{"player 1", "player 1", "player 1"},
			want:  []string{"player 1"},
		},
		{
			desc:  "Add player to full lobby",
			input: []string{"player 1", "player 2", "player 3"},
			want:  []string{"player 1", "player 2"},
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			l := lobby.NewRoom(123, "123")
			l.MaxPlayers = 2

			for _, p := range tC.input {
				l.AddPlayerToRoom(p)
			}

			ans := l.Players

			if !reflect.DeepEqual(ans, tC.want) {
				t.Errorf("FAIL %s: got %v, want %v", tC.desc, ans, tC.want)
			}
		})
	}
}

func TestPlayerRemoval(t *testing.T) {
	testCases := []struct {
		desc  string
		input []string
		want  []string
	}{
		{
			desc:  "No players",
			input: []string{},
			want:  []string{"player 1", "player 2"},
		},
		{
			desc:  "Remove one player multiple times",
			input: []string{"player 1", "player 1", "player 1"},
			want:  []string{"player 2"},
		},
		{
			desc:  "Remove all players",
			input: []string{"player 1", "player 2"},
			want:  []string{},
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			l := lobby.NewRoom(123, "123")
			l.MaxPlayers = 2
			l.Players = []string{"player 1", "player 2"}

			for _, p := range tC.input {
				l.RemovePlayerFromRoom(p)
			}

			ans := l.Players

			if !reflect.DeepEqual(ans, tC.want) {
				t.Errorf("FAIL %s: got %v, want %v", tC.desc, ans, tC.want)
			}
		})
	}
}
