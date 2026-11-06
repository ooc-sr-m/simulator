package services

import (
	"fmt"
	"testing"

	"simulator/models/creatures/dog"
)

func TestCreateCreaturesCreatesTenDogsWithinSpeciesRanges(t *testing.T) {
	service := NewWorldService(42)

	dogs, err := service.CreateCreatures()
	if err != nil {
		t.Fatalf("CreateCreatures() error = %v", err)
	}
	if len(dogs) != defaultCreatureCount {
		t.Fatalf("CreateCreatures() created %d dogs, want %d", len(dogs), defaultCreatureCount)
	}

	profile := dog.GenomeProfile()
	ids := make(map[string]struct{}, len(dogs))
	for index, createdDog := range dogs {
		if createdDog == nil {
			t.Fatalf("CreateCreatures() dog %d is nil", index+1)
		}
		if createdDog.Name == nil || *createdDog.Name != fmt.Sprintf("Dog %d", index+1) {
			t.Errorf("CreateCreatures() dog %d name = %v, want Dog %d", index+1, createdDog.Name, index+1)
		}
		if createdDog.Creature.Species != profile.ID {
			t.Errorf("CreateCreatures() dog %d species = %q, want %q", index+1, createdDog.Creature.Species, profile.ID)
		}
		if createdDog.Creature.Genome == nil {
			t.Errorf("CreateCreatures() dog %d genome is nil", index+1)
			continue
		}
		if hungerMax := createdDog.Creature.Genome.HungerMax; hungerMax < profile.HungerMax.Min || hungerMax > profile.HungerMax.Max {
			t.Errorf("CreateCreatures() dog %d hunger max = %d, want within [%d, %d]", index+1, hungerMax, profile.HungerMax.Min, profile.HungerMax.Max)
		}
		if hungerRate := createdDog.Creature.Genome.HungerRate; hungerRate < profile.HungerRate.Min || hungerRate > profile.HungerRate.Max {
			t.Errorf("CreateCreatures() dog %d hunger rate = %f, want within [%f, %f]", index+1, hungerRate, profile.HungerRate.Min, profile.HungerRate.Max)
		}
		if createdDog.Creature.Metabolism == nil || createdDog.Creature.Metabolism.HungerRate != createdDog.Creature.Genome.HungerRate {
			t.Errorf("CreateCreatures() dog %d metabolism does not match its genome", index+1)
		}

		id := createdDog.Creature.ID.String()
		if _, exists := ids[id]; exists {
			t.Errorf("CreateCreatures() generated duplicate ID %s", id)
		}
		ids[id] = struct{}{}
	}
}

func TestCreateCreaturesRequiresInitializedService(t *testing.T) {
	var service *WorldService

	if _, err := service.CreateCreatures(); err == nil {
		t.Fatal("CreateCreatures() error = nil, want initialization error")
	}
}
