package console

import (
	"fmt"
	"net/http"
	"sync"
)

// Event represents a server-sent event.
type Event struct {
	Type string
	Data string
}

// Broadcaster manages SSE client subscriptions and broadcasts events.
type Broadcaster struct {
	mu      sync.RWMutex
	clients map[<-chan Event]chan Event
}

// NewBroadcaster creates a new SSE broadcaster.
func NewBroadcaster() *Broadcaster {
	return &Broadcaster{
		clients: make(map[<-chan Event]chan Event),
	}
}

// Subscribe registers a new client and returns its event channel.
func (b *Broadcaster) Subscribe() <-chan Event {
	ch := make(chan Event, 16)
	b.mu.Lock()
	b.clients[ch] = ch
	b.mu.Unlock()
	return ch
}

// Unsubscribe removes a client and closes its channel.
func (b *Broadcaster) Unsubscribe(ch <-chan Event) {
	b.mu.Lock()
	if writeCh, ok := b.clients[ch]; ok {
		delete(b.clients, ch)
		close(writeCh)
	}
	b.mu.Unlock()
}

// Send broadcasts an event to all connected clients. Non-blocking: if a
// client's buffer is full the event is dropped for that client.
func (b *Broadcaster) Send(e Event) {
	b.mu.RLock()
	defer b.mu.RUnlock()
	for _, writeCh := range b.clients {
		select {
		case writeCh <- e:
		default:
		}
	}
}

// ClientCount returns the number of connected clients.
func (b *Broadcaster) ClientCount() int {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return len(b.clients)
}

// handleSSE streams server-sent events to the client.
func (s *Server) handleSSE(w http.ResponseWriter, r *http.Request) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming not supported", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.WriteHeader(http.StatusOK)
	flusher.Flush()

	ch := s.sse.Subscribe()
	defer s.sse.Unsubscribe(ch)

	for {
		select {
		case <-r.Context().Done():
			return
		case e, ok := <-ch:
			if !ok {
				return
			}
			fmt.Fprintf(w, "event: %s\ndata: %s\n\n", e.Type, e.Data)
			flusher.Flush()
		}
	}
}
