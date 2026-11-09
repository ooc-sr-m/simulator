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
	world      *worldmap.GlobalScene
	dogs       []*dog.Dog
	mutex      sync.Mutex
}

// WorldGenerationConfig define la cuadrícula y la escala de los chunks.
type WorldGenerationConfig struct {
	Columns          int
	Rows             int
	ChunkWidth       float64
	ChunkHeight      float64
	ChunkPixelWidth  int
	ChunkPixelHeight int
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
	Hunger       int      `json:"hunger"`
	HungerMax    int      `json:"hunger_max"`
	HungerRate   float64  `json:"hunger_rate"`
	Thirst       int      `json:"thirst"`
	ThirstMax    int      `json:"thirst_max"`
	ThirstRate   float64  `json:"thirst_rate"`
	Capabilities []string `json:"capabilities"`
}

// CreatureStatus representa el estado actual de una criatura
type CreatureStatus struct {
	ID         string  `json:"id"`
	Name       string  `json:"name"`
	Species    string  `json:"species"`
	Hunger     int     `json:"hunger"`
	MaxHunger  int     `json:"max_hunger"`
	HungerRate float64 `json:"hunger_rate"`
	Thirst     int     `json:"thirst"`
	MaxThirst  int     `json:"max_thirst"`
	ThirstRate float64 `json:"thirst_rate"`
}

// CreatureInitializationObserver es un callback para eventos de inicialización de criaturas
type CreatureInitializationObserver func(CreatureInitializationEvent)

const defaultCreatureCount = 10

var defaultWorldGenerationConfig = WorldGenerationConfig{
	Columns:          1,
	Rows:             1,
	ChunkWidth:       100,
	ChunkHeight:      100,
	ChunkPixelWidth:  100,
	ChunkPixelHeight: 100,
}

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
			ID:      uuid.New(),
			Species: profile.ID,
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

// CreateWorld genera un mundo mínimo usando la configuración predeterminada.
func (service *WorldService) CreateWorld() (*worldmap.GlobalScene, error) {
	return service.CreateWorldWithConfig(defaultWorldGenerationConfig)
}

// CreateWorldWithConfig genera todos los chunks en orden de filas y almacena
// la escena resultante como el mundo activo del servicio.
func (service *WorldService) CreateWorldWithConfig(config WorldGenerationConfig) (*worldmap.GlobalScene, error) {
	if service == nil || service.nameRandom == nil {
		return nil, fmt.Errorf("world service is not initialized")
	}
	if config.Columns <= 0 || config.Rows <= 0 {
		return nil, fmt.Errorf("world grid must contain at least one row and one column")
	}
	if config.ChunkWidth <= 0 || config.ChunkHeight <= 0 {
		return nil, fmt.Errorf("chunk dimensions must be greater than zero")
	}
	if config.ChunkPixelWidth <= 0 || config.ChunkPixelHeight <= 0 {
		return nil, fmt.Errorf("chunk pixel dimensions must be greater than zero")
	}

	chunks := make([]worldmap.Chunk, 0, config.Columns*config.Rows)
	for gridY := 0; gridY < config.Rows; gridY++ {
		for gridX := 0; gridX < config.Columns; gridX++ {
			chunks = append(chunks, worldmap.Chunk{
				GridX:       gridX,
				GridY:       gridY,
				Width:       config.ChunkWidth,
				Height:      config.ChunkHeight,
				PixelWidth:  config.ChunkPixelWidth,
				PixelHeight: config.ChunkPixelHeight,
				EntityIDs:   make([]string, 0),
			})
		}
	}

	world := &worldmap.GlobalScene{Chunks: chunks}
	service.mutex.Lock()
	service.world = world
	service.mutex.Unlock()

	return world, nil
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
				Hunger:       dogInstance.Creature.Genome.HungerMax / 2,
				HungerMax:    dogInstance.Creature.Genome.HungerMax,
				HungerRate:   dogInstance.Creature.Genome.HungerRate,
				Thirst:       dogInstance.Creature.Genome.ThirstMax / 2,
				ThirstMax:    dogInstance.Creature.Genome.ThirstMax,
				ThirstRate:   dogInstance.Creature.Genome.ThirstRate,
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
				ID:         dog.Creature.ID.String(),
				Name:       *dog.Name,
				Species:    dog.Creature.Species,
				Hunger:     dog.Creature.Genome.HungerMax / 2,
				MaxHunger:  dog.Creature.Genome.HungerMax,
				HungerRate: dog.Creature.Genome.HungerRate,
				Thirst:     dog.Creature.Genome.ThirstMax / 2,
				MaxThirst:  dog.Creature.Genome.ThirstMax,
				ThirstRate: dog.Creature.Genome.ThirstRate,
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
			ID:         dog.Creature.ID.String(),
			Name:       *dog.Name,
			Species:    dog.Creature.Species,
			Hunger:     dog.Creature.Genome.HungerMax / 2,
			MaxHunger:  dog.Creature.Genome.HungerMax,
			HungerRate: dog.Creature.Genome.HungerRate,
			Thirst:     dog.Creature.Genome.ThirstMax / 2,
			MaxThirst:  dog.Creature.Genome.ThirstMax,
			ThirstRate: dog.Creature.Genome.ThirstRate,
		})
	}

	return statuses, nil
}
