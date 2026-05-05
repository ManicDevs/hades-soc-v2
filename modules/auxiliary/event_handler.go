package auxiliary

import (
	"context"
	"fmt"
	"time"

	"hades-v2/pkg/sdk"
)

// EventType represents types of events
type EventType string

const (
	EventSecurity    EventType = "security"
	EventSystem      EventType = "system"
	EventNetwork     EventType = "network"
	EventApplication EventType = "application"
)

// EventHandler provides event-driven automation functionality
type EventHandler struct {
	*sdk.BaseModule
	eventType EventType
}

// NewEventHandler creates a new event handler instance
func NewEventHandler() *EventHandler {
	return &EventHandler{
		BaseModule: sdk.NewBaseModule(
			"event_handler",
			"Event-driven trigger system for automation",
			sdk.CategoryReporting,
		),
		eventType: EventSecurity,
	}
}

// Execute starts event handling
func (eh *EventHandler) Execute(ctx context.Context) error {
	eh.SetStatus(sdk.StatusRunning)
	defer eh.SetStatus(sdk.StatusIdle)

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(180 * time.Millisecond):
		eh.SetStatus(sdk.StatusCompleted)
		return nil
	}
}

// SetEventType configures the event type
func (eh *EventHandler) SetEventType(eventType EventType) error {
	switch eventType {
	case EventSecurity, EventSystem, EventNetwork, EventApplication:
		eh.eventType = eventType
		return nil
	default:
		return fmt.Errorf("hades.auxiliary.event_handler: invalid event type: %s", eventType)
	}
}

// GetResult returns handler status
func (eh *EventHandler) GetResult() string {
	return fmt.Sprintf("Event handler initialized for: %s events", eh.eventType)
}
