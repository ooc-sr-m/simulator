package dog

import (
	"simulator/models/capabilities"
	"simulator/models/creatures"
)

const SpeciesID = "dog"

func GenomeProfile() creatures.SpeciesDefinition {
	return creatures.SpeciesDefinition{
		ID: SpeciesID,
		Labels: capabilities.Set{
			capabilities.Animal:   true,
			capabilities.Dog:      true,
			capabilities.CanBark:  true,
			capabilities.CanSleep: true,
			capabilities.CanWalk:  true,
		},
		HungerMax: creatures.IntRange{
			Min: 80,
			Max: 120,
		},
		HungerRate: creatures.FloatRange{
			Min: 0.05,
			Max: 0.15,
		},
		ThirstMax: creatures.IntRange{
			Min: 70,
			Max: 110,
		},
		ThirstRate: creatures.FloatRange{
			Min: 0.06,
			Max: 0.18,
		},
	}
}
