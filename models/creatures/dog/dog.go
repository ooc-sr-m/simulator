package dog

import (
	"github.com/google/uuid"

	"simulator/models/cognitive_functions"
	"simulator/models/creatures"
)

type Dog struct {
	Creature    creatures.Creature
	Breed       string
	Name        *string
	Father      *uuid.UUID
	Mother      *uuid.UUID
	Personality cognitive_functions.Personality
}
