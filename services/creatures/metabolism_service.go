package creatures

import (
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	creaturemodels "simulator/models/creatures"
	"simulator/models/physiological_capacities"
)

const defaultMetabolismRate = 0.001

func CalculateNewHunger(currentHunger, hungerRate float64, deltaTimeSeconds float64, hungerMax int) int {
	if hungerMax <= 0 {
		hungerMax = 100
	}
	if hungerRate <= 0 {
		hungerRate = defaultMetabolismRate
	}

	return int(calculateNewLevel(currentHunger, hungerRate, deltaTimeSeconds, hungerMax))
}

func CalculateNewThirst(currentThirst, thirstRate float64, deltaTimeSeconds float64, thirstMax int) int {
	if thirstMax <= 0 {
		thirstMax = 100
	}
	if thirstRate <= 0 {
		thirstRate = defaultMetabolismRate
	}

	return int(calculateNewLevel(currentThirst, thirstRate, deltaTimeSeconds, thirstMax))
}

func calculateNewLevel(current, rate, deltaTimeSeconds float64, maximum int) float64 {
	updated := current + rate*deltaTimeSeconds
	if updated > float64(maximum) {
		return float64(maximum)
	}
	return updated
}

func ValidateMetabolism(metabolism *physiological_capacities.Metabolism) error {
	if metabolism == nil {
		return errors.New("metabolism cannot be nil")
	}

	if metabolism.Hunger < 0 || metabolism.Hunger > metabolism.HungerMax {
		return fmt.Errorf("hunger %d out of bounds [0, %d]", metabolism.Hunger, metabolism.HungerMax)
	}
	if metabolism.Thirst < 0 || metabolism.Thirst > metabolism.ThirstMax {
		return fmt.Errorf("thirst %d out of bounds [0, %d]", metabolism.Thirst, metabolism.ThirstMax)
	}
	if metabolism.HungerRate < 0 {
		return errors.New("hunger rate cannot be negative")
	}
	if metabolism.ThirstRate < 0 {
		return errors.New("thirst rate cannot be negative")
	}

	return nil
}

type MetabolismService struct {
	mutex       sync.RWMutex
	genome      *creaturemodels.Genome
	hunger      float64
	thirst      float64
	lastHunger  time.Time
	lastThirst  time.Time
	updateCount atomic.Int64
}

func NewMetabolismService(genome *creaturemodels.Genome) (*MetabolismService, error) {
	if genome == nil {
		return nil, fmt.Errorf("genome is not initialized")
	}

	now := time.Now()
	return &MetabolismService{
		genome:     genome,
		hunger:     float64(genome.HungerMax) / 2,
		thirst:     float64(genome.ThirstMax) / 2,
		lastHunger: now,
		lastThirst: now,
	}, nil
}

func (service *MetabolismService) UpdateHunger() error {
	service.mutex.Lock()
	defer service.mutex.Unlock()

	if service.genome == nil {
		return errors.New("metabolism is not initialized")
	}

	now := time.Now()
	deltaSeconds := now.Sub(service.lastHunger).Seconds()
	service.hunger = calculateNewLevel(service.hunger, service.genome.HungerRate, deltaSeconds, service.genome.HungerMax)
	service.lastHunger = now
	service.updateCount.Add(1)
	return nil
}

func (service *MetabolismService) UpdateThirst() error {
	service.mutex.Lock()
	defer service.mutex.Unlock()

	if service.genome == nil {
		return errors.New("metabolism is not initialized")
	}

	now := time.Now()
	deltaSeconds := now.Sub(service.lastThirst).Seconds()
	service.thirst = calculateNewLevel(service.thirst, service.genome.ThirstRate, deltaSeconds, service.genome.ThirstMax)
	service.lastThirst = now
	service.updateCount.Add(1)
	return nil
}

func (service *MetabolismService) Update() error {
	service.mutex.Lock()
	defer service.mutex.Unlock()

	if service.genome == nil {
		return errors.New("metabolism is not initialized")
	}

	now := time.Now()
	hungerDeltaSeconds := now.Sub(service.lastHunger).Seconds()
	thirstDeltaSeconds := now.Sub(service.lastThirst).Seconds()
	service.hunger = calculateNewLevel(service.hunger, service.genome.HungerRate, hungerDeltaSeconds, service.genome.HungerMax)
	service.thirst = calculateNewLevel(service.thirst, service.genome.ThirstRate, thirstDeltaSeconds, service.genome.ThirstMax)
	service.lastHunger = now
	service.lastThirst = now
	service.updateCount.Add(1)
	return nil
}

func (service *MetabolismService) GetCurrentState() (hunger, thirst int) {
	service.mutex.RLock()
	defer service.mutex.RUnlock()

	return int(service.hunger), int(service.thirst)

}

func (service *MetabolismService) GetMaxHunger() int {
	service.mutex.RLock()
	defer service.mutex.RUnlock()

	if service.genome == nil {
		return 0
	}
	return service.genome.HungerMax
}

func (service *MetabolismService) GetMaxThirst() int {
	service.mutex.RLock()
	defer service.mutex.RUnlock()

	if service.genome == nil {
		return 0
	}
	return service.genome.ThirstMax
}

func (service *MetabolismService) GetRates() (hungerRate, thirstRate float64) {
	service.mutex.RLock()
	defer service.mutex.RUnlock()

	if service.genome == nil {
		return 0, 0
	}
	return service.genome.HungerRate, service.genome.ThirstRate
}

func (service *MetabolismService) UpdateCount() int64 {
	return service.updateCount.Load()
}
