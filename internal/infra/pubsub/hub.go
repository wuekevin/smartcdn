package pubsub

import (
	"encoding/json"
	"sync"
)

// Event is a typed message pushed to subscribed clients.
type Event struct {
	Type    string      `json:"type"`    // e.g. "alert", "metric", "node_status"
	Payload interface{} `json:"payload"` // arbitrary JSON-serializable data
}

// Hub manages SSE client subscriptions and fan-out delivery.
type Hub struct {
	mu      sync.RWMutex
	clients map[string]map[chan []byte]struct{} // topic → set of subscriber channels
}

func NewHub() *Hub {
	return &Hub{clients: make(map[string]map[chan []byte]struct{})}
}

// Subscribe registers a channel to receive events for the given topic.
// Returns an unsubscribe func that must be called when the client disconnects.
func (h *Hub) Subscribe(topic string) (ch chan []byte, unsubscribe func()) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.clients[topic] == nil {
		h.clients[topic] = make(map[chan []byte]struct{})
	}
	c := make(chan []byte, 32)
	h.clients[topic][c] = struct{}{}
	return c, func() {
		h.mu.Lock()
		defer h.mu.Unlock()
		delete(h.clients[topic], c)
		close(c)
	}
}

// Publish sends an event to all subscribers of the given topic.
// Non-blocking: slow consumers are skipped (channel full).
func (h *Hub) Publish(topic string, evt Event) {
	data, err := json.Marshal(evt)
	if err != nil {
		return
	}
	h.mu.RLock()
	defer h.mu.RUnlock()
	for c := range h.clients[topic] {
		select {
		case c <- data:
		default:
		}
	}
}

// PublishAll broadcasts an event to every subscriber regardless of topic.
func (h *Hub) PublishAll(evt Event) {
	data, err := json.Marshal(evt)
	if err != nil {
		return
	}
	h.mu.RLock()
	defer h.mu.RUnlock()
	for _, subs := range h.clients {
		for c := range subs {
			select {
			case c <- data:
			default:
			}
		}
	}
}
