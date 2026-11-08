package creatures

import (
	"fmt"
	"math"
	"math/rand"
	"strings"

	"simulator/models/capabilities"
	"simulator/models/creatures"
)

type GenomeGenerator struct {
	random *rand.Rand
}

func NewGenomeGenerator(seed int64) *GenomeGenerator {
	return &GenomeGenerator{
		random: rand.New(rand.NewSource(seed)),
	}
}

func (generator *GenomeGenerator) Generate(species creatures.SpeciesDefinition) (creatures.Genome, error) {
	if generator == nil || generator.random == nil {
		return creatures.Genome{}, fmt.Errorf("genome generator is not initialized")
	}
	if err := ValidateSpeciesDefinition(species); err != nil {
		return creatures.Genome{}, err
	}

	return creatures.Genome{
		SpeciesID:  species.ID,
		Labels:     cloneCapabilities(species.Labels),
		HungerMax:  randomInt(generator.random, species.HungerMax),
		HungerRate: randomFloat(generator.random, species.HungerRate),
		ThirstMax:  randomInt(generator.random, species.ThirstMax),
		ThirstRate: randomFloat(generator.random, species.ThirstRate),
	}, nil
}

func ValidateSpeciesDefinition(species creatures.SpeciesDefinition) error {
	if strings.TrimSpace(species.ID) == "" {
		return fmt.Errorf("species id is required")
	}
	if species.HungerMax.Min < 0 {
		return fmt.Errorf("hunger max minimum cannot be negative")
	}
	if species.HungerMax.Min > species.HungerMax.Max {
		return fmt.Errorf("invalid hunger max range: minimum exceeds maximum")
	}
	if math.IsNaN(species.HungerRate.Min) || math.IsNaN(species.HungerRate.Max) ||
		math.IsInf(species.HungerRate.Min, 0) || math.IsInf(species.HungerRate.Max, 0) {
		return fmt.Errorf("hunger rate range must contain finite numbers")
	}
	if species.HungerRate.Min < 0 {
		return fmt.Errorf("hunger rate minimum cannot be negative")
	}
	if species.HungerRate.Min > species.HungerRate.Max {
		return fmt.Errorf("invalid hunger rate range: minimum exceeds maximum")
	}
	if species.ThirstMax.Min < 0 {
		return fmt.Errorf("thirst max minimum cannot be negative")
	}
	if species.ThirstMax.Min > species.ThirstMax.Max {
		return fmt.Errorf("invalid thirst max range: minimum exceeds maximum")
	}
	if math.IsNaN(species.ThirstRate.Min) || math.IsNaN(species.ThirstRate.Max) ||
		math.IsInf(species.ThirstRate.Min, 0) || math.IsInf(species.ThirstRate.Max, 0) {
		return fmt.Errorf("thirst rate range must contain finite numbers")
	}
	if species.ThirstRate.Min < 0 {
		return fmt.Errorf("thirst rate minimum cannot be negative")
	}
	if species.ThirstRate.Min > species.ThirstRate.Max {
		return fmt.Errorf("invalid thirst rate range: minimum exceeds maximum")
	}
	return nil
}

func randomInt(random *rand.Rand, valueRange creatures.IntRange) int {
	if valueRange.Min == valueRange.Max {
		return valueRange.Min
	}
	return valueRange.Min + random.Intn(valueRange.Max-valueRange.Min+1)
}

func randomFloat(random *rand.Rand, valueRange creatures.FloatRange) float64 {
	if valueRange.Min == valueRange.Max {
		return valueRange.Min
	}
	return valueRange.Min + random.Float64()*(valueRange.Max-valueRange.Min)
}

func cloneCapabilities(source capabilities.Set) capabilities.Set {
	if source == nil {
		return nil
	}

	clone := make(capabilities.Set, len(source))
	for label, enabled := range source {
		clone[label] = enabled
	}
	return clone
}
