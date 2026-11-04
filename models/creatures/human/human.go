package creatures

import "simulator/models/cognitive_functions"

type Human struct {
	Creature

	Name		*string
	Father		*uuid.UUID
	Mother 		*uuid.UUID
	Personality cognitive_functions.Personality
	Reasoning   cognitive_functions.Reasoning
	Goals       []cognitive_functions.Goal
}
