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
	dog    *dog.Dog
	logger helper.GroupLogger
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

	return &DogService{
		dog:    createdDog,
		logger: logger,
	}, nil
}

// Update advances this dog's state by one simulation step.
// It only logs a greeting for now; future dog-specific rules belong here.
func (service *DogService) Update() error {
	if service == nil || service.dog == nil || service.logger == nil {
		return fmt.Errorf("dog service is not initialized")
	}
	if service.dog.Name == nil || strings.TrimSpace(*service.dog.Name) == "" {
		return fmt.Errorf("dog name is empty")
	}

	name := strings.TrimSpace(*service.dog.Name)
	group := fmt.Sprintf("perro %s · %s", name, service.dog.Creature.ID)
	message := fmt.Sprintf("hola mundo, soy el perro %s", name)

	if err := service.logger.Log(group, message); err != nil {
		return fmt.Errorf("update dog %s: %w", name, err)
	}

	return nil
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
