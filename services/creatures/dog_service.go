package creatures

import (
	"context"
	"fmt"
	"strings"
	"time"

	"simulator/helper"
	"simulator/models/creatures/dog"
)

// DogService owns the business logic and state updates for one dog.
// A different instance must be created for every dog in the world.
type DogService struct {
	dog           *dog.Dog
	logger        helper.GroupLogger
	metabolismSvc *MetabolismService
}

const dogUpdateInterval = time.Second

func NewDogService(createdDog *dog.Dog, logger helper.GroupLogger) (*DogService, error) {
	if createdDog == nil {
		return nil, fmt.Errorf("dog is not initialized")
	}
	if createdDog.Name == nil || strings.TrimSpace(*createdDog.Name) == "" {
		return nil, fmt.Errorf("dog name is empty")
	}
	if logger == nil {
		return nil, fmt.Errorf("dog logger is not initialized")
	}

	metabolismSvc, err := NewMetabolismService(createdDog.Creature.Genome)
	if err != nil {
		return nil, fmt.Errorf("create metabolism service: %w", err)
	}

	return &DogService{
		dog:           createdDog,
		logger:        logger,
		metabolismSvc: metabolismSvc,
	}, nil
}

// Update advances this dog's state by one simulation step.
func (service *DogService) Update() error {
	if service == nil || service.dog == nil || service.logger == nil {
		return fmt.Errorf("dog service is not initialized")
	}
	if service.dog.Name == nil || strings.TrimSpace(*service.dog.Name) == "" {
		return fmt.Errorf("dog name is empty")
	}

	if err := service.metabolismSvc.Update(); err != nil {
		return fmt.Errorf("update metabolism: %w", err)
	}

	name := strings.TrimSpace(*service.dog.Name)
	hunger, thirst := service.metabolismSvc.GetCurrentState()
	group := fmt.Sprintf("perro %s · %s", name, service.dog.Creature.ID)
	message := fmt.Sprintf("hola mundo, soy el perro %s | hambre: %d/%d | sed: %d/%d",
		name, hunger, service.metabolismSvc.GetMaxHunger(), thirst, service.metabolismSvc.GetMaxThirst())

	if err := service.logger.Log(group, message); err != nil {
		return fmt.Errorf("update dog %s: %w", name, err)
	}

	return nil
}

// GetMetabolismState returns the current state used by the simulation loop.
func (service *DogService) GetMetabolismState() (hunger, hungerMax int, hungerRate float64, thirst, thirstMax int, thirstRate float64, err error) {
	if service == nil || service.metabolismSvc == nil {
		return 0, 0, 0, 0, 0, 0, fmt.Errorf("dog metabolism service is not initialized")
	}

	hunger, thirst = service.metabolismSvc.GetCurrentState()
	hungerMax = service.metabolismSvc.GetMaxHunger()
	thirstMax = service.metabolismSvc.GetMaxThirst()
	hungerRate, thirstRate = service.metabolismSvc.GetRates()
	return hunger, hungerMax, hungerRate, thirst, thirstMax, thirstRate, nil
}

// Run keeps this dog alive until the simulation is stopped. The first update
// happens immediately and subsequent updates happen once per second.
func (service *DogService) Run(ctx context.Context) error {
	if ctx == nil {
		return fmt.Errorf("dog service context is not initialized")
	}

	if err := service.Update(); err != nil {
		return err
	}

	ticker := time.NewTicker(dogUpdateInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
			if err := service.Update(); err != nil {
				return err
			}
		}
	}
}
