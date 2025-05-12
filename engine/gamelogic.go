package engine

import (
	"errors"

	"github.com/deeean/go-vector/vector2"
	"github.com/ungerik/go3d/float64/quaternion"
	"github.com/ungerik/go3d/float64/vec3"
)

var MarbleTypes = []MarbleType{
	{
		Name:        "Marble",
		Description: "Scores normally.",
		Radius:      30,
		Mass:        10,
	},
	{
		Name:        "Big Marble",
		Description: "Big. Scores normally.",
		Radius:      50,
		Mass:        20,
	},
	{
		Name:        "Small Marble",
		Description: "Small. Scores normally.",
		Radius:      15,
		Mass:        5,
	},
}

func NewMarbleGame() *MarbleGame {
	return &MarbleGame{
		Players: make(map[string]*Player),
		Frames:  []MarbleGameFrame{{Marbles: []Marble{}}},
		Config: MarbleGameConfig{
			PlayerLimit:                         2,
			ScoringZoneRadius:                   150.0,
			ScoringZoneMaxScore:                 20,
			ScoringZoneMinScore:                 5,
			BullseyeZoneRadius:                  15.0,
			BullseyeZoneScore:                   40,
			Width:                               600,
			Height:                              480,
			RemoveMarblesFromOutsideScoringZone: true,
		},
		TurnOrder:         []*Player{},
		ActivePlayerIndex: 0,
	}
}

// Handles validating a legal game action
// returns an error if invalid
// if it's valid it will send a new MarbleGameFrame with the new Marble
func (marbleGame *MarbleGame) ValidateGameAction(action Action, frame MarbleGameFrame) (MarbleGameFrame, error) {
	player, exists := marbleGame.Players[action.UserToken]
	if !exists {
		return MarbleGameFrame{}, errors.New("Invalid Player")
	}

	// check if it's the players turn
	if marbleGame.TurnOrder[marbleGame.ActivePlayerIndex] != player {
		return MarbleGameFrame{}, errors.New("Not your turn")
	}

	// check if the player has it in their inventory
	var newMarbleType MarbleType
	if action.InventorySlot >= 0 && action.InventorySlot < len(player.Inventory) {
		newMarbleType = player.Inventory[action.InventorySlot]
	} else {
		if action.InventorySlot == -1 {
			return MarbleGameFrame{}, errors.New("No Inventory Slot selected")
		}
		return MarbleGameFrame{}, errors.New("Inventory Slot out of range")
	}

	// remove it from inventory
	newInventory := []MarbleType{}
	for i := range player.Inventory {
		if i != action.InventorySlot {
			newInventory = append(newInventory, player.Inventory[i])
		}

	}
	player.Inventory = newInventory

	// check if the ball fits

	// check if the velocity is not too large, and clamp if so

	// adjust player's inventory

	// success, add the new ball
	velClampedInsideGameField := action.Vel.Copy()
	if velClampedInsideGameField.X < 0 {
		velClampedInsideGameField.X = 0
	}
	if velClampedInsideGameField.X > float64(marbleGame.Config.Width) {
		velClampedInsideGameField.X = float64(marbleGame.Config.Width)
	}
	if velClampedInsideGameField.Y < 0 {
		velClampedInsideGameField.Y = 0
	}
	if velClampedInsideGameField.Y > float64(marbleGame.Config.Height) {
		velClampedInsideGameField.Y = float64(marbleGame.Config.Height)
	}
	vel := action.Pos.Sub(velClampedInsideGameField).MulScalar(-1.0)
	velNormal := vel.Normalize()
	velMagnitude := vel.Magnitude()
	maxMagnitude := 215.0
	if velMagnitude > maxMagnitude {
		vel = velNormal.MulScalar(maxMagnitude)
	}

	newMarble := Marble{
		Pos:            action.Pos,
		Vel:            *vel,
		Rot:            quaternion.Ident, // identity quaternion
		Score:          0,
		Type:           newMarbleType,
		Collided:       false,
		HighlightColor: "",
		Owner:          player,
	}
	var newFrame MarbleGameFrame
	for _, m := range frame.Marbles {
		newFrame.Marbles = append(newFrame.Marbles, m)
	}
	newFrame.Marbles = append(newFrame.Marbles, newMarble)

	return newFrame, nil
}

func (marbleGame *MarbleGame) GenerateNewGameFrames(action *Action, frame *MarbleGameFrame) []MarbleGameFrame {
	previousFrame := frame

	var newGameFrames []MarbleGameFrame
	newGameFrames = append(newGameFrames, *previousFrame) // cheeky, add the previous frame bcuz
	for {
		// copy over marbles to new frame
		var newFrame MarbleGameFrame
		for _, m := range previousFrame.Marbles {
			newFrame.Marbles = append(newFrame.Marbles, m)
		}

		for i := range newFrame.Marbles {
			m := &newFrame.Marbles[i]
			m.Pos = *m.Pos.Add(m.Vel.MulScalar(-0.1))
			m.Vel = *m.Vel.MulScalar(0.96) // friction
			if m.Vel.Magnitude() < 1 {
				m.Vel = vector2.Vector2{X: 0, Y: 0}
			}
		}

		newFrame.HandleCollisions(marbleGame)
		newFrame.HandleScoring(marbleGame)

		newGameFrames = append(newGameFrames, newFrame)
		if newFrame.AreMarblesSettled() {
			break
		}
		previousFrame = &newFrame
	}

	finalFrame := &newGameFrames[len(newGameFrames)-1]

	finalFrame.ResetRotations()

	if marbleGame.Config.RemoveMarblesFromOutsideScoringZone {
		removeMeByIndex := []int{}

		// first pass, mark for removal
		for i, m := range finalFrame.Marbles {
			distanceToCenter := m.Pos.Distance(&vector2.Vector2{X: float64(marbleGame.Config.Width) / 2, Y: float64(marbleGame.Config.Height) / 2})
			if distanceToCenter-m.Type.Radius > marbleGame.Config.ScoringZoneRadius {
				removeMeByIndex = append(removeMeByIndex, i)
			}
		}

		// second pass, remove em
		safeMarbles := []Marble{}
		for i, m := range finalFrame.Marbles {
			toBeRemoved := false
			for _, j := range removeMeByIndex {
				if i == j {
					toBeRemoved = true
					break
				}
			}
			if !toBeRemoved {
				safeMarbles = append(safeMarbles, m)
			}
		}

		finalFrame.Marbles = safeMarbles
	}

	finalFrame.ResetCollidedFlags()

	return newGameFrames
}

func (frame *MarbleGameFrame) ResetCollidedFlags() {
	for i := range frame.Marbles {
		marble := &frame.Marbles[i]
		marble.Collided = false
	}
}

func (frame *MarbleGameFrame) ResetRotations() {
	for i := range frame.Marbles {
		marble := &frame.Marbles[i]
		marble.Rot = quaternion.Ident
	}
}

func (frame *MarbleGameFrame) HandleCollisions(marbleGame *MarbleGame) {
	// handle rotating the marbles
	for i := range frame.Marbles {
		marble := &frame.Marbles[i]

		velocity := marble.Vel
		radius := marble.Type.Radius
		//
		qPrev := marble.Rot

		// Compute the new rotation quaternion
		qNew := CalculateRotationQuaternion(velocity.X, velocity.Y, radius)

		// Multiply previous rotation by new rotation to accumulate rotation over frames
		qFinal := quaternion.Mul(&qPrev, &qNew)

		// qNew := CalculateRotationQuaternion(velocity.X, velocity.Y, radius)
		// qFinal := MultiplyQuaternions(qPrev, qNew)
		//
		marble.Rot = qFinal
		// marble.Rot += 0.01
	}

	// reset Collided flags
	frame.ResetCollidedFlags()

	// collisions between marbles
	for i := range frame.Marbles {
		marble1 := &frame.Marbles[i]
		for j := i + 1; j < len(frame.Marbles); j++ {
			marble2 := &frame.Marbles[j]

			centersDistance := vector2.Distance(&marble1.Pos, &marble2.Pos)
			minDistance := marble1.Type.Radius + marble2.Type.Radius

			if centersDistance < minDistance {
				// Mark these as collided
				marble1.Collided = true
				marble2.Collided = true

				// Resolve overlap (shove them apart)
				overlap := minDistance - centersDistance
				separationDir := marble1.Pos.Sub(&marble2.Pos).Normalize()
				marble1.Pos = *marble1.Pos.Add(separationDir.MulScalar((overlap / 2) + 1))
				marble2.Pos = *marble2.Pos.Sub(separationDir.MulScalar((overlap / 2) + 1))

				// Compute new velocities using 2D elastic collision formula
				m1, m2 := marble1.Type.Mass, marble2.Type.Mass
				v1, v2 := marble1.Vel, marble2.Vel

				normal := separationDir
				tangent := vector2.Vector2{X: -normal.Y, Y: normal.X} // Perpendicular vector

				// Decompose velocities into normal and tangential components
				v1n := normal.Dot(&v1)
				v1t := tangent.Dot(&v1)
				v2n := normal.Dot(&v2)
				v2t := tangent.Dot(&v2)

				// Compute new normal velocities using 1D elastic collision formula
				v1nFinal := (v1n*(m1-m2) + 2*m2*v2n) / (m1 + m2)
				v2nFinal := (v2n*(m2-m1) + 2*m1*v1n) / (m1 + m2)

				// Convert back to 2D velocity
				marble1.Vel = *normal.MulScalar(v1nFinal).Add(tangent.MulScalar(v1t))
				marble2.Vel = *normal.MulScalar(v2nFinal).Add(tangent.MulScalar(v2t))
			}
		}
	}

	// collisions on border walls
	for i := range frame.Marbles {
		marble := &frame.Marbles[i]
		isInLeftWall := marble.Pos.X-marble.Type.Radius < 0
		isInRightWall := marble.Pos.X+marble.Type.Radius > float64(marbleGame.Config.Width)
		isInTopWall := marble.Pos.Y-marble.Type.Radius < 0
		isInBottomWall := marble.Pos.Y+marble.Type.Radius > float64(marbleGame.Config.Height)

		if isInLeftWall {
			// push marble out of wall, then reverse momentum
			marble.Pos.X = 0 + marble.Type.Radius
			marble.Vel.X = -marble.Vel.X
		}
		if isInRightWall {
			// push marble out of wall, then reverse momentum
			marble.Pos.X = float64(marbleGame.Config.Width) - marble.Type.Radius
			marble.Vel.X = -marble.Vel.X
		}
		if isInTopWall {
			// push marble out of wall, then reverse momentum
			marble.Pos.Y = 0 + marble.Type.Radius
			marble.Vel.Y = -marble.Vel.Y
		}
		if isInBottomWall {
			// push marble out of wall, then reverse momentum
			marble.Pos.Y = float64(marbleGame.Config.Height) - marble.Type.Radius
			marble.Vel.Y = -marble.Vel.Y
		}
	}
}

func CalculateRotationQuaternion(vx, vy, radius float64) quaternion.T {
	// Calculate the velocity vector and its magnitude
	velocity := vec3.T{vx, vy, 0}
	speed := velocity.Length()

	// If there's no movement, return identity quaternion (no rotation)
	if speed == 0 {
		return quaternion.Ident
	}

	// Calculate the rotation axis (perpendicular to velocity in 2D plane)
	axis := vec3.T{-vy, vx, 0} // Perpendicular to velocity in the X-Y plane
	axis.Normalize()           // Ensure the axis is a unit vector

	// Calculate the rotation angle based on the speed and radius
	theta := speed / radius / 10

	// Create a quaternion from the rotation axis and angle
	q := quaternion.FromAxisAngle(&axis, theta)

	return q
}

func (frame *MarbleGameFrame) HandleScoring(marbleGame *MarbleGame) {
	// reset scoring to 0
	for _, v := range marbleGame.Players {
		v.Score = 0
	}

	// 1st pass: base scoring
	for i := range frame.Marbles {
		marble1 := &frame.Marbles[i]
		distanceToCenter := marble1.Pos.Distance(&vector2.Vector2{X: 300, Y: 240})

		score := 0
		scoringZoneRadius := marbleGame.Config.ScoringZoneRadius
		scoringZoneMaxScore := marbleGame.Config.ScoringZoneMaxScore
		scoringZoneMinScore := marbleGame.Config.ScoringZoneMinScore
		bullseyeZoneRadius := marbleGame.Config.BullseyeZoneRadius
		bullseyeZoneScore := marbleGame.Config.BullseyeZoneScore

		isInScoringZone := distanceToCenter <= scoringZoneRadius+marble1.Type.Radius
		isInBullseyeZone := distanceToCenter <= bullseyeZoneRadius+marble1.Type.Radius

		if isInScoringZone {
			if isInBullseyeZone {
				score = bullseyeZoneScore
				marble1.HighlightColor = "#ffffffff"
			} else {
				distanceToBullseyeZone := distanceToCenter - marble1.Type.Radius - bullseyeZoneRadius
				distanceFromBullseyeToScoringZone := scoringZoneRadius - bullseyeZoneRadius
				percentage := distanceToBullseyeZone / distanceFromBullseyeToScoringZone
				score = scoringZoneMinScore + int((1-percentage)*float64(scoringZoneMaxScore-scoringZoneMinScore+1))
				marble1.HighlightColor = "#ffffff80"
			}
		} else {
			marble1.HighlightColor = "#12121200"
		}
		marble1.Score = score
		marble1.Owner.Score += score
	}
}

func (frame *MarbleGameFrame) AreMarblesSettled() bool {
	isSettled := true

	for i := range frame.Marbles {
		m := &frame.Marbles[i]
		if m.Vel.Magnitude() != 0 {
			isSettled = false
			return isSettled
		}
	}

	return isSettled
}
