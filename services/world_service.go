package services

import (
	"fmt"
	"math/rand"
	"sort"

	"github.com/google/uuid"

	"simulator/models/creatures"
	"simulator/models/creatures/dog"
	"simulator/models/physiological_capacities"
)

type WorldService struct {
	genomeGenerator *GenomeGenerator
	nameRandom      *rand.Rand
}

const defaultCreatureCount = 10

// CreatureInitializationEvent describes a creature that has been added to the
// world without exposing its concrete Go type. Consumers such as a graphical
// loader can therefore display dogs, humans, or future species uniformly.
type CreatureInitializationEvent struct {
	Sequence     int      `json:"sequence"`
	Total        int      `json:"total"`
	Name         string   `json:"name"`
	Species      string   `json:"species"`
	Variant      string   `json:"variant,omitempty"`
	ID           string   `json:"id"`
	Age          int      `json:"age"`
	HungerMax    int      `json:"hunger_max"`
	HungerRate   float64  `json:"hunger_rate"`
	Capabilities []string `json:"capabilities"`
}

type CreatureInitializationObserver func(CreatureInitializationEvent)

func NewWorldService(seed int64) *WorldService {
	return &WorldService{
		genomeGenerator: NewGenomeGenerator(seed),
		nameRandom:      rand.New(rand.NewSource(seed)),
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
	return service.CreateCreaturesWithObserver(nil)
}

// CreateCreaturesWithObserver creates the world's initial population and
// reports each initialized creature. The observer is optional and is called
// synchronously, in creation order.
func (service *WorldService) CreateCreaturesWithObserver(observer CreatureInitializationObserver) ([]*dog.Dog, error) {
	if service == nil || service.genomeGenerator == nil || service.nameRandom == nil {
		return nil, fmt.Errorf("world service is not initialized")
	}

	dogs := make([]*dog.Dog, 0, defaultCreatureCount)

	for index := 1; index <= defaultCreatureCount; index++ {
		name, err := dog.RandomName(service.nameRandom)
		if err != nil {
			return nil, fmt.Errorf("select dog %d name: %w", index, err)
		}

		createdDog, err := service.CreateDog(name, "Mixed breed")
		if err != nil {
			return nil, fmt.Errorf("create dog %d: %w", index, err)
		}

		dogs = append(dogs, createdDog)
		if observer != nil {
			capabilities := make([]string, 0, len(createdDog.Creature.Capabilities))
			for capability, enabled := range createdDog.Creature.Capabilities {
				if enabled {
					capabilities = append(capabilities, capability)
				}
			}
			sort.Strings(capabilities)

			observer(CreatureInitializationEvent{
				Sequence:     index,
				Total:        defaultCreatureCount,
				Name:         *createdDog.Name,
				Species:      createdDog.Creature.Species,
				Variant:      createdDog.Breed,
				ID:           createdDog.Creature.ID.String(),
				Age:          createdDog.Creature.Age,
				HungerMax:    createdDog.Creature.Genome.HungerMax,
				HungerRate:   createdDog.Creature.Metabolism.HungerRate,
				Capabilities: capabilities,
			})
		}
	}

	return dogs, nil
}
