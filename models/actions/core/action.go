package core

import (
	"encoding/json"

	"github.com/google/uuid"
)

type ActionType string

type ActionRequest struct {
	ActorID    uuid.UUID       `json:"actor_id"`
	Type       ActionType      `json:"type"`
	Parameters json.RawMessage `json:"parameters,omitempty"`
}
