package creatures

import (
	"github.com/google/uuid"

	"simulator/models/cognitive_functions"
	worldmap "simulator/models/map"
	"simulator/models/physiological_capacities"
)

type Creature struct {
	ID       uuid.UUID
	Species  string
	Age      int
	Location uuid.UUID
	Position worldmap.Position

	Senses     *physiological_capacities.Senses
	Metabolism *physiological_capacities.Metabolism
	Mobility   *physiological_capacities.Mobility
	VitalSigns *physiological_capacities.VitalSigns

	Attention *cognitive_functions.Attention
	Memory    *cognitive_functions.Memory
	Emotion   *cognitive_functions.Emotion
}
