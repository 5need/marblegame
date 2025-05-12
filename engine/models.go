package engine

import (
	"github.com/deeean/go-vector/vector2"
	"github.com/ungerik/go3d/float64/quaternion"
)

type MarbleGame struct {
	Players           map[string]*Player `json:"players"`
	Frames            []MarbleGameFrame  `json:"frames"`
	Config            MarbleGameConfig   `json:"config"`
	TurnOrder         []*Player          `json:"turnOrder"`
	ActivePlayerIndex int                `json:"activePlayerIndex"` // index from TurnOrder, whose turn it is
}

type MarbleGameConfig struct {
	PlayerLimit                         int     `json:"playerLimit"`
	ScoringZoneRadius                   float64 `json:"scoringZoneRadius"`
	ScoringZoneMaxScore                 int     `json:"scoringZoneMaxScore"`
	ScoringZoneMinScore                 int     `json:"scoringZoneMinScore"`
	BullseyeZoneRadius                  float64 `json:"bullseyeZoneRadius"`
	BullseyeZoneScore                   int     `json:"bullseyeZoneScore"`
	Width                               int     `json:"width"`
	Height                              int     `json:"height"`
	RemoveMarblesFromOutsideScoringZone bool    `json:"removeMarblesFromOutsideScoringZone"`
}

// A game frame is sent as a representation of the entire game state.
// Multiple game frames are sent every action (like hitting a marble).
type MarbleGameFrame struct {
	Marbles []Marble `json:"marbles"`
}

// A struct represeting the player.
type Player struct {
	UserToken         string       `json:"userToken"`
	DisplayName       string       `json:"displayName"`
	Score             int          `json:"score"`
	Hue               int          `json:"hue"`
	ShouldSkipMyTurns bool         `json:"shouldSkipMyTurns"`
	TurnsTaken        int          `json:"turnsTaken"`
	Inventory         []MarbleType `json:"inventory"`
}

type Action struct {
	InventorySlot int             `json:"inventorySlot"`
	Pos           vector2.Vector2 `json:"pos"`
	Vel           vector2.Vector2 `json:"vel"`
	UserToken     string          `json:"userToken"`
}

type Marble struct {
	Pos            vector2.Vector2 `json:"pos"`
	Vel            vector2.Vector2 `json:"vel"`
	Rot            quaternion.T    `json:"rot"`
	Score          int             `json:"score"`
	Type           MarbleType      `json:"type"`
	Collided       bool            `json:"collided"`
	HighlightColor string          `json:"highlightColor"`
	Owner          *Player         `json:"owner"`
}

type MarbleType struct {
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Radius      float64 `json:"radius"`
	Mass        float64 `json:"mass"`
}
