package human

import (
	"github.com/google/uuid"

	"simulator/models/cognitive_functions"
	"simulator/models/creatures"
)

type Human struct {
	creatures.Creature

	Name        *string
	Father      *uuid.UUID
	Mother      *uuid.UUID
	Personality cognitive_functions.Personality
	Reasoning   cognitive_functions.Reasoning
	Goals       []cognitive_functions.Goal
}
