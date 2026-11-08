package services

import (
	"fmt"
	"math/rand"
	"sort"
	"sync"

	"github.com/google/uuid"

	"simulator/models/capabilities"
	"simulator/models/creatures"
	"simulator/models/creatures/dog"
	worldmap "simulator/models/map"
	"simulator/models/physiological_capacities"
)

// WorldService gestiona el mundo de simulación y la creación de criaturas
type WorldService struct {
	nameRandom *rand.Rand
	dogs       []*dog.Dog
	mutex      sync.Mutex
}

// CreatureInitializationEvent representa un evento de inicialización de criatura
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

// CreatureStatus representa el estado actual de una criatura
type CreatureStatus struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Species   string `json:"species"`
	Hunger    int    `json:"hunger"`
	MaxHunger int    `json:"max_hunger"`
	Thirst    int    `json:"thirst"`
	MaxThirst int    `json:"max_thirst"`
}

// CreatureInitializationObserver es un callback para eventos de inicialización de criaturas
type CreatureInitializationObserver func(CreatureInitializationEvent)

const defaultCreatureCount = 10

// NewWorldService crea un nuevo servicio de mundo con la semilla especificada
func NewWorldService(seed int64) *WorldService {
	return &WorldService{
		nameRandom: rand.New(rand.NewSource(seed)),
	}
}

// cloneCapabilities crea una copia profunda de un conjunto de capacidades
func cloneCapabilities(source capabilities.Set) capabilities.Set {
	if source == nil {
		return make(capabilities.Set)
	}

	result := make(capabilities.Set, len(source))
	for k, v := range source {
		result[k] = v
	}
	return result
}

// CreateDog crea un nuevo perro en el mundo
func (service *WorldService) CreateDog(name string, breed string) (*dog.Dog, error) {
	if service == nil || service.nameRandom == nil {
		return nil, fmt.Errorf("world service is not initialized")
	}

	profile := dog.GenomeProfile()
	capabilities := cloneCapabilities(profile.Labels)

	hungerMax := profile.HungerMax.Min + service.nameRandom.Intn(profile.HungerMax.Max-profile.HungerMax.Min+1)
	hungerRate := profile.HungerRate.Min + service.nameRandom.Float64()*(profile.HungerRate.Max-profile.HungerRate.Min)
	thirstMax := profile.ThirstMax.Min + service.nameRandom.Intn(profile.ThirstMax.Max-profile.ThirstMax.Min+1)
	thirstRate := profile.ThirstRate.Min + service.nameRandom.Float64()*(profile.ThirstRate.Max-profile.ThirstRate.Min)

	createdDog := &dog.Dog{
		Creature: creatures.Creature{
			ID:           uuid.New(),
			Species:      profile.ID,
			Genome: &creatures.Genome{
				SpeciesID:  profile.ID,
				Labels:     capabilities,
				HungerMax:  hungerMax,
				HungerRate: hungerRate,
				ThirstMax:  thirstMax,
				ThirstRate: thirstRate,
			},
			Capabilities: capabilities,
			Age:          0,
			Location:     uuid.New(),
			Position:     worldmap.Position{X: 0, Y: 0},
			Body: &physiological_capacities.Body{
				Height: 1.0,
				Weight: 10.0,
			},
		},
		Breed: breed,
		Name:  &name,
	}

	return createdDog, nil
}

// CreateCreatures crea un conjunto de criaturas predeterminadas en el mundo
func (service *WorldService) CreateCreatures() ([]*dog.Dog, error) {
	return service.CreateCreaturesWithObserver(nil)
}

// CreateCreaturesWithObserver crea criaturas con un observador para eventos de inicialización
func (service *WorldService) CreateCreaturesWithObserver(observer CreatureInitializationObserver) ([]*dog.Dog, error) {
	if service == nil || service.nameRandom == nil {
		return nil, fmt.Errorf("world service is not initialized")
	}

	dogs := make([]*dog.Dog, 0, defaultCreatureCount)

	for index := 1; index <= defaultCreatureCount; index++ {
		name, err := dog.RandomName(service.nameRandom)
		if err != nil {
			return nil, fmt.Errorf("generate dog name: %w", err)
		}

		dogInstance, err := service.CreateDog(name, "Mixed")
		if err != nil {
			return nil, fmt.Errorf("create dog: %w", err)
		}

		if observer != nil {
			capabilities := make([]string, 0, len(dogInstance.Creature.Capabilities))
			for capability, enabled := range dogInstance.Creature.Capabilities {
				if enabled {
					capabilities = append(capabilities, capability)
				}
			}
			sort.Strings(capabilities)

			observer(CreatureInitializationEvent{
				Sequence:     index,
				Total:        defaultCreatureCount,
				Name:         *dogInstance.Name,
				Species:      dogInstance.Creature.Species,
				Variant:      dogInstance.Breed,
				ID:           dogInstance.Creature.ID.String(),
				Age:          dogInstance.Creature.Age,
				HungerMax:    dogInstance.Creature.Genome.HungerMax,
				HungerRate:   dogInstance.Creature.Genome.HungerRate,
				Capabilities: capabilities,
			})
		}

		dogs = append(dogs, dogInstance)
	}

	service.mutex.Lock()
	service.dogs = dogs
	service.mutex.Unlock()

	return dogs, nil
}

// GetCreatureStatus devuelve el estado actual de una criatura específica
func (service *WorldService) GetCreatureStatus(creatureID string) (*CreatureStatus, error) {
	service.mutex.Lock()
	defer service.mutex.Unlock()

	for _, dog := range service.dogs {
		if dog.Creature.ID.String() == creatureID {
			return &CreatureStatus{
				ID:        dog.Creature.ID.String(),
				Name:      *dog.Name,
				Species:   dog.Creature.Species,
				Hunger:    0, // Se llenará con el servicio de metabolismo
				MaxHunger: dog.Creature.Genome.HungerMax,
				Thirst:    0, // Se llenará con el servicio de metabolismo
				MaxThirst: dog.Creature.Genome.ThirstMax,
			}, nil
		}
	}

	return nil, fmt.Errorf("creature not found: %s", creatureID)
}

// GetAllCreatureStatus devuelve el estado actual de todas las criaturas
func (service *WorldService) GetAllCreatureStatus() ([]*CreatureStatus, error) {
	service.mutex.Lock()
	defer service.mutex.Unlock()

	statuses := make([]*CreatureStatus, 0, len(service.dogs))
	for _, dog := range service.dogs {
		statuses = append(statuses, &CreatureStatus{
			ID:        dog.Creature.ID.String(),
			Name:      *dog.Name,
			Species:   dog.Creature.Species,
			Hunger:    0, // Se llenará con el servicio de metabolismo
			MaxHunger: dog.Creature.Genome.HungerMax,
			Thirst:    0, // Se llenará con el servicio de metabolismo
			MaxThirst: dog.Creature.Genome.ThirstMax,
		})
	}

	return statuses, nil
}
