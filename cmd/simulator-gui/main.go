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
	Count int `json:"count"`
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
	mux := http.NewServeMux()
	mux.HandleFunc("GET /{$}", serveAsset("index.html", "text/html; charset=utf-8"))
	mux.HandleFunc("GET /styles.css", serveAsset("styles.css", "text/css; charset=utf-8"))
	mux.HandleFunc("GET /app.js", serveAsset("app.js", "text/javascript; charset=utf-8"))
	mux.HandleFunc("GET /api/world/initialize", initializationHandler(config))
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

func initializationHandler(config guiConfig) http.HandlerFunc {
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
		createdCreatures, err := worldService.CreateCreaturesWithObserver(func(event services.CreatureInitializationEvent) {
			if streamErr != nil {
				return
			}
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
