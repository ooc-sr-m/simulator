package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"simulator/services"
)

func TestGraphicalEntrypointServesInterface(t *testing.T) {
	handler := newTestHandler()
	request := httptest.NewRequest(http.MethodGet, "/", nil)
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("GET / status = %d, want %d", response.Code, http.StatusOK)
	}
	if body := response.Body.String(); !strings.Contains(body, "INITIALIZING WORLD") {
		t.Fatal("GET / does not contain the loading interface")
	}
}

func TestGraphicalEntrypointStreamsCreatureEvents(t *testing.T) {
	handler := newTestHandler()
	request := httptest.NewRequest(http.MethodGet, "/api/world/initialize", nil)
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("initialize status = %d, want %d", response.Code, http.StatusOK)
	}
	body := response.Body.String()
	if count := strings.Count(body, "event: creature\n"); count != 10 {
		t.Fatalf("creature event count = %d, want 10", count)
	}
	if !strings.Contains(body, "event: complete\n") {
		t.Fatal("initialize stream does not contain a completion event")
	}
	if count := strings.Count(body, "event: log\n"); count < 3 {
		t.Fatalf("log event count = %d, want at least 3", count)
	}

	firstCreatureData := eventData(t, body, "creature")
	var event services.CreatureInitializationEvent
	if err := json.Unmarshal([]byte(firstCreatureData), &event); err != nil {
		t.Fatalf("decode creature event: %v", err)
	}
	if event.Sequence != 1 || event.Total != 10 || event.Species != "dog" {
		t.Fatalf("first creature event = %+v, want first of ten dogs", event)
	}
	if event.Name == "" || event.Name == "Dog 1" {
		t.Fatalf("first creature name = %q, want a name from the embedded catalog", event.Name)
	}
	if event.HungerMax == 0 || event.HungerRate == 0 || len(event.Capabilities) == 0 {
		t.Fatalf("first creature event has incomplete details: %+v", event)
	}
}

func newTestHandler() http.Handler {
	return newHandler(guiConfig{
		newWorld: func() *services.WorldService {
			return services.NewWorldService(worldSeed)
		},
	})
}

func eventData(t *testing.T, body string, eventName string) string {
	t.Helper()
	prefix := "event: " + eventName + "\ndata: "
	start := strings.Index(body, prefix)
	if start < 0 {
		t.Fatalf("event %q not found", eventName)
	}
	start += len(prefix)
	end := strings.Index(body[start:], "\n\n")
	if end < 0 {
		t.Fatalf("event %q has no terminator", eventName)
	}
	return body[start : start+end]
}
