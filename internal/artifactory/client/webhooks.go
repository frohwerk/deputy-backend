package client

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"sync"
)

type EventHandlerFunc func(*ArtifactInfo) error

type Webhooks struct {
	m                sync.Mutex
	artifactDeployed []EventHandlerFunc
}

func (w *Webhooks) OnArtifactDeployed(f EventHandlerFunc) {
	w.m.Lock()
	defer w.m.Unlock()
	w.artifactDeployed = append(w.artifactDeployed, f)
}

func (w *Webhooks) WebhookHandler(rw http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeResponse(rw, http.StatusMethodNotAllowed, "Method %s not allowed", r.Method)
		return
	}

	event, err := decode(r.Body)

	if err != nil {
		log.Printf("Error decoding request message: %s", err)
		writeResponse(rw, http.StatusBadRequest, "Error decoding request message: %s", err)
		return
	}

	switch object := event.Data.(type) {
	case *ArtifactInfo:
		for _, eventHandler := range w.artifactDeployed {
			if err := eventHandler(object); err != nil {
				log.Printf("error in event handler: %v\n", err)
			}
		}
	default:
		writeResponse(rw, http.StatusBadRequest, "Unsupported domain: '%v'", event.Domain)
	}
}

func decode(r io.Reader) (*ArtifactEvent, error) {
	event := new(ArtifactEvent)
	object := new(json.RawMessage)
	event.Data = object
	if err := json.NewDecoder(r).Decode(event); err != nil {
		return nil, err
	}

	switch event.Domain {
	case "artifact":
		art := new(ArtifactInfo)
		if err := json.Unmarshal(*object, art); err != nil {
			return nil, err
		}
		event.Data = art
	default:
		return nil, fmt.Errorf("Unsupported domain: '%v'", event.Domain)
	}

	return event, nil
}

func writeResponse(rw http.ResponseWriter, status int, msg string, args ...interface{}) {
	rw.WriteHeader(status)
	if _, err := rw.Write([]byte(fmt.Sprintf(msg, args...))); err != nil {
		log.Printf("Error sending response message: %v", err)
	}
}
