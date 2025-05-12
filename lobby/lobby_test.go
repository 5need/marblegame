package lobby_test

import (
	"marblegame/models"
	"reflect"
)

func TestLobby(t *testing.T) {
	testCases := []struct {
		desc  string
		input models.Lobby
		want  models.Lobby
	}{
		{desc: "", input: models.Lobby{}, want: models.Lobby{}},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			ans := myfunction(tC.input)
			if !reflect.DeepEqual(ans, tC.want) {
				t.Errorf("FAIL %s: got %v, want %v", tC.desc, ans, tC.want)
			}
		})
	}
}
