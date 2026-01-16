package valueobject

import (
	"fmt"
)

// GameType represents the type of Vietlott lottery game
type GameType string

const (
	// Mega645 is the Mega 6/45 game (select 6 numbers from 01-45)
	Mega645 GameType = "MEGA_6_45"
	// Power655 is the Power 6/55 game (select 6 numbers from 01-55)
	Power655 GameType = "POWER_6_55"
)

// NumberRange returns the minimum and maximum valid numbers for this game type
func (gt GameType) NumberRange() (int, int) {
	switch gt {
	case Mega645:
		return 1, 45
	case Power655:
		return 1, 55
	default:
		return 1, 45
	}
}

// NumberCount returns the count of numbers to select (always 6 for Vietlott)
func (gt GameType) NumberCount() int {
	return 6
}

// Validate checks if the game type is valid
func (gt GameType) Validate() error {
	if gt != Mega645 && gt != Power655 {
		return fmt.Errorf("invalid game type: %s", gt)
	}
	return nil
}

// String returns the string representation of the game type
func (gt GameType) String() string {
	return string(gt)
}
