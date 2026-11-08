package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"simulator/helper"
	creatureservices "simulator/services/creatures"
	"simulator/models/creatures/dog"
	"simulator/services"
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
	Hunger       int    `json:"hunger,omitempty"`
	Thirst       int    `json:"thirst,omitempty"`
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
	runtime.running = true
	runtime.mutex.Unlock()

	for _, createdDog := range dogs {
		go func(dog *dog.Dog) {
			logger := creatureEventLogger{
				runtime:      runtime,
				creatureID:   dog.Creature.ID.String(),
				creatureName: "Unnamed dog",
			}
			if dog.Name != nil {
				logger.creatureName = *dog.Name
			}

			dogService, err := creatureservices.NewDogService(dog, logger)
			if err != nil {
				runtime.publish(creatureLogEvent{
					Time:         time.Now().Format("15:04:05.000"),
					Level:        "ERROR",
					CreatureID:   dog.Creature.ID.String(),
					CreatureName: logger.creatureName,
					Scope:        "DOG SERVICE",
					Message:      fmt.Sprintf("create service: %v", err),
				})
				return
			}

			if err := dogService.Run(context.Background()); err != nil {
				runtime.publish(creatureLogEvent{
					Time:         time.Now().Format("15:04:05.000"),
					Level:        "ERROR",
					CreatureID:   dog.Creature.ID.String(),
					CreatureName: logger.creatureName,
					Scope:        "DOG SERVICE",
					Message:      err.Error(),
				})
			}
		}(createdDog)
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
		delete(runtime.subscribers, updates)
		runtime.mutex.Unlock()
		close(updates)
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
		Started bool `json:"started"`
	}{Started: started})
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

	subscriber, history, unsubscribe := runtime.subscribe()
	defer unsubscribe()

	for _, event := range history {
		if err := writeEvent(response, flusher, "creature-log", event); err != nil {
			return
		}
	}

	for event := range subscriber {
		if err := writeEvent(response, flusher, "creature-log", event); err != nil {
			return
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

	for _, message := range messages {
		hunger, thirst := 0, 0
		
		// Parse hunger and thirst from the message format: "hambre: X/Y | sed: A/B"
		if idx := strings.Index(message, "hambre:"); idx >= 0 {
			parts := strings.Split(message[idx:], "|")
			for _, part := range parts {
				part = strings.TrimSpace(part)
				if strings.HasPrefix(part, "hambre:") {
					fmt.Sscanf(strings.TrimPrefix(part, "hambre:"), "%d/%d", &hunger, &thirst)
				} else if strings.HasPrefix(part, "sed:") {
					fmt.Sscanf(strings.TrimPrefix(part, "sed:"), "%d/%d", &thirst, &hunger)
				}
			}
		}

		logger.runtime.publish(creatureLogEvent{
			Time:         time.Now().Format("15:04:05.000"),
			Level:        "INFO",
			CreatureID:   logger.creatureID,
			CreatureName: logger.creatureName,
			Scope:        group,
			Message:      message,
			Hunger:       hunger,
			Thirst:       thirst,
		})
	}

	return nil
}

func writeEvent(destination io.Writer, flusher http.Flusher, eventName string, payload any) error {
	encodedPayload, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("encode %s event: %w", eventName, err)
	}
	if _, err := fmt.Fprintf(destination, "event: %s\ndata: %s\n\n", eventName, encodedPayload); err != nil {
		return fmt.Errorf("write %s event: %w", eventName, err)
	}
	flusher.Flush()
	return nil
}
