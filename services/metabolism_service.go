package services

import (
	"fmt"

	"simulator/models/creatures"
)

// MetabolismService gestiona el metabolismo de una criatura
type MetabolismService struct {
	genome     *creatures.Genome
	hunger     int
	thirst     int
}

// NewMetabolismService crea un nuevo servicio de metabolismo
func NewMetabolismService(genome *creatures.Genome) (*MetabolismService, error) {
	if genome == nil {
		return nil, fmt.Errorf("genome is not initialized")
	}

	return &MetabolismService{
		genome: genome,
		hunger: genome.HungerMax / 2,
		thirst: genome.ThirstMax / 2,
	}, nil
}

// Update actualiza el estado del metabolismo
func (s *MetabolismService) Update() error {
	if s.genome == nil {
		return fmt.Errorf("metabolism service is not initialized")
	}

	s.hunger = min(s.hunger+int(s.genome.HungerRate*100), s.genome.HungerMax)
	s.thirst = min(s.thirst+int(s.genome.ThirstRate*100), s.genome.ThirstMax)

	return nil
}

// GetCurrentState devuelve el estado actual de hambre y sed
func (s *MetabolismService) GetCurrentState() (hunger, thirst int) {
	return s.hunger, s.thirst
}

// GetMaxHunger devuelve el máximo de hambre
func (s *MetabolismService) GetMaxHunger() int {
	if s.genome == nil {
		return 0
	}
	return s.genome.HungerMax
}

// GetMaxThirst devuelve el máximo de sed
func (s *MetabolismService) GetMaxThirst() int {
	if s.genome == nil {
		return 0
	}
	return s.genome.ThirstMax
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
