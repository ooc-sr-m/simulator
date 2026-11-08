package main

import (
	"embed"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"simulator/services"
)

const (
	worldSeed     int64 = 42
	creatureDelay       = 140 * time.Millisecond
)

//go:embed index.html styles.css app.js
var assets embed.FS

type guiConfig struct {
	newWorld   func() *services.WorldService
	eventDelay time.Duration
}

type completionEvent struct {
	Count   int  `json:"count"`
	Running bool `json:"running"`
}

type statusEvent struct {
	Message string `json:"message"`
}

type logEvent struct {
	Time    string `json:"time"`
	Level   string `json:"level"`
	Scope   string `json:"scope"`
	Message string `json:"message"`
}

func main() {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		log.Fatalf("start graphical simulator: %v", err)
	}

	url := "http://" + listener.Addr().String()
	server := &http.Server{
		Handler: newHandler(guiConfig{
			newWorld: func() *services.WorldService {
				return services.NewWorldService(worldSeed)
			},
			eventDelay: creatureDelay,
		}),
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		if err := openBrowser(url); err != nil {
			log.Printf("Could not open the browser automatically: %v", err)
			log.Printf("Open %s manually", url)
		}
	}()

	log.Printf("Graphical simulator available at %s", url)
	if err := server.Serve(listener); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatalf("serve graphical simulator: %v", err)
	}
}

func newHandler(config guiConfig) http.Handler {
	simRuntime := newSimulationRuntime()
	mux := http.NewServeMux()
	mux.HandleFunc("GET /{$}", serveAsset("index.html", "text/html; charset=utf-8"))
	mux.HandleFunc("GET /world", serveAsset("index.html", "text/html; charset=utf-8"))
	mux.HandleFunc("GET /styles.css", serveAsset("styles.css", "text/css; charset=utf-8"))
	mux.HandleFunc("GET /app.js", serveAsset("app.js", "text/javascript; charset=utf-8"))
	mux.HandleFunc("GET /api/world/initialize", initializationHandler(config, simRuntime))
	mux.HandleFunc("POST /api/world/start", simRuntime.startHandler)
	mux.HandleFunc("GET /api/world/logs", simRuntime.logsHandler)
	mux.HandleFunc("GET /api/world/status/{id}", statusHandler(simRuntime))

	return mux
}

func serveAsset(name string, contentType string) http.HandlerFunc {
	return func(response http.ResponseWriter, request *http.Request) {
		response.Header().Set("Cache-Control", "no-store")
		response.Header().Set("Content-Type", contentType)
		asset, err := assets.ReadFile(name)
		if err != nil {
			http.Error(response, "graphical interface is unavailable", http.StatusInternalServerError)
			return
		}
		_, _ = response.Write(asset)
	}
}

func initializationHandler(config guiConfig, simRuntime *simulationRuntime) http.HandlerFunc {
	return func(response http.ResponseWriter, request *http.Request) {
		flusher, ok := response.(http.Flusher)
		if !ok {
			http.Error(response, "streaming is unavailable", http.StatusInternalServerError)
			return
		}

		response.Header().Set("Content-Type", "text/event-stream")
		response.Header().Set("Cache-Control", "no-cache")
		response.Header().Set("Connection", "keep-alive")
		response.Header().Set("X-Content-Type-Options", "nosniff")

		if existingCreatures, running := simRuntime.worldSnapshot(); len(existingCreatures) > 0 {
			status := "World ready to start"
			if running {
				status = "World simulation online"
			}
			if err := writeEvent(response, flusher, "status", statusEvent{Message: status}); err != nil {
				return
			}
			for _, event := range existingCreatures {
				if err := writeEvent(response, flusher, "creature", event); err != nil {
					return
				}
			}
			_ = writeEvent(response, flusher, "complete", completionEvent{Count: len(existingCreatures), Running: running})
			return
		}

		if err := writeEvent(response, flusher, "status", statusEvent{Message: "Preparing ecosystem"}); err != nil {
			return
		}
		if err := writeEvent(response, flusher, "log", newLogEvent("INFO", "BOOT", "World initialization sequence started")); err != nil {
			return
		}
		if err := writeEvent(response, flusher, "log", newLogEvent("INFO", "GENOME", fmt.Sprintf("Genome generator online · seed %d", worldSeed))); err != nil {
			return
		}

		worldService := config.newWorld()
		var streamErr error
		creatureEvents := make([]services.CreatureInitializationEvent, 0)
		createdCreatures, err := worldService.CreateCreaturesWithObserver(func(event services.CreatureInitializationEvent) {
			if streamErr != nil {
				return
			}
			creatureEvents = append(creatureEvents, event)
			streamErr = writeEvent(response, flusher, "creature", event)
			if streamErr == nil {
				streamErr = writeEvent(response, flusher, "log", newLogEvent(
					"OK",
					"CREATURE",
					fmt.Sprintf("%s initialized · %s · %s", event.Name, event.Species, event.ID[:8]),
				))
			}
			if streamErr == nil && config.eventDelay > 0 {
				time.Sleep(config.eventDelay)
			}
		})
		if streamErr != nil {
			return
		}
		if err != nil {
			_ = writeEvent(response, flusher, "world-error", statusEvent{Message: err.Error()})
			return
		}
		simRuntime.setWorld(createdCreatures, creatureEvents)

		if err := writeEvent(response, flusher, "log", newLogEvent("READY", "WORLD", fmt.Sprintf("World online · %d creatures active", len(createdCreatures)))); err != nil {
			return
		}
		_ = writeEvent(response, flusher, "complete", completionEvent{Count: len(createdCreatures)})
	}
}

func newLogEvent(level string, scope string, message string) logEvent {
	return logEvent{
		Time:    time.Now().Format("15:04:05.000"),
		Level:   level,
		Scope:   scope,
		Message: message,
	}
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

func openBrowser(url string) error {
	var command *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		command = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
	case "darwin":
		command = exec.Command("open", url)
	default:
		command = exec.Command("xdg-open", url)
	}
	return command.Start()
}

func statusHandler(simRuntime *simulationRuntime) http.HandlerFunc {
	return func(response http.ResponseWriter, request *http.Request) {
		if request.Method != http.MethodGet {
			http.Error(response, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// Extract creature ID from URL path
		path := strings.TrimPrefix(request.URL.Path, "/api/world/status/")
		id := path
		if id == "" {
			http.Error(response, "creature ID required", http.StatusBadRequest)
			return
		}

		status, err := simRuntime.getCreatureStatus(id)
		if err != nil {
			http.Error(response, fmt.Sprintf("get creature status: %v", err), http.StatusNotFound)
			return
		}

		response.Header().Set("Content-Type", "application/json")
		json.NewEncoder(response).Encode(status)
	}
}
