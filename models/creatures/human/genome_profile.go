package human

import (
	"simulator/models/capabilities"
	"simulator/models/creatures"
)

const SpeciesID = "human"

func GenomeProfile() creatures.SpeciesDefinition {
	return creatures.SpeciesDefinition{
		ID: SpeciesID,
		Labels: capabilities.Set{
			capabilities.Animal:   true,
			capabilities.Human:    true,
			capabilities.CanSleep: true,
			capabilities.CanWalk:  true,
		},
		HungerMax: creatures.IntRange{
			Min: 90,
			Max: 130,
		},
		HungerRate: creatures.FloatRange{
			Min: 0.03,
			Max: 0.10,
		},
	}
}
