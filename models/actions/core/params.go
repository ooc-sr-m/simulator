package core

import "github.com/google/uuid"

type TargetParams struct {
	TargetID uuid.UUID `json:"target_id"`
}

type ItemParams struct {
	ItemID uuid.UUID `json:"item_id"`
}
