package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	"simulator/models/creatures/dog"
	worldmap "simulator/models/map"
	"simulator/services"
	creatureservices "simulator/services/creatures"
)

const (
	logSubscriberBuffer = 512
	maxLogsPerCreature  = 100
)

type creatureLogEvent struct {
	Time         string `json:"time"`
	Level        string `json:"level"`
	CreatureID   string `json:"creature_id"`
	CreatureName string `json:"creature_name"`
	Scope        string `json:"scope"`
	Message      string `json:"message"`
}

type worldMapSnapshot struct {
	Columns       int             `json:"columns"`
	Rows          int             `json:"rows"`
	CreatureCount int             `json:"creature_count"`
	Chunks        []worldMapChunk `json:"chunks"`
}

type worldMapChunk struct {
	ID          string           `json:"id"`
	GridX       int              `json:"grid_x"`
	GridY       int              `json:"grid_y"`
	Width       float64          `json:"width"`
	Height      float64          `json:"height"`
	PixelWidth  int              `json:"pixel_width"`
	PixelHeight int              `json:"pixel_height"`
	Entities    []worldMapEntity `json:"entities"`
}

type worldMapEntity struct {
	ID      string  `json:"id"`
	Name    string  `json:"name"`
	Species string  `json:"species"`
	X       float64 `json:"x"`
	Y       float64 `json:"y"`
}

type simulationRuntime struct {
	mutex       sync.Mutex
	world       *worldmap.GlobalScene
	dogs        []*dog.Dog
	dogServices map[string]*creatureservices.DogService
	creatures   []services.CreatureInitializationEvent
	running     bool
	history     map[string][]creatureLogEvent
	subscribers map[chan creatureLogEvent]struct{}
}

func newSimulationRuntime() *simulationRuntime {
	return &simulationRuntime{
		dogServices: make(map[string]*creatureservices.DogService),
		history:     make(map[string][]creatureLogEvent),
		subscribers: make(map[chan creatureLogEvent]struct{}),
	}
}

func (runtime *simulationRuntime) setWorld(world *worldmap.GlobalScene, dogs []*dog.Dog, creatures []services.CreatureInitializationEvent) {
	runtime.mutex.Lock()
	defer runtime.mutex.Unlock()

	runtime.world = world
	runtime.dogs = append([]*dog.Dog(nil), dogs...)
	runtime.creatures = append([]services.CreatureInitializationEvent(nil), creatures...)
}

func (runtime *simulationRuntime) mapHandler(response http.ResponseWriter, request *http.Request) {
	snapshot, err := runtime.getMapSnapshot()
	if err != nil {
		http.Error(response, err.Error(), http.StatusConflict)
		return
	}

	response.Header().Set("Content-Type", "application/json")
	response.Header().Set("Cache-Control", "no-store")
	_ = json.NewEncoder(response).Encode(snapshot)
}

func (runtime *simulationRuntime) getMapSnapshot() (*worldMapSnapshot, error) {
	runtime.mutex.Lock()
	defer runtime.mutex.Unlock()

	if runtime.world == nil || len(runtime.world.Chunks) == 0 {
		return nil, fmt.Errorf("world map has not been initialized")
	}

	dogsByID := make(map[string]*dog.Dog, len(runtime.dogs))
	for _, currentDog := range runtime.dogs {
		dogsByID[currentDog.Creature.ID.String()] = currentDog
	}

	snapshot := &worldMapSnapshot{
		CreatureCount: len(runtime.dogs),
		Chunks:        make([]worldMapChunk, 0, len(runtime.world.Chunks)),
	}
	for _, chunk := range runtime.world.Chunks {
		if chunk.GridX+1 > snapshot.Columns {
			snapshot.Columns = chunk.GridX + 1
		}
		if chunk.GridY+1 > snapshot.Rows {
			snapshot.Rows = chunk.GridY + 1
		}

		mapChunk := worldMapChunk{
			ID:          chunk.ID.String(),
			GridX:       chunk.GridX,
			GridY:       chunk.GridY,
			Width:       chunk.Width,
			Height:      chunk.Height,
			PixelWidth:  chunk.PixelWidth,
			PixelHeight: chunk.PixelHeight,
			Entities:    make([]worldMapEntity, 0, len(chunk.EntityIDs)),
		}
		for _, entityID := range chunk.EntityIDs {
			currentDog := dogsByID[entityID]
			if currentDog == nil {
				continue
			}
			name := "Unnamed creature"
			if currentDog.Name != nil {
				name = *currentDog.Name
			}
			mapChunk.Entities = append(mapChunk.Entities, worldMapEntity{
				ID:      entityID,
				Name:    name,
				Species: currentDog.Creature.Species,
				X:       currentDog.Creature.Position.X,
				Y:       currentDog.Creature.Position.Y,
			})
		}
		snapshot.Chunks = append(snapshot.Chunks, mapChunk)
	}

	return snapshot, nil
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
	dogServices := make(map[string]*creatureservices.DogService, len(dogs))
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
		dogServices[createdDog.Creature.ID.String()] = dogService
	}
	runtime.dogServices = dogServices
	runtime.running = true
	runtime.mutex.Unlock()

	for _, createdDog := range dogs {
		dogService := dogServices[createdDog.Creature.ID.String()]
		go func(currentDog *dog.Dog, service *creatureservices.DogService) {
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
		}(createdDog, dogService)
	}

	return true, nil
}

func (runtime *simulationRuntime) publish(event creatureLogEvent) {
	runtime.mutex.Lock()
	defer runtime.mutex.Unlock()

	creatureHistory := append(runtime.history[event.CreatureID], event)
	if len(creatureHistory) > maxLogsPerCreature {
		creatureHistory = creatureHistory[len(creatureHistory)-maxLogsPerCreature:]
	}
	runtime.history[event.CreatureID] = creatureHistory

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
	history := make([]creatureLogEvent, 0, len(runtime.history)*maxLogsPerCreature)
	for _, creatureHistory := range runtime.history {
		history = append(history, creatureHistory...)
	}

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

func (runtime *simulationRuntime) getCreatureStatus(creatureID string) (*services.CreatureStatus, error) {
	runtime.mutex.Lock()
	var selectedDog *dog.Dog
	for _, currentDog := range runtime.dogs {
		if currentDog.Creature.ID.String() == creatureID {
			selectedDog = currentDog
			break
		}
	}
	dogService := runtime.dogServices[creatureID]
	runtime.mutex.Unlock()

	if selectedDog == nil {
		return nil, fmt.Errorf("creature not found")
	}

	hunger := selectedDog.Creature.Genome.HungerMax / 2
	hungerMax := selectedDog.Creature.Genome.HungerMax
	hungerRate := selectedDog.Creature.Genome.HungerRate
	thirst := selectedDog.Creature.Genome.ThirstMax / 2
	thirstMax := selectedDog.Creature.Genome.ThirstMax
	thirstRate := selectedDog.Creature.Genome.ThirstRate
	if dogService != nil {
		var err error
		hunger, hungerMax, hungerRate, thirst, thirstMax, thirstRate, err = dogService.GetMetabolismState()
		if err != nil {
			return nil, fmt.Errorf("get dog metabolism: %w", err)
		}
	}

	name := "Unnamed dog"
	if selectedDog.Name != nil {
		name = *selectedDog.Name
	}
	return &services.CreatureStatus{
		ID:         selectedDog.Creature.ID.String(),
		Name:       name,
		Species:    selectedDog.Creature.Species,
		Hunger:     hunger,
		MaxHunger:  hungerMax,
		HungerRate: hungerRate,
		Thirst:     thirst,
		MaxThirst:  thirstMax,
		ThirstRate: thirstRate,
	}, nil
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
		logger.runtime.publish(creatureLogEvent{
			Time:         time.Now().Format("15:04:05.000"),
			Level:        "INFO",
			CreatureID:   logger.creatureID,
			CreatureName: logger.creatureName,
			Scope:        group,
			Message:      message,
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
