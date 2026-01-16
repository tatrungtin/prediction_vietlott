package entity

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/tool_predict/internal/domain/valueobject"
)

// Draw represents a historical Vietlott lottery draw result
type Draw struct {
	ID         string                   `json:"id"`
	GameType   valueobject.GameType     `json:"game_type"`
	DrawNumber int                      `json:"draw_number"`
	Numbers    valueobject.Numbers      `json:"numbers"`
	DrawDate   time.Time                `json:"draw_date"`
	Jackpot    float64                  `json:"jackpot"`
	Winners    int                      `json:"winners"`
	CreatedAt  time.Time                `json:"created_at"`
}

// NewDraw creates a new Draw entity with validation
func NewDraw(
	gameType valueobject.GameType,
	drawNumber int,
	numbers valueobject.Numbers,
	drawDate time.Time,
	jackpot float64,
	winners int,
) (*Draw, error) {
	// Validate game type
	if err := gameType.Validate(); err != nil {
		return nil, fmt.Errorf("invalid game type: %w", err)
	}

	// Validate draw number
	if drawNumber <= 0 {
		return nil, fmt.Errorf("draw number must be positive, got %d", drawNumber)
	}

	// Validate numbers against game type range
	minRange, maxRange := gameType.NumberRange()
	for _, num := range numbers {
		if num < minRange || num > maxRange {
			return nil, fmt.Errorf("number %d is out of range for game type %s (%d-%d)",
				num, gameType, minRange, maxRange)
		}
	}

	// Validate jackpot
	if jackpot < 0 {
		return nil, fmt.Errorf("jackpot cannot be negative, got %f", jackpot)
	}

	// Validate winners
	if winners < 0 {
		return nil, fmt.Errorf("winners cannot be negative, got %d", winners)
	}

	return &Draw{
		ID:         uuid.New().String(),
		GameType:   gameType,
		DrawNumber: drawNumber,
		Numbers:    numbers,
		DrawDate:   drawDate,
		Jackpot:    jackpot,
		Winners:    winners,
		CreatedAt:  time.Now(),
	}, nil
}

// GetID returns the unique identifier of the draw
func (d *Draw) GetID() string {
	return d.ID
}

// GetGameType returns the game type
func (d *Draw) GetGameType() valueobject.GameType {
	return d.GameType
}

// String returns a string representation of the draw
func (d *Draw) String() string {
	return fmt.Sprintf("Draw #%d (%s) on %s: %s, Jackpot: %.0f VND",
		d.DrawNumber,
		d.GameType,
		d.DrawDate.Format("2006-01-02"),
		d.Numbers,
		d.Jackpot,
	)
}
