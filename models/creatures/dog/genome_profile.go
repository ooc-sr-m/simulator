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
			// Units per real second. Starting at half capacity, this range
			// reaches the maximum in roughly 8 to 18 hours.
			Min: 0.0009,
			Max: 0.0014,
		},
		ThirstMax: creatures.IntRange{
			Min: 70,
			Max: 110,
		},
		ThirstRate: creatures.FloatRange{
			// Thirst rises somewhat faster than hunger: approximately 4 to
			// 11 hours from half capacity to the maximum.
			Min: 0.0014,
			Max: 0.0022,
		},
	}
}
