package creatures

import "simulator/models/capabilities"

type CapabilitySet = capabilities.Set

type SpeciesDefinition struct {
	ID         string        `json:"id"`
	Labels     CapabilitySet `json:"labels,omitempty"`
	HungerMax  IntRange      `json:"hunger_max"`
	HungerRate FloatRange    `json:"hunger_rate"`
}
