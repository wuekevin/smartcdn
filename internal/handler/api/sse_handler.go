package api

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/wuekevin/smartcdn/internal/infra/pubsub"
)

// SSEHandler streams real-time events to browsers via Server-Sent Events.
type SSEHandler struct {
	hub *pubsub.Hub
}

func NewSSEHandler(hub *pubsub.Hub) *SSEHandler {
	return &SSEHandler{hub: hub}
}

func (h *SSEHandler) RegisterRoutes(r *gin.RouterGroup) {
	r.GET("/events", h.Stream)
	r.GET("/events/:topic", h.StreamTopic)
}

// Stream subscribes the client to ALL events.
func (h *SSEHandler) Stream(c *gin.Context) {
	h.streamTopic(c, "")
}

// StreamTopic subscribes the client to a specific topic (e.g. "alerts", "metrics", "nodes").
func (h *SSEHandler) StreamTopic(c *gin.Context) {
	topic := c.Param("topic")
	h.streamTopic(c, topic)
}

func (h *SSEHandler) streamTopic(c *gin.Context, topic string) {
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("X-Accel-Buffering", "no") // disable nginx buffering

	var (
		ch          chan []byte
		unsubscribe func()
	)
	if topic == "" {
		// Subscribe to all topics via a wildcard channel; use a dedicated "all" topic key.
		ch, unsubscribe = h.hub.Subscribe("__all__")
	} else {
		ch, unsubscribe = h.hub.Subscribe(topic)
	}
	defer unsubscribe()

	// Send an initial keep-alive comment so the browser doesn't time out before first event.
	fmt.Fprintf(c.Writer, ": connected\n\n")
	c.Writer.Flush()

	ticker := time.NewTicker(25 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case data, ok := <-ch:
			if !ok {
				return
			}
			fmt.Fprintf(c.Writer, "data: %s\n\n", data)
			c.Writer.Flush()

		case <-ticker.C:
			// Keep-alive ping.
			fmt.Fprintf(c.Writer, ": ping\n\n")
			c.Writer.Flush()

		case <-c.Request.Context().Done():
			return
		}
	}
}

// PublishMiddleware returns a gin middleware that exposes the Hub on the context.
// Downstream handlers can retrieve it with: hub := c.MustGet("hub").(*pubsub.Hub)
func PublishMiddleware(hub *pubsub.Hub) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set("hub", hub)
		c.Next()
	}
}

// --- Webhook: node/probe status push helper (called from other handlers) ---

// PushAlertEvent pushes an alert-fired event to the "alerts" SSE topic.
func PushAlertEvent(hub *pubsub.Hub, payload interface{}) {
	hub.Publish("alerts", pubsub.Event{Type: "alert_fired", Payload: payload})
	hub.Publish("__all__", pubsub.Event{Type: "alert_fired", Payload: payload})
}

// PushMetricEvent pushes a metric snapshot to the "metrics" SSE topic.
func PushMetricEvent(hub *pubsub.Hub, payload interface{}) {
	hub.Publish("metrics", pubsub.Event{Type: "metric_updated", Payload: payload})
	hub.Publish("__all__", pubsub.Event{Type: "metric_updated", Payload: payload})
}

// PushNodeEvent pushes a node status change to the "nodes" SSE topic.
func PushNodeEvent(hub *pubsub.Hub, payload interface{}) {
	hub.Publish("nodes", pubsub.Event{Type: "node_status", Payload: payload})
	hub.Publish("__all__", pubsub.Event{Type: "node_status", Payload: payload})
}

// SSEResponse is the JSON envelope sent as the HTTP response for non-SSE clients
// that hit an SSE endpoint without Accept: text/event-stream.
func sseNotSupported(c *gin.Context) {
	c.JSON(http.StatusNotAcceptable, gin.H{"error": "use Accept: text/event-stream"})
}
