package creatures

type Genome struct {
	SpeciesID  string        `json:"species_id"`
	Labels     CapabilitySet `json:"labels,omitempty"`
	HungerMax  int           `json:"hunger_max"`
	HungerRate float64       `json:"hunger_rate"`
}
