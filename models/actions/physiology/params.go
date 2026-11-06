package physiology

import "github.com/google/uuid"

type EatParams struct {
	ItemID   uuid.UUID `json:"item_id"`
	Quantity *float64  `json:"quantity,omitempty"`
}

type DrinkParams struct {
	ItemID   uuid.UUID `json:"item_id"`
	Quantity *float64  `json:"quantity,omitempty"`
}

type SleepParams struct {
	Duration *float64 `json:"duration,omitempty"`
}
