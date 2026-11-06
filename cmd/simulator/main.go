package main

import (
	"fmt"
	"log"
	"sort"

	"simulator/services"
)

func main() {
	worldService := services.NewWorldService(42)

	dogs, err := worldService.CreateCreatures()
	if err != nil {
		log.Fatalf("create creatures: %v", err)
	}

	fmt.Printf("%d dogs created successfully\n", len(dogs))
	for _, createdDog := range dogs {
		capabilities := make([]string, 0, len(createdDog.Creature.Capabilities))
		for capability, enabled := range createdDog.Creature.Capabilities {
			if enabled {
				capabilities = append(capabilities, capability)
			}
		}
		sort.Strings(capabilities)

		fmt.Printf("\nName: %s\n", *createdDog.Name)
		fmt.Printf("Breed: %s\n", createdDog.Breed)
		fmt.Printf("ID: %s\n", createdDog.Creature.ID)
		fmt.Printf("Species: %s\n", createdDog.Creature.Species)
		fmt.Printf("Hunger maximum: %d\n", createdDog.Creature.Genome.HungerMax)
		fmt.Printf("Hunger rate: %.4f\n", createdDog.Creature.Metabolism.HungerRate)
		fmt.Printf("Capabilities: %v\n", capabilities)
	}
}
