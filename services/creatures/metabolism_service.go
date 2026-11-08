package creatures

import (
	"errors"
	"fmt"
	"sync/atomic"
	"time"

	"simulator/models/physiological_capacities"
)

func CalculateNewHunger(currentHunger, hungerRate float64, deltaTimeSeconds float64, hungerMax int) int {
	if hungerMax <= 0 {
		hungerMax = 100
	}
	if hungerRate <= 0 {
		hungerRate = 0.1
	}

	newHunger := currentHunger + (hungerRate * deltaTimeSeconds)

	if newHunger > float64(hungerMax) {
		newHunger = float64(hungerMax)
	}

	return int(newHunger)
}

func CalculateNewThirst(currentThirst, thirstRate float64, deltaTimeSeconds float64, thirstMax int) int {
	if thirstMax <= 0 {
		thirstMax = 100
	}
	if thirstRate <= 0 {
		thirstRate = 0.1
	}

	newThirst := currentThirst + (thirstRate * deltaTimeSeconds)

	if newThirst > float64(thirstMax) {
		newThirst = float64(thirstMax)
	}

	return int(newThirst)
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
	metabolism   *physiological_capacities.Metabolism
	lastUpdate   time.Time
	updateCount  atomic.Int64
}

func NewMetabolismService(metabolism *physiological_capacities.Metabolism) (*MetabolismService, error) {
	if err := ValidateMetabolism(metabolism); err != nil {
		return nil, fmt.Errorf("invalid metabolism: %w", err)
	}

	return &MetabolismService{
		metabolism: metabolism,
		lastUpdate: time.Now(),
	}, nil
}

func (service *MetabolismService) UpdateHunger() error {
	if service.metabolism == nil {
		return errors.New("metabolism is not initialized")
	}

	deltaSeconds := time.Since(service.lastUpdate).Seconds()
	service.lastUpdate = time.Now()

	currentHunger := float64(service.metabolism.Hunger)
	newHunger := CalculateNewHunger(currentHunger, service.metabolism.HungerRate, deltaSeconds, service.metabolism.HungerMax)
	service.metabolism.Hunger = newHunger
	service.updateCount.Add(1)

	return nil
}

func (service *MetabolismService) UpdateThirst() error {
	if service.metabolism == nil {
		return errors.New("metabolism is not initialized")
	}

	deltaSeconds := time.Since(service.lastUpdate).Seconds()
	service.lastUpdate = time.Now()

	currentThirst := float64(service.metabolism.Thirst)
	newThirst := CalculateNewThirst(currentThirst, service.metabolism.ThirstRate, deltaSeconds, service.metabolism.ThirstMax)
	service.metabolism.Thirst = newThirst
	service.updateCount.Add(1)

	return nil
}

func (service *MetabolismService) Update() error {
	if err := service.UpdateHunger(); err != nil {
		return fmt.Errorf("update hunger: %w", err)
	}
	if err := service.UpdateThirst(); err != nil {
		return fmt.Errorf("update thirst: %w", err)
	}
	return nil
}

func (service *MetabolismService) GetCurrentState() (hunger, thirst int) {
	if service.metabolism != nil {
		return service.metabolism.Hunger, service.metabolism.Thirst
	}
	return 0, 0
}

func (service *MetabolismService) UpdateCount() int64 {
	return service.updateCount.Load()
}
