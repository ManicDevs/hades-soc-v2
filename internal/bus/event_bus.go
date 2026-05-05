package bus

import (
	"fmt"
	"sync"
	"time"
)

type EventType string

const (
	EventTypeReconComplete        EventType = "recon.complete"
	EventTypeExploitationComplete EventType = "exploitation.complete"
	EventTypeScanComplete         EventType = "scan.complete"
	EventTypeThreatDetected       EventType = "threat.detected"
	EventTypeIncidentCreated      EventType = "incident.created"

	// Agentic SOC Event Types
	EventTypeDomainFound           EventType = "agent.domain.found"
	EventTypePortDiscovered        EventType = "agent.port.discovered"
	EventTypeVulnerabilityDetected EventType = "agent.vulnerability.detected"
	EventTypeCredentialFound       EventType = "agent.credential.found"
	EventTypeExploitAvailable      EventType = "agent.exploit.available"
	EventTypeCriticalThreat        EventType = "agent.threat.critical"
	EventTypeAgentDecision         EventType = "agent.decision"
	EventTypeNodeIsolated          EventType = "agent.node.isolated"
	EventTypeModuleLaunched        EventType = "agent.module.launched"
	EventTypeModuleCompleted       EventType = "agent.module.completed"

	// Logging & Reasoning Event Types
	EventTypeLogEvent EventType = "log.event"

	// Authentication & Security Event Types
	EventTypeAuthFailure            EventType = "auth.failure"
	EventTypeSecurityUpgradeRequest EventType = "security.upgrade.request"

	// Autonomous Cascade Event Types
	EventTypeNewAsset            EventType = "asset.new"
	EventTypeActionRequest       EventType = "action.request"
	EventTypeThreat              EventType = "threat.detected"
	EventTypeVulnerabilityFound  EventType = "vulnerability.found"
	EventTypeModuleHotSwap       EventType = "module.hot_swap"
	EventTypeVerificationScan    EventType = "verification.scan"
	EventTypeLateralMovement     EventType = "lateral.movement"
	EventTypeHoneyTokenTriggered EventType = "honey_token.triggered"
	EventTypeHoneyFileAccessed   EventType = "honey_file.accessed"
)

type Event struct {
	ID        string                 `json:"id"`
	Type      EventType              `json:"type"`
	Source    string                 `json:"source"`
	Target    string                 `json:"target"`
	Payload   map[string]interface{} `json:"payload"`
	Timestamp time.Time              `json:"timestamp"`
}

type EventHandler func(Event) error

type Subscription struct {
	ID        string
	EventType EventType
	Handler   EventHandler
	Filter    func(Event) bool
}

type EventBus struct {
	subscribers map[EventType][]*Subscription
	mu          sync.RWMutex
	eventChan   chan Event
	stopChan    chan struct{}
	wg          sync.WaitGroup
}

var (
	defaultBus *EventBus
	once       sync.Once
)

func New() *EventBus {
	eb := &EventBus{
		subscribers: make(map[EventType][]*Subscription),
		eventChan:   make(chan Event, 1000),
		stopChan:    make(chan struct{}),
	}

	eb.wg.Add(1)
	go eb.dispatchLoop()

	return eb
}

func Default() *EventBus {
	once.Do(func() {
		defaultBus = New()
	})
	return defaultBus
}

func (eb *EventBus) Subscribe(eventType EventType, handler EventHandler) *Subscription {
	return eb.SubscribeWithFilter(eventType, handler, nil)
}

func (eb *EventBus) SubscribeWithFilter(eventType EventType, handler EventHandler, filter func(Event) bool) *Subscription {
	eb.mu.Lock()
	defer eb.mu.Unlock()

	sub := &Subscription{
		ID:        fmt.Sprintf("sub_%d", time.Now().UnixNano()),
		EventType: eventType,
		Handler:   handler,
		Filter:    filter,
	}

	eb.subscribers[eventType] = append(eb.subscribers[eventType], sub)
	return sub
}

func (eb *EventBus) Unsubscribe(sub *Subscription) {
	eb.mu.Lock()
	defer eb.mu.Unlock()

	subs := eb.subscribers[sub.EventType]
	for i, s := range subs {
		if s.ID == sub.ID {
			eb.subscribers[sub.EventType] = append(subs[:i], subs[i+1:]...)
			return
		}
	}
}

func (eb *EventBus) Publish(event Event) {
	if event.ID == "" {
		event.ID = fmt.Sprintf("evt_%d", time.Now().UnixNano())
	}
	if event.Timestamp.IsZero() {
		event.Timestamp = time.Now()
	}

	select {
	case eb.eventChan <- event:
	default:
		fmt.Printf("event bus: channel full, dropping event %s\n", event.ID)
	}
}

func (eb *EventBus) PublishAsync(event Event) {
	go eb.Publish(event)
}

func (eb *EventBus) dispatchLoop() {
	defer eb.wg.Done()

	for {
		select {
		case <-eb.stopChan:
			return
		case event := <-eb.eventChan:
			eb.dispatch(event)
		}
	}
}

func (eb *EventBus) dispatch(event Event) {
	eb.mu.RLock()
	defer eb.mu.RUnlock()

	subs, ok := eb.subscribers[event.Type]
	if !ok {
		return
	}

	for _, sub := range subs {
		if sub.Filter != nil && !sub.Filter(event) {
			continue
		}

		if err := sub.Handler(event); err != nil {
			fmt.Printf("event bus: handler error for event %s: %v\n", event.ID, err)
		}
	}
}

func (eb *EventBus) Stop() {
	close(eb.stopChan)
	eb.wg.Wait()
	close(eb.eventChan)
}

func (eb *EventBus) SubscriberCount(eventType EventType) int {
	eb.mu.RLock()
	defer eb.mu.RUnlock()
	return len(eb.subscribers[eventType])
}

func (eb *EventBus) TotalSubscribers() int {
	eb.mu.RLock()
	defer eb.mu.RUnlock()

	count := 0
	for _, subs := range eb.subscribers {
		count += len(subs)
	}
	return count
}
