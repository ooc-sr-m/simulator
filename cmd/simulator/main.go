package main

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"os/signal"
	"sort"
	"strings"
	"sync"
	"syscall"

	"simulator/helper"
	"simulator/services"
	creatureservices "simulator/services/creatures"
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

	start, err := confirmSimulationStart(os.Stdin, os.Stdout)
	if err != nil {
		log.Fatalf("confirm simulation start: %v", err)
	}
	if !start {
		fmt.Println("La simulación no se ha iniciado.")
		return
	}

	groupedLogger := helper.NewGroupedLogger(os.Stdout)
	dogServices := make([]*creatureservices.DogService, 0, len(dogs))
	for _, createdDog := range dogs {
		dogService, err := creatureservices.NewDogService(createdDog, groupedLogger)
		if err != nil {
			log.Fatalf("create dog service: %v", err)
		}
		dogServices = append(dogServices, dogService)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	fmt.Println("\nSimulación iniciada. Pulsa Ctrl+C para detenerla.")

	var waitGroup sync.WaitGroup
	errorsChannel := make(chan error, len(dogServices))
	for _, dogService := range dogServices {
		waitGroup.Add(1)
		go func(service *creatureservices.DogService) {
			defer waitGroup.Done()

			if err := service.Run(ctx); err != nil {
				errorsChannel <- err
				stop()
			}
		}(dogService)
	}

	waitGroup.Wait()
	close(errorsChannel)

	var simulationErrors []error
	for err := range errorsChannel {
		simulationErrors = append(simulationErrors, err)
	}
	if err := errors.Join(simulationErrors...); err != nil {
		log.Fatalf("run simulation: %v", err)
	}

	fmt.Println("\nSimulación detenida.")
}

func confirmSimulationStart(input io.Reader, output io.Writer) (bool, error) {
	if input == nil || output == nil {
		return false, fmt.Errorf("simulation confirmation input/output is not initialized")
	}

	if _, err := fmt.Fprint(output, "\nEl mundo está listo. ¿Quieres iniciar la simulación? [s/N]: "); err != nil {
		return false, fmt.Errorf("write confirmation prompt: %w", err)
	}

	scanner := bufio.NewScanner(input)
	if !scanner.Scan() {
		if err := scanner.Err(); err != nil {
			return false, fmt.Errorf("read confirmation: %w", err)
		}
		return false, nil
	}

	answer := strings.ToLower(strings.TrimSpace(scanner.Text()))
	switch answer {
	case "s", "si", "sí", "y", "yes":
		return true, nil
	default:
		return false, nil
	}
}
