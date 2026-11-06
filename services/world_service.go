package services

import (
	"fmt"

	"github.com/google/uuid"

	"simulator/models/creatures"
	"simulator/models/creatures/dog"
	"simulator/models/physiological_capacities"
)

type WorldService struct {
	genomeGenerator *GenomeGenerator
}

const defaultCreatureCount = 10

func NewWorldService(seed int64) *WorldService {
	return &WorldService{
		genomeGenerator: NewGenomeGenerator(seed),
	}
}

func (service *WorldService) CreateDog(name string, breed string) (*dog.Dog, error) {
	if service == nil || service.genomeGenerator == nil {
		return nil, fmt.Errorf("world service is not initialized")
	}

	profile := dog.GenomeProfile()
	genome, err := service.genomeGenerator.Generate(profile)
	if err != nil {
		return nil, fmt.Errorf("generate dog genome: %w", err)
	}

	createdDog := &dog.Dog{
		Creature: creatures.Creature{
			ID:           uuid.New(),
			Species:      profile.ID,
			Genome:       &genome,
			Capabilities: cloneCapabilities(genome.Labels),
			Metabolism: &physiological_capacities.Metabolism{
				HungerRate: genome.HungerRate,
			},
		},
		Breed: breed,
		Name:  &name,
	}

	return createdDog, nil
}

func (service *WorldService) CreateCreatures() ([]*dog.Dog, error) {
	dogs := make([]*dog.Dog, 0, defaultCreatureCount)

	for index := 1; index <= defaultCreatureCount; index++ {
		createdDog, err := service.CreateDog(fmt.Sprintf("Dog %d", index), "Mixed breed")
		if err != nil {
			return nil, fmt.Errorf("create dog %d: %w", index, err)
		}

		dogs = append(dogs, createdDog)
	}

	return dogs, nil
}
