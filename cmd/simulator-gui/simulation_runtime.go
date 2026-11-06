package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"simulator/models/creatures/dog"
	"simulator/services"
	creatureservices "simulator/services/creatures"
)

const (
	logSubscriberBuffer = 512
	maxLogHistory       = 2500
)

type creatureLogEvent struct {
	Time         string `json:"time"`
	Level        string `json:"level"`
	CreatureID   string `json:"creature_id"`
	CreatureName string `json:"creature_name"`
	Scope        string `json:"scope"`
	Message      string `json:"message"`
}

type simulationRuntime struct {
	mutex       sync.Mutex
	dogs        []*dog.Dog
	creatures   []services.CreatureInitializationEvent
	running     bool
	history     []creatureLogEvent
	subscribers map[chan creatureLogEvent]struct{}
}

func newSimulationRuntime() *simulationRuntime {
	return &simulationRuntime{
		subscribers: make(map[chan creatureLogEvent]struct{}),
	}
}

func (runtime *simulationRuntime) setWorld(dogs []*dog.Dog, creatures []services.CreatureInitializationEvent) {
	runtime.mutex.Lock()
	defer runtime.mutex.Unlock()

	runtime.dogs = append([]*dog.Dog(nil), dogs...)
	runtime.creatures = append([]services.CreatureInitializationEvent(nil), creatures...)
}

func (runtime *simulationRuntime) worldSnapshot() ([]services.CreatureInitializationEvent, bool) {
	runtime.mutex.Lock()
	defer runtime.mutex.Unlock()

	return append([]services.CreatureInitializationEvent(nil), runtime.creatures...), runtime.running
}

func (runtime *simulationRuntime) start() (bool, error) {
	runtime.mutex.Lock()
	if len(runtime.dogs) == 0 {
		runtime.mutex.Unlock()
		return false, fmt.Errorf("world has not been initialized")
	}
	if runtime.running {
		runtime.mutex.Unlock()
		return false, nil
	}

	dogs := append([]*dog.Dog(nil), runtime.dogs...)
	dogServices := make([]*creatureservices.DogService, 0, len(dogs))
	for _, createdDog := range dogs {
		name := "Unnamed dog"
		if createdDog.Name != nil {
			name = *createdDog.Name
		}
		logger := creatureEventLogger{
			runtime:      runtime,
			creatureID:   createdDog.Creature.ID.String(),
			creatureName: name,
		}
		dogService, err := creatureservices.NewDogService(createdDog, logger)
		if err != nil {
			runtime.mutex.Unlock()
			return false, fmt.Errorf("create dog service: %w", err)
		}
		dogServices = append(dogServices, dogService)
	}
	runtime.running = true
	runtime.mutex.Unlock()

	for index, dogService := range dogServices {
		createdDog := dogs[index]
		go func(service *creatureservices.DogService, currentDog *dog.Dog) {
			if err := service.Run(context.Background()); err != nil {
				name := "Unnamed dog"
				if currentDog.Name != nil {
					name = *currentDog.Name
				}
				runtime.publish(creatureLogEvent{
					Time:         time.Now().Format("15:04:05.000"),
					Level:        "ERROR",
					CreatureID:   currentDog.Creature.ID.String(),
					CreatureName: name,
					Scope:        "DOG SERVICE",
					Message:      err.Error(),
				})
			}
		}(dogService, createdDog)
	}

	return true, nil
}

func (runtime *simulationRuntime) publish(event creatureLogEvent) {
	runtime.mutex.Lock()
	defer runtime.mutex.Unlock()

	runtime.history = append(runtime.history, event)
	if len(runtime.history) > maxLogHistory {
		runtime.history = append([]creatureLogEvent(nil), runtime.history[len(runtime.history)-maxLogHistory:]...)
	}

	for subscriber := range runtime.subscribers {
		select {
		case subscriber <- event:
		default:
		}
	}
}

func (runtime *simulationRuntime) subscribe() (<-chan creatureLogEvent, []creatureLogEvent, func()) {
	runtime.mutex.Lock()
	defer runtime.mutex.Unlock()

	updates := make(chan creatureLogEvent, logSubscriberBuffer)
	runtime.subscribers[updates] = struct{}{}
	history := append([]creatureLogEvent(nil), runtime.history...)

	unsubscribe := func() {
		runtime.mutex.Lock()
		defer runtime.mutex.Unlock()
		delete(runtime.subscribers, updates)
	}

	return updates, history, unsubscribe
}

func (runtime *simulationRuntime) startHandler(response http.ResponseWriter, request *http.Request) {
	started, err := runtime.start()
	if err != nil {
		http.Error(response, err.Error(), http.StatusConflict)
		return
	}

	response.Header().Set("Content-Type", "application/json")
	response.Header().Set("Cache-Control", "no-store")
	_ = json.NewEncoder(response).Encode(struct {
		Running bool `json:"running"`
		Started bool `json:"started"`
	}{Running: true, Started: started})
}

func (runtime *simulationRuntime) logsHandler(response http.ResponseWriter, request *http.Request) {
	flusher, ok := response.(http.Flusher)
	if !ok {
		http.Error(response, "streaming is unavailable", http.StatusInternalServerError)
		return
	}

	response.Header().Set("Content-Type", "text/event-stream")
	response.Header().Set("Cache-Control", "no-cache")
	response.Header().Set("Connection", "keep-alive")
	response.Header().Set("X-Content-Type-Options", "nosniff")

	updates, history, unsubscribe := runtime.subscribe()
	defer unsubscribe()

	for _, event := range history {
		if err := writeEvent(response, flusher, "creature-log", event); err != nil {
			return
		}
	}

	heartbeat := time.NewTicker(15 * time.Second)
	defer heartbeat.Stop()

	for {
		select {
		case <-request.Context().Done():
			return
		case event := <-updates:
			if err := writeEvent(response, flusher, "creature-log", event); err != nil {
				return
			}
		case <-heartbeat.C:
			if _, err := fmt.Fprint(response, ": keep-alive\n\n"); err != nil {
				return
			}
			flusher.Flush()
		}
	}
}

type creatureEventLogger struct {
	runtime      *simulationRuntime
	creatureID   string
	creatureName string
}

func (logger creatureEventLogger) Log(group string, messages ...string) error {
	if logger.runtime == nil {
		return fmt.Errorf("simulation runtime is not initialized")
	}
	_ = group

	for _, message := range messages {
		logger.runtime.publish(creatureLogEvent{
			Time:         time.Now().Format("15:04:05.000"),
			Level:        "INFO",
			CreatureID:   logger.creatureID,
			CreatureName: logger.creatureName,
			Scope:        "DOG",
			Message:      message,
		})
	}

	return nil
}
