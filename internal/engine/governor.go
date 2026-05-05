package engine

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"hades-v2/internal/bus"
	"hades-v2/internal/database"
)

// SafetyGovernor enforces the 5-block/hour limit and Manual ACK requirements using database persistence
type SafetyGovernor struct {
	mu                        sync.RWMutex
	maxAutomatedBlocksPerHour int
	requireManualAck          bool
	db                        *database.DatabaseManager
	ctx                       context.Context
	cancel                    context.CancelFunc
}

// NewSafetyGovernor creates a new Safety Governor instance with database persistence
func NewSafetyGovernor(db *database.DatabaseManager) *SafetyGovernor {
	ctx, cancel := context.WithCancel(context.Background())

	return &SafetyGovernor{
		maxAutomatedBlocksPerHour: 5,
		requireManualAck:          true,
		db:                        db,
		ctx:                       ctx,
		cancel:                    cancel,
	}
}

// ActionRequest represents a request for automated action
type ActionRequest struct {
	ActionName       string                 `json:"action_name"`
	Target           string                 `json:"target"`
	Reasoning        string                 `json:"reasoning"`
	RequiresApproval bool                   `json:"requires_approval"`
	Requester        string                 `json:"requester"`
	Timestamp        time.Time              `json:"timestamp"`
	Metadata         map[string]interface{} `json:"metadata,omitempty"`
}

// ActionResponse represents the response to an action request
type ActionResponse struct {
	Approved          bool      `json:"approved"`
	ActionID          string    `json:"action_id"`
	Reason            string    `json:"reason"`
	Timestamp         time.Time `json:"timestamp"`
	RequiresManualAck bool      `json:"requires_manual_ack"`
}

// Start begins the Safety Governor monitoring and initializes database tables
func (sg *SafetyGovernor) Start() {
	// Initialize database tables
	if err := sg.db.CreateGovernorActionTable(sg.ctx); err != nil {
		log.Printf("SafetyGovernor: Failed to create database tables: %v", err)
		return
	}

	go sg.monitorBlockLimits()
	log.Println("SafetyGovernor: Started monitoring with 5-block/hour limit (database persistence)")
}

// Stop gracefully shuts down the Safety Governor
func (sg *SafetyGovernor) Stop() {
	if sg.cancel != nil {
		sg.cancel()
	}
	log.Println("SafetyGovernor: Stopped monitoring")
}

// RequestAction checks if an action should be allowed based on Safety Governor rules using database persistence
func (sg *SafetyGovernor) RequestAction(action ActionRequest) (*ActionResponse, error) {
	sg.mu.Lock()
	defer sg.mu.Unlock()

	startTime := time.Now()
	actionID := generateActionID(action)

	// Get current block count from database (rolling 1-hour window)
	currentBlockCount, err := sg.db.GetBlockCountInLastHour(sg.ctx)
	if err != nil {
		log.Printf("SafetyGovernor: Failed to get block count from database: %v", err)
		return nil, fmt.Errorf("database error: %w", err)
	}

	// Check if action requires manual approval
	if sg.requireManualAck && action.RequiresApproval {
		log.Printf("SafetyGovernor: Action '%s' requires manual approval - blocking", action.ActionName)

		// Record the manual ack requirement in database
		govAction := &database.GovernorAction{
			ActionID:          actionID,
			ActionName:        action.ActionName,
			Target:            action.Target,
			Reasoning:         action.Reasoning,
			Requester:         action.Requester,
			Status:            string(database.GovernorActionStatusManualAckRequired),
			RequiresApproval:  action.RequiresApproval,
			Approved:          false,
			RequiresManualAck: true,
			BlockReason:       "Manual ACK required for destructive action",
			ExecutionTime:     time.Since(startTime).Milliseconds(),
			CreatedAt:         time.Now(),
			UpdatedAt:         time.Now(),
		}

		if err := sg.db.RecordGovernorAction(sg.ctx, govAction); err != nil {
			log.Printf("SafetyGovernor: Failed to record manual ack action: %v", err)
		}

		// Publish Manual ACK request to WebSocket for UI approval
		bus.Default().Publish(bus.Event{
			Type:   bus.EventTypeActionRequest,
			Source: "safety_governor",
			Target: "dashboard",
			Payload: map[string]interface{}{
				"action_name":       action.ActionName,
				"target":            action.Target,
				"reasoning":         action.Reasoning,
				"requires_approval": true,
				"requester":         action.Requester,
				"timestamp":         time.Now().Unix(),
				"governor_block":    true,
				"block_reason":      "Manual ACK required for destructive action",
			},
		})

		return &ActionResponse{
			Approved:          false,
			ActionID:          actionID,
			Reason:            "Manual ACK required for destructive action",
			Timestamp:         time.Now(),
			RequiresManualAck: true,
		}, nil
	}

	// Check 5-block/hour limit using database count
	if currentBlockCount >= sg.maxAutomatedBlocksPerHour {
		log.Printf("SafetyGovernor: Hourly block limit reached (%d/%d) - blocking action '%s'",
			currentBlockCount, sg.maxAutomatedBlocksPerHour, action.ActionName)

		// Record the block action in database
		govAction := &database.GovernorAction{
			ActionID:          actionID,
			ActionName:        action.ActionName,
			Target:            action.Target,
			Reasoning:         action.Reasoning,
			Requester:         action.Requester,
			Status:            string(database.GovernorActionStatusBlocked),
			RequiresApproval:  action.RequiresApproval,
			Approved:          false,
			RequiresManualAck: false,
			BlockReason:       fmt.Sprintf("Hourly automated block limit reached (%d/%d)", currentBlockCount, sg.maxAutomatedBlocksPerHour),
			ExecutionTime:     time.Since(startTime).Milliseconds(),
			CreatedAt:         time.Now(),
			UpdatedAt:         time.Now(),
		}

		if err := sg.db.RecordGovernorAction(sg.ctx, govAction); err != nil {
			log.Printf("SafetyGovernor: Failed to record blocked action: %v", err)
		}

		// Publish block notification
		bus.Default().Publish(bus.Event{
			Type:   bus.EventTypeLogEvent,
			Source: "safety_governor",
			Target: action.Target,
			Payload: map[string]interface{}{
				"agent_name": "safety_governor",
				"message":    fmt.Sprintf("Automated action '%s' blocked - hourly limit reached", action.ActionName),
				"internal_reasoning": fmt.Sprintf("Safety Governor blocked action '%s' on target '%s'. Current blocks: %d/%d",
					action.ActionName, action.Target, currentBlockCount, sg.maxAutomatedBlocksPerHour),
				"timestamp": time.Now().Unix(),
				"status":    "blocked",
				"severity":  "high",
			},
		})

		return &ActionResponse{
			Approved:          false,
			ActionID:          actionID,
			Reason:            fmt.Sprintf("Hourly automated block limit reached (%d/%d)", currentBlockCount, sg.maxAutomatedBlocksPerHour),
			Timestamp:         time.Now(),
			RequiresManualAck: false,
		}, fmt.Errorf("hourly block limit exceeded")
	}

	// Check for duplicate actions within cooldown period using database
	if sg.isDuplicateAction(action) {
		log.Printf("SafetyGovernor: Duplicate action '%s' on target '%s' blocked", action.ActionName, action.Target)

		// Record the duplicate action in database
		govAction := &database.GovernorAction{
			ActionID:          actionID,
			ActionName:        action.ActionName,
			Target:            action.Target,
			Reasoning:         action.Reasoning,
			Requester:         action.Requester,
			Status:            string(database.GovernorActionStatusBlocked),
			RequiresApproval:  action.RequiresApproval,
			Approved:          false,
			RequiresManualAck: false,
			BlockReason:       "Duplicate action within cooldown period",
			ExecutionTime:     time.Since(startTime).Milliseconds(),
			CreatedAt:         time.Now(),
			UpdatedAt:         time.Now(),
		}

		if err := sg.db.RecordGovernorAction(sg.ctx, govAction); err != nil {
			log.Printf("SafetyGovernor: Failed to record duplicate action: %v", err)
		}

		return &ActionResponse{
			Approved:          false,
			ActionID:          actionID,
			Reason:            "Duplicate action within cooldown period",
			Timestamp:         time.Now(),
			RequiresManualAck: false,
		}, fmt.Errorf("duplicate action blocked")
	}

	// Action is approved - record it in database
	govAction := &database.GovernorAction{
		ActionID:          actionID,
		ActionName:        action.ActionName,
		Target:            action.Target,
		Reasoning:         action.Reasoning,
		Requester:         action.Requester,
		Status:            string(database.GovernorActionStatusApproved),
		RequiresApproval:  action.RequiresApproval,
		Approved:          true,
		RequiresManualAck: false,
		ExecutionTime:     time.Since(startTime).Milliseconds(),
		CreatedAt:         time.Now(),
		UpdatedAt:         time.Now(),
	}

	if err := sg.db.RecordGovernorAction(sg.ctx, govAction); err != nil {
		log.Printf("SafetyGovernor: Failed to record approved action: %v", err)
		// Continue even if recording fails - the action is approved
	}

	log.Printf("SafetyGovernor: Action '%s' approved (%d/%d blocks used this hour)",
		action.ActionName, currentBlockCount+1, sg.maxAutomatedBlocksPerHour)

	// Publish approval event
	bus.Default().Publish(bus.Event{
		Type:   bus.EventTypeLogEvent,
		Source: "safety_governor",
		Target: action.Target,
		Payload: map[string]interface{}{
			"agent_name": "safety_governor",
			"message":    fmt.Sprintf("Automated action '%s' approved", action.ActionName),
			"internal_reasoning": fmt.Sprintf("Safety Governor approved action '%s' on target '%s'. Remaining blocks this hour: %d/%d",
				action.ActionName, action.Target, sg.maxAutomatedBlocksPerHour-(currentBlockCount+1), sg.maxAutomatedBlocksPerHour),
			"timestamp": time.Now().Unix(),
			"status":    "approved",
		},
	})

	return &ActionResponse{
		Approved:          true,
		ActionID:          actionID,
		Reason:            "Action approved by Safety Governor",
		Timestamp:         time.Now(),
		RequiresManualAck: false,
	}, nil
}

// GetStatus returns current Safety Governor status from database
func (sg *SafetyGovernor) GetStatus() map[string]interface{} {
	sg.mu.RLock()
	defer sg.mu.RUnlock()

	// Get statistics from database
	stats, err := sg.db.GetGovernorStats(sg.ctx)
	if err != nil {
		log.Printf("SafetyGovernor: Failed to get stats from database: %v", err)
		// Return fallback status
		return map[string]interface{}{
			"max_blocks_per_hour": sg.maxAutomatedBlocksPerHour,
			"current_block_count": 0,
			"remaining_blocks":    sg.maxAutomatedBlocksPerHour,
			"manual_ack_required": sg.requireManualAck,
			"database_error":      true,
			"error_message":       err.Error(),
		}
	}

	// Add governor configuration to stats
	stats["max_blocks_per_hour"] = sg.maxAutomatedBlocksPerHour
	stats["manual_ack_required"] = sg.requireManualAck

	// Add current_block_count for test compatibility
	if approvedLastHour, ok := stats["approved_last_hour"].(int); ok {
		stats["current_block_count"] = approvedLastHour
		stats["remaining_blocks"] = sg.maxAutomatedBlocksPerHour - approvedLastHour
	} else {
		stats["current_block_count"] = 0
		stats["remaining_blocks"] = sg.maxAutomatedBlocksPerHour
	}

	return stats
}

// SetManualAckRequirement enables/disables manual ACK requirement
func (sg *SafetyGovernor) SetManualAckRequirement(require bool) {
	sg.mu.Lock()
	defer sg.mu.Unlock()

	sg.requireManualAck = require
	log.Printf("SafetyGovernor: Manual ACK requirement set to %t", require)
}

// monitorBlockLimits runs periodic monitoring of block limits
func (sg *SafetyGovernor) monitorBlockLimits() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-sg.ctx.Done():
			return
		case <-ticker.C:
			sg.checkAndPublishStatus()
		}
	}
}

// checkAndPublishStatus publishes current governor status from database
func (sg *SafetyGovernor) checkAndPublishStatus() {
	sg.mu.RLock()
	status := sg.GetStatus()
	sg.mu.RUnlock()

	// Publish status event for dashboard
	bus.Default().Publish(bus.Event{
		Type:   bus.EventTypeLogEvent,
		Source: "safety_governor",
		Target: "dashboard",
		Payload: map[string]interface{}{
			"agent_name": "safety_governor",
			"message":    "Safety Governor status update",
			"internal_reasoning": fmt.Sprintf("Current block usage: %d/%d, Time until reset: %v",
				status["current_block_count"], status["max_blocks_per_hour"], status["time_until_reset"]),
			"timestamp":       time.Now().Unix(),
			"status":          "monitoring",
			"governor_status": status,
		},
	})
}

// isDuplicateAction checks if the same action was performed recently using database
func (sg *SafetyGovernor) isDuplicateAction(action ActionRequest) bool {
	// Get recent actions from the last 5 minutes
	fiveMinutesAgo := time.Now().Add(-5 * time.Minute)
	recentActions, err := sg.db.GetRecentActions(sg.ctx, fiveMinutesAgo, 100)
	if err != nil {
		log.Printf("SafetyGovernor: Failed to get recent actions for duplicate check: %v", err)
		return false // Allow action if we can't check for duplicates
	}

	// Check for duplicate actions
	for _, recentAction := range recentActions {
		if recentAction.ActionName == action.ActionName &&
			recentAction.Target == action.Target &&
			recentAction.Status == string(database.GovernorActionStatusApproved) {
			return true
		}
	}

	return false
}

// generateActionID generates a unique action ID
func generateActionID(action ActionRequest) string {
	return fmt.Sprintf("action_%s_%s_%d", action.ActionName, action.Target, time.Now().UnixNano())
}
