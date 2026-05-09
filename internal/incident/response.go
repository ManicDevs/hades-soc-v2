package incident

import (
	"context"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"hades-v2/internal/threat"
)

// IncidentResponseManager manages automated incident response workflows
type IncidentResponseManager struct {
	workflows           map[string]*Workflow
	actions             map[string]*ResponseAction
	engines             map[string]*ResponseEngine
	activeIncidents     map[string]*Incident
	mu                  sync.RWMutex
	threatDetector      *threat.ThreatDetector
	notificationService *NotificationService
	escalationManager   *EscalationManager
}

// Workflow represents an automated response workflow
type Workflow struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Triggers    []Trigger              `json:"triggers"`
	Steps       []WorkflowStep         `json:"steps"`
	Enabled     bool                   `json:"enabled"`
	Priority    int                    `json:"priority"`
	Created     time.Time              `json:"created"`
	Updated     time.Time              `json:"updated"`
	Metadata    map[string]interface{} `json:"metadata"`
}

// Trigger represents a workflow trigger condition
type Trigger struct {
	Type       string                 `json:"type"`       // "threat_type", "severity", "confidence", "custom"
	Condition  string                 `json:"condition"`  // "equals", "contains", "greater_than", "less_than"
	Value      interface{}            `json:"value"`      // The value to compare against
	Field      string                 `json:"field"`      // The field to check
	Parameters map[string]interface{} `json:"parameters"` // Additional parameters
}

// WorkflowStep represents a step in a workflow
type WorkflowStep struct {
	ID         string                 `json:"id"`
	Name       string                 `json:"name"`
	Type       string                 `json:"type"`        // "action", "condition", "delay", "parallel", "notification"
	Action     string                 `json:"action"`      // The action to execute
	Parameters map[string]interface{} `json:"parameters"`  // Action parameters
	Conditions []Condition            `json:"conditions"`  // Conditions to execute this step
	Timeout    time.Duration          `json:"timeout"`     // Step timeout
	RetryCount int                    `json:"retry_count"` // Number of retries
	NextSteps  []string               `json:"next_steps"`  // Next step IDs
	OnFailure  string                 `json:"on_failure"`  // Step ID to execute on failure
	Parallel   bool                   `json:"parallel"`    // Execute in parallel with other steps
}

// Condition represents a condition for workflow execution
type Condition struct {
	Field     string      `json:"field"`
	Operator  string      `json:"operator"` // "equals", "not_equals", "contains", "greater_than", "less_than"
	Value     interface{} `json:"value"`
	LogicalOp string      `json:"logical_op"` // "and", "or"
}

// ResponseAction represents a response action
type ResponseAction struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Type        string                 `json:"type"` // "block_ip", "isolate_system", "disable_account", "notify", "custom"
	Description string                 `json:"description"`
	Parameters  map[string]interface{} `json:"parameters"`
	Enabled     bool                   `json:"enabled"`
	Required    bool                   `json:"required"` // Whether this action is required
	Timeout     time.Duration          `json:"timeout"`
	RetryCount  int                    `json:"retry_count"`
}

// ResponseEngine executes response actions
type ResponseEngine struct {
	Name      string
	Actions   map[string]*ResponseAction
	Executors map[string]ActionExecutor
}

// ActionExecutor interface for different action types
type ActionExecutor interface {
	Name() string
	Execute(ctx context.Context, action *ResponseAction, incident *Incident) (*ActionResult, error)
	CanExecute(action *ResponseAction, incident *Incident) bool
	Validate(action *ResponseAction) error
}

// ActionResult represents the result of an action execution
type ActionResult struct {
	Success    bool                   `json:"success"`
	Message    string                 `json:"message"`
	Data       map[string]interface{} `json:"data"`
	ExecutedAt time.Time              `json:"executed_at"`
	Duration   time.Duration          `json:"duration"`
	Error      string                 `json:"error,omitempty"`
	RetryCount int                    `json:"retry_count"`
	Metadata   map[string]interface{} `json:"metadata"`
}

// Incident represents a security incident
type Incident struct {
	ID              string                 `json:"id"`
	Title           string                 `json:"title"`
	Description     string                 `json:"description"`
	Severity        string                 `json:"severity"`
	Status          string                 `json:"status"` // "new", "investigating", "resolved", "closed"
	Priority        int                    `json:"priority"`
	Source          string                 `json:"source"` // "automated", "manual"
	ThreatAlert     *threat.ThreatAlert    `json:"threat_alert"`
	Created         time.Time              `json:"created"`
	Updated         time.Time              `json:"updated"`
	AssignedTo      string                 `json:"assigned_to"`
	Tags            []string               `json:"tags"`
	Actions         []IncidentAction       `json:"actions"`
	WorkflowSteps   []WorkflowStepResult   `json:"workflow_steps"`
	Evidence        []Evidence             `json:"evidence"`
	Notes           []IncidentNote         `json:"notes"`
	Metadata        map[string]interface{} `json:"metadata"`
	EscalationLevel int                    `json:"escalation_level"`
}

// IncidentAction represents an action taken on an incident
type IncidentAction struct {
	ID          string                 `json:"id"`
	Type        string                 `json:"type"`
	Description string                 `json:"description"`
	Status      string                 `json:"status"` // "pending", "executing", "completed", "failed"
	ExecutedAt  time.Time              `json:"executed_at"`
	ExecutedBy  string                 `json:"executed_by"`
	Result      *ActionResult          `json:"result"`
	Metadata    map[string]interface{} `json:"metadata"`
}

// WorkflowStepResult represents the result of a workflow step
type WorkflowStepResult struct {
	StepID     string                 `json:"step_id"`
	Status     string                 `json:"status"` // "pending", "executing", "completed", "failed", "skipped"
	ExecutedAt time.Time              `json:"executed_at"`
	Duration   time.Duration          `json:"duration"`
	Result     *ActionResult          `json:"result"`
	Error      string                 `json:"error"`
	RetryCount int                    `json:"retry_count"`
	Metadata   map[string]interface{} `json:"metadata"`
}

// Evidence represents evidence collected for an incident
type Evidence struct {
	ID          string                 `json:"id"`
	Type        string                 `json:"type"` // "log", "network", "file", "memory", "screenshot"
	Description string                 `json:"description"`
	Data        interface{}            `json:"data"`
	CollectedAt time.Time              `json:"collected_at"`
	CollectedBy string                 `json:"collected_by"`
	Hash        string                 `json:"hash"`
	Metadata    map[string]interface{} `json:"metadata"`
}

// IncidentNote represents a note on an incident
type IncidentNote struct {
	ID        string    `json:"id"`
	Content   string    `json:"content"`
	Author    string    `json:"author"`
	CreatedAt time.Time `json:"created_at"`
	Private   bool      `json:"private"`
}

// NotificationService handles incident notifications
type NotificationService struct {
	channels map[string]NotificationChannel
}

// NotificationChannel interface for different notification channels
type NotificationChannel interface {
	Name() string
	Send(ctx context.Context, notification *Notification) error
	Validate(notification *Notification) error
}

// Notification represents a notification
type Notification struct {
	Type      string                 `json:"type"` // "email", "sms", "slack", "webhook"
	Recipient string                 `json:"recipient"`
	Subject   string                 `json:"subject"`
	Message   string                 `json:"message"`
	Data      map[string]interface{} `json:"data"`
	Priority  string                 `json:"priority"` // "low", "medium", "high", "critical"
	Timestamp time.Time              `json:"timestamp"`
}

// EscalationManager handles incident escalation
type EscalationManager struct {
	policies map[string]*EscalationPolicy
	mu       sync.RWMutex
}

// EscalationPolicy represents an escalation policy
type EscalationPolicy struct {
	ID             string                 `json:"id"`
	Name           string                 `json:"name"`
	Conditions     []EscalationCondition  `json:"conditions"`
	Actions        []EscalationAction     `json:"actions"`
	Enabled        bool                   `json:"enabled"`
	LastEscalation time.Time              `json:"last_escalation"`
	Metadata       map[string]interface{} `json:"metadata"`
}

// EscalationCondition represents an escalation condition
type EscalationCondition struct {
	Field    string        `json:"field"`
	Operator string        `json:"operator"`
	Value    interface{}   `json:"value"`
	Duration time.Duration `json:"duration"` // Time condition must persist
}

// EscalationAction represents an escalation action
type EscalationAction struct {
	Type       string                 `json:"type"`
	Parameters map[string]interface{} `json:"parameters"`
	Delay      time.Duration          `json:"delay"`
}

// NewIncidentResponseManager creates a new incident response manager
func NewIncidentResponseManager(threatDetector *threat.ThreatDetector) *IncidentResponseManager {
	irm := &IncidentResponseManager{
		workflows:           make(map[string]*Workflow),
		actions:             make(map[string]*ResponseAction),
		engines:             make(map[string]*ResponseEngine),
		activeIncidents:     make(map[string]*Incident),
		threatDetector:      threatDetector,
		notificationService: NewNotificationService(),
		escalationManager:   NewEscalationManager(),
	}

	// Initialize default workflows and actions
	irm.initializeDefaultWorkflows()
	irm.initializeDefaultActions()
	irm.initializeDefaultEngines()

	return irm
}

// initializeDefaultWorkflows initializes default response workflows
func (irm *IncidentResponseManager) initializeDefaultWorkflows() {
	defaultWorkflows := []*Workflow{
		{
			ID:          "malware_response",
			Name:        "Malware Detection Response",
			Description: "Automated response for malware detection",
			Triggers: []Trigger{
				{
					Type:      "threat_type",
					Condition: "equals",
					Value:     "malware",
					Field:     "threat_type",
				},
			},
			Steps: []WorkflowStep{
				{
					ID:         "isolate_system",
					Name:       "Isolate Affected System",
					Type:       "action",
					Action:     "isolate_system",
					Parameters: map[string]interface{}{"isolation_level": "full"},
					Timeout:    30 * time.Second,
					RetryCount: 3,
					NextSteps:  []string{"scan_system"},
				},
				{
					ID:         "scan_system",
					Name:       "Scan System",
					Type:       "action",
					Action:     "security_scan",
					Parameters: map[string]interface{}{"scan_type": "full"},
					Timeout:    5 * time.Minute,
					RetryCount: 2,
					NextSteps:  []string{"notify_team"},
				},
				{
					ID:     "notify_team",
					Name:   "Notify Security Team",
					Type:   "notification",
					Action: "send_notification",
					Parameters: map[string]interface{}{
						"channels": []string{"email", "slack"},
						"priority": "high",
					},
					Timeout:    10 * time.Second,
					RetryCount: 3,
				},
			},
			Enabled:  true,
			Priority: 1,
		},
		{
			ID:          "ddos_response",
			Name:        "DDoS Attack Response",
			Description: "Automated response for DDoS attacks",
			Triggers: []Trigger{
				{
					Type:      "threat_type",
					Condition: "equals",
					Value:     "ddos",
					Field:     "threat_type",
				},
			},
			Steps: []WorkflowStep{
				{
					ID:         "activate_ddos_protection",
					Name:       "Activate DDoS Protection",
					Type:       "action",
					Action:     "activate_ddos_protection",
					Parameters: map[string]interface{}{"protection_level": "high"},
					Timeout:    30 * time.Second,
					RetryCount: 3,
					NextSteps:  []string{"block_ips"},
				},
				{
					ID:         "block_ips",
					Name:       "Block Malicious IPs",
					Type:       "action",
					Action:     "block_ip",
					Parameters: map[string]interface{}{"duration": "24h"},
					Timeout:    10 * time.Second,
					RetryCount: 2,
					NextSteps:  []string{"notify_team"},
				},
				{
					ID:     "notify_team",
					Name:   "Notify Security Team",
					Type:   "notification",
					Action: "send_notification",
					Parameters: map[string]interface{}{
						"channels": []string{"email", "slack", "sms"},
						"priority": "critical",
					},
					Timeout:    10 * time.Second,
					RetryCount: 3,
				},
			},
			Enabled:  true,
			Priority: 1,
		},
		{
			ID:          "sql_injection_response",
			Name:        "SQL Injection Response",
			Description: "Automated response for SQL injection attacks",
			Triggers: []Trigger{
				{
					Type:      "threat_type",
					Condition: "equals",
					Value:     "sql_injection",
					Field:     "threat_type",
				},
			},
			Steps: []WorkflowStep{
				{
					ID:         "block_ip",
					Name:       "Block Attacker IP",
					Type:       "action",
					Action:     "block_ip",
					Parameters: map[string]interface{}{"duration": "72h"},
					Timeout:    10 * time.Second,
					RetryCount: 3,
					NextSteps:  []string{"rotate_credentials"},
				},
				{
					ID:         "rotate_credentials",
					Name:       "Rotate Database Credentials",
					Type:       "action",
					Action:     "rotate_credentials",
					Parameters: map[string]interface{}{"scope": "affected_database"},
					Timeout:    2 * time.Minute,
					RetryCount: 2,
					NextSteps:  []string{"notify_team"},
				},
				{
					ID:     "notify_team",
					Name:   "Notify Security Team",
					Type:   "notification",
					Action: "send_notification",
					Parameters: map[string]interface{}{
						"channels": []string{"email", "slack"},
						"priority": "high",
					},
					Timeout:    10 * time.Second,
					RetryCount: 3,
				},
			},
			Enabled:  true,
			Priority: 1,
		},
	}

	for _, workflow := range defaultWorkflows {
		workflow.Created = time.Now()
		workflow.Updated = time.Now()
		irm.workflows[workflow.ID] = workflow
	}
}

// initializeDefaultActions initializes default response actions
func (irm *IncidentResponseManager) initializeDefaultActions() {
	defaultActions := []*ResponseAction{
		{
			ID:          "isolate_system",
			Name:        "Isolate System",
			Type:        "system_isolation",
			Description: "Isolate a system from the network",
			Parameters: map[string]interface{}{
				"isolation_level": "full",
				"preserve_data":   true,
			},
			Enabled:    true,
			Required:   false,
			Timeout:    30 * time.Second,
			RetryCount: 3,
		},
		{
			ID:          "block_ip",
			Name:        "Block IP Address",
			Type:        "network_blocking",
			Description: "Block an IP address at the firewall",
			Parameters: map[string]interface{}{
				"duration": "24h",
				"scope":    "global",
			},
			Enabled:    true,
			Required:   false,
			Timeout:    10 * time.Second,
			RetryCount: 3,
		},
		{
			ID:          "disable_account",
			Name:        "Disable User Account",
			Type:        "account_management",
			Description: "Disable a user account",
			Parameters: map[string]interface{}{
				"reason": "security_incident",
			},
			Enabled:    true,
			Required:   false,
			Timeout:    15 * time.Second,
			RetryCount: 2,
		},
		{
			ID:          "security_scan",
			Name:        "Security Scan",
			Type:        "security_scanning",
			Description: "Run a security scan on the system",
			Parameters: map[string]interface{}{
				"scan_type":  "full",
				"quarantine": true,
			},
			Enabled:    true,
			Required:   false,
			Timeout:    5 * time.Minute,
			RetryCount: 2,
		},
		{
			ID:          "rotate_credentials",
			Name:        "Rotate Credentials",
			Type:        "credential_management",
			Description: "Rotate system credentials",
			Parameters: map[string]interface{}{
				"scope": "all",
			},
			Enabled:    true,
			Required:   false,
			Timeout:    2 * time.Minute,
			RetryCount: 2,
		},
		{
			ID:          "activate_ddos_protection",
			Name:        "Activate DDoS Protection",
			Type:        "network_protection",
			Description: "Activate DDoS protection mechanisms",
			Parameters: map[string]interface{}{
				"protection_level": "high",
				"rate_limit":       "1000/s",
			},
			Enabled:    true,
			Required:   false,
			Timeout:    30 * time.Second,
			RetryCount: 3,
		},
	}

	for _, action := range defaultActions {
		irm.actions[action.ID] = action
	}
}

// initializeDefaultEngines initializes default response engines
func (irm *IncidentResponseManager) initializeDefaultEngines() {
	engine := &ResponseEngine{
		Name:      "default",
		Actions:   irm.actions,
		Executors: make(map[string]ActionExecutor),
	}

	// Initialize action executors
	engine.Executors["system_isolation"] = &SystemIsolationExecutor{}
	engine.Executors["network_blocking"] = &NetworkBlockingExecutor{}
	engine.Executors["account_management"] = &AccountManagementExecutor{}
	engine.Executors["security_scanning"] = &SecurityScanningExecutor{}
	engine.Executors["credential_management"] = &CredentialManagementExecutor{}
	engine.Executors["network_protection"] = &NetworkProtectionExecutor{}

	irm.engines["default"] = engine
}

// ProcessThreatAlert processes a threat alert and triggers appropriate workflows
func (irm *IncidentResponseManager) ProcessThreatAlert(ctx context.Context, alert *threat.ThreatAlert) (*Incident, error) {
	// Create incident
	incident := &Incident{
		ID:              fmt.Sprintf("incident_%d", time.Now().UnixNano()),
		Title:           fmt.Sprintf("%s Incident", alert.ThreatType),
		Description:     alert.Description,
		Severity:        alert.Severity,
		Status:          "new",
		Priority:        irm.calculatePriority(alert),
		Source:          "automated",
		ThreatAlert:     alert,
		Created:         time.Now(),
		Updated:         time.Now(),
		Tags:            []string{alert.ThreatType, "automated"},
		Actions:         make([]IncidentAction, 0),
		WorkflowSteps:   make([]WorkflowStepResult, 0),
		Evidence:        make([]Evidence, 0),
		Notes:           make([]IncidentNote, 0),
		Metadata:        make(map[string]interface{}),
		EscalationLevel: 0,
	}

	// Store incident
	irm.mu.Lock()
	irm.activeIncidents[incident.ID] = incident
	irm.mu.Unlock()

	// Find matching workflows
	matchingWorkflows := irm.findMatchingWorkflows(alert)

	// Execute workflows
	for _, workflow := range matchingWorkflows {
		if workflow.Enabled {
			go irm.executeWorkflow(ctx, workflow, incident)
		}
	}

	// Check for escalation
	go irm.checkEscalation(ctx, incident)

	return incident, nil
}

// findMatchingWorkflows finds workflows that match the threat alert
func (irm *IncidentResponseManager) findMatchingWorkflows(alert *threat.ThreatAlert) []*Workflow {
	var matching []*Workflow

	for _, workflow := range irm.workflows {
		if irm.workflowMatches(workflow, alert) {
			matching = append(matching, workflow)
		}
	}

	// Sort by priority
	for i := 0; i < len(matching)-1; i++ {
		for j := i + 1; j < len(matching); j++ {
			if matching[i].Priority > matching[j].Priority {
				matching[i], matching[j] = matching[j], matching[i]
			}
		}
	}

	return matching
}

// workflowMatches checks if a workflow matches a threat alert
func (irm *IncidentResponseManager) workflowMatches(workflow *Workflow, alert *threat.ThreatAlert) bool {
	for _, trigger := range workflow.Triggers {
		if irm.evaluateTrigger(trigger, alert) {
			return true
		}
	}
	return false
}

// evaluateTrigger evaluates a trigger condition
func (irm *IncidentResponseManager) evaluateTrigger(trigger Trigger, alert *threat.ThreatAlert) bool {
	var fieldValue interface{}

	switch trigger.Field {
	case "threat_type":
		fieldValue = alert.ThreatType
	case "severity":
		fieldValue = alert.Severity
	case "confidence":
		fieldValue = alert.Confidence
	case "source_ip":
		fieldValue = alert.SourceIP
	default:
		return false
	}

	return irm.evaluateCondition(trigger.Condition, fieldValue, trigger.Value)
}

// evaluateCondition evaluates a condition
func (irm *IncidentResponseManager) evaluateCondition(operator string, fieldValue, conditionValue interface{}) bool {
	switch operator {
	case "equals":
		return fieldValue == conditionValue
	case "contains":
		fieldStr, fieldOk := fieldValue.(string)
		conditionStr, conditionOk := conditionValue.(string)
		if fieldOk && conditionOk {
			return strings.Contains(fieldStr, conditionStr)
		}
	case "greater_than":
		fieldFloat, fieldOk := fieldValue.(float64)
		conditionFloat, conditionOk := conditionValue.(float64)
		if fieldOk && conditionOk {
			return fieldFloat > conditionFloat
		}
	case "less_than":
		fieldFloat, fieldOk := fieldValue.(float64)
		conditionFloat, conditionOk := conditionValue.(float64)
		if fieldOk && conditionOk {
			return fieldFloat < conditionFloat
		}
	}

	return false
}

// executeWorkflow executes a workflow for an incident
func (irm *IncidentResponseManager) executeWorkflow(ctx context.Context, workflow *Workflow, incident *Incident) {
	log.Printf("Executing workflow %s for incident %s", workflow.ID, incident.ID)

	// Update incident status
	irm.updateIncidentStatus(incident.ID, "investigating")

	// Execute workflow steps
	for _, step := range workflow.Steps {
		if irm.shouldExecuteStep(step, incident) {
			result := irm.executeWorkflowStep(ctx, step, incident)
			irm.addWorkflowStepResult(incident.ID, result)

			if result.Status == "failed" && step.OnFailure != "" {
				// Execute failure step
				if failureStep := irm.findStepByID(step.OnFailure, workflow); failureStep != nil {
					irm.executeWorkflowStep(ctx, *failureStep, incident)
				}
			}

			// Determine next steps
			nextSteps := irm.getNextSteps(step, result, workflow)
			for _, nextStepID := range nextSteps {
				if nextStep := irm.findStepByID(nextStepID, workflow); nextStep != nil {
					irm.executeWorkflowStep(ctx, *nextStep, incident)
				}
			}
		}
	}

	// Update incident status
	irm.updateIncidentStatus(incident.ID, "resolved")
}

// shouldExecuteStep determines if a step should be executed
func (irm *IncidentResponseManager) shouldExecuteStep(step WorkflowStep, incident *Incident) bool {
	// Check conditions
	for _, condition := range step.Conditions {
		if !irm.evaluateStepCondition(condition, incident) {
			return false
		}
	}

	return true
}

// evaluateStepCondition evaluates a step condition
func (irm *IncidentResponseManager) evaluateStepCondition(condition Condition, incident *Incident) bool {
	var fieldValue interface{}

	switch condition.Field {
	case "severity":
		fieldValue = incident.Severity
	case "priority":
		fieldValue = incident.Priority
	case "status":
		fieldValue = incident.Status
	default:
		return false
	}

	return irm.evaluateCondition(condition.Operator, fieldValue, condition.Value)
}

// executeWorkflowStep executes a single workflow step
func (irm *IncidentResponseManager) executeWorkflowStep(ctx context.Context, step WorkflowStep, incident *Incident) *WorkflowStepResult {
	start := time.Now()

	log.Printf("Executing workflow step %s: %s", step.ID, step.Name)

	result := &WorkflowStepResult{
		StepID:     step.ID,
		Status:     "executing",
		ExecutedAt: start,
		RetryCount: 0,
		Metadata:   make(map[string]interface{}),
	}

	// Execute step based on type
	switch step.Type {
	case "action":
		actionResult := irm.executeAction(ctx, step, incident)
		result.Result = actionResult
		if actionResult.Success {
			result.Status = "completed"
		} else {
			result.Status = "failed"
			result.Error = actionResult.Error
		}
	case "notification":
		notificationResult := irm.executeNotification(ctx, step, incident)
		result.Result = notificationResult
		if notificationResult.Success {
			result.Status = "completed"
		} else {
			result.Status = "failed"
			result.Error = notificationResult.Error
		}
	case "delay":
		time.Sleep(step.Timeout)
		result.Status = "completed"
		result.Result = &ActionResult{
			Success:    true,
			Message:    "Delay completed",
			ExecutedAt: time.Now(),
			Duration:   step.Timeout,
		}
	default:
		result.Status = "failed"
		result.Error = fmt.Sprintf("Unknown step type: %s", step.Type)
	}

	result.Duration = time.Since(start)
	return result
}

// executeAction executes an action step
func (irm *IncidentResponseManager) executeAction(ctx context.Context, step WorkflowStep, incident *Incident) *ActionResult {
	// Find the action
	action, exists := irm.actions[step.Action]
	if !exists {
		return &ActionResult{
			Success: false,
			Error:   fmt.Sprintf("Action not found: %s", step.Action),
		}
	}

	// Get the appropriate engine and executor
	engine := irm.engines["default"]
	if engine == nil {
		return &ActionResult{
			Success: false,
			Error:   "No response engine available",
		}
	}

	executor, exists := engine.Executors[action.Type]
	if !exists {
		return &ActionResult{
			Success: false,
			Error:   fmt.Sprintf("Executor not found for action type: %s", action.Type),
		}
	}

	// Check if action can be executed
	if !executor.CanExecute(action, incident) {
		return &ActionResult{
			Success: false,
			Error:   "Action cannot be executed for this incident",
		}
	}

	// Execute the action
	result, err := executor.Execute(ctx, action, incident)
	if err != nil {
		return &ActionResult{
			Success: false,
			Error:   err.Error(),
		}
	}

	// Add action to incident
	irm.addIncidentAction(incident.ID, IncidentAction{
		ID:          fmt.Sprintf("action_%d", time.Now().UnixNano()),
		Type:        action.Type,
		Description: action.Description,
		Status:      "completed",
		ExecutedAt:  time.Now(),
		ExecutedBy:  "automated",
		Result:      result,
		Metadata:    make(map[string]interface{}),
	})

	return result
}

// executeNotification executes a notification step
func (irm *IncidentResponseManager) executeNotification(ctx context.Context, step WorkflowStep, incident *Incident) *ActionResult {
	// Create notification
	notification := &Notification{
		Type:      "email",
		Recipient: "security-team@company.com",
		Subject:   fmt.Sprintf("Incident Alert: %s", incident.Title),
		Message:   fmt.Sprintf("Incident %s: %s\n\nSeverity: %s\nDescription: %s", incident.ID, incident.Title, incident.Severity, incident.Description),
		Data: map[string]interface{}{
			"incident_id": incident.ID,
			"severity":    incident.Severity,
			"threat_type": incident.ThreatAlert.ThreatType,
		},
		Priority:  "high",
		Timestamp: time.Now(),
	}

	// Send notification
	err := irm.notificationService.Send(ctx, notification)
	if err != nil {
		return &ActionResult{
			Success: false,
			Error:   err.Error(),
		}
	}

	return &ActionResult{
		Success:    true,
		Message:    "Notification sent successfully",
		ExecutedAt: time.Now(),
		Duration:   0,
		Data: map[string]interface{}{
			"notification_id": notification.Type,
		},
	}
}

// findStepByID finds a step by ID in a workflow
func (irm *IncidentResponseManager) findStepByID(stepID string, workflow *Workflow) *WorkflowStep {
	for _, step := range workflow.Steps {
		if step.ID == stepID {
			return &step
		}
	}
	return nil
}

// getNextSteps determines the next steps to execute
func (irm *IncidentResponseManager) getNextSteps(currentStep WorkflowStep, result *WorkflowStepResult, workflow *Workflow) []string {
	if result.Status == "completed" {
		return currentStep.NextSteps
	}

	if currentStep.OnFailure != "" {
		return []string{currentStep.OnFailure}
	}

	return []string{}
}

// calculatePriority calculates incident priority based on threat alert
func (irm *IncidentResponseManager) calculatePriority(alert *threat.ThreatAlert) int {
	priority := 1

	switch alert.Severity {
	case "critical":
		priority = 1
	case "high":
		priority = 2
	case "medium":
		priority = 3
	case "low":
		priority = 4
	}

	// Adjust based on confidence
	if alert.Confidence > 0.9 {
		priority -= 1
	} else if alert.Confidence < 0.5 {
		priority += 1
	}

	// Ensure priority is within valid range
	if priority < 1 {
		priority = 1
	} else if priority > 4 {
		priority = 4
	}

	return priority
}

// updateIncidentStatus updates the status of an incident
func (irm *IncidentResponseManager) updateIncidentStatus(incidentID, status string) {
	irm.mu.Lock()
	defer irm.mu.Unlock()

	if incident, exists := irm.activeIncidents[incidentID]; exists {
		incident.Status = status
		incident.Updated = time.Now()
	}
}

// addWorkflowStepResult adds a workflow step result to an incident
func (irm *IncidentResponseManager) addWorkflowStepResult(incidentID string, result *WorkflowStepResult) {
	irm.mu.Lock()
	defer irm.mu.Unlock()

	if incident, exists := irm.activeIncidents[incidentID]; exists {
		incident.WorkflowSteps = append(incident.WorkflowSteps, *result)
		incident.Updated = time.Now()
	}
}

// addIncidentAction adds an action to an incident
func (irm *IncidentResponseManager) addIncidentAction(incidentID string, action IncidentAction) {
	irm.mu.Lock()
	defer irm.mu.Unlock()

	if incident, exists := irm.activeIncidents[incidentID]; exists {
		incident.Actions = append(incident.Actions, action)
		incident.Updated = time.Now()
	}
}

// checkEscalation checks if an incident should be escalated
func (irm *IncidentResponseManager) checkEscalation(ctx context.Context, incident *Incident) {
	policies := irm.escalationManager.GetMatchingPolicies(incident)

	for _, policy := range policies {
		if policy.Enabled && irm.shouldEscalate(policy, incident) {
			go irm.executeEscalation(ctx, policy, incident)
		}
	}
}

// shouldEscalate determines if an incident should be escalated
func (irm *IncidentResponseManager) shouldEscalate(policy *EscalationPolicy, incident *Incident) bool {
	for _, condition := range policy.Conditions {
		if irm.evaluateEscalationCondition(condition, incident) {
			return true
		}
	}
	return false
}

// evaluateEscalationCondition evaluates an escalation condition
func (irm *IncidentResponseManager) evaluateEscalationCondition(condition EscalationCondition, incident *Incident) bool {
	var fieldValue interface{}

	switch condition.Field {
	case "severity":
		fieldValue = incident.Severity
	case "priority":
		fieldValue = incident.Priority
	case "status":
		fieldValue = incident.Status
	case "escalation_level":
		fieldValue = incident.EscalationLevel
	default:
		return false
	}

	return irm.evaluateCondition(condition.Operator, fieldValue, condition.Value)
}

// executeEscalation executes escalation actions
func (irm *IncidentResponseManager) executeEscalation(ctx context.Context, policy *EscalationPolicy, incident *Incident) {
	log.Printf("Executing escalation policy %s for incident %s", policy.ID, incident.ID)

	// Update escalation level
	irm.mu.Lock()
	incident.EscalationLevel++
	incident.Updated = time.Now()
	irm.mu.Unlock()

	// Execute escalation actions
	for _, action := range policy.Actions {
		if action.Delay > 0 {
			time.Sleep(action.Delay)
		}

		switch action.Type {
		case "notify":
			irm.executeEscalationNotification(ctx, action, incident)
		case "assign":
			irm.executeEscalationAssignment(ctx, action, incident)
		case "escalate":
			irm.executeEscalationLevelIncrease(ctx, action, incident)
		}
	}
}

// executeEscalationNotification executes escalation notification
func (irm *IncidentResponseManager) executeEscalationNotification(ctx context.Context, action EscalationAction, incident *Incident) {
	notification := &Notification{
		Type:      "email",
		Recipient: "management@company.com",
		Subject:   fmt.Sprintf("ESCALATED: Incident %s", incident.ID),
		Message:   fmt.Sprintf("Incident %s has been escalated to level %d", incident.ID, incident.EscalationLevel),
		Data: map[string]interface{}{
			"incident_id":      incident.ID,
			"escalation_level": incident.EscalationLevel,
		},
		Priority:  "critical",
		Timestamp: time.Now(),
	}

	if err := irm.notificationService.Send(ctx, notification); err != nil {
		fmt.Printf("Warning: failed to send notification: %v\n", err)
	}
}

// executeEscalationAssignment executes escalation assignment
func (irm *IncidentResponseManager) executeEscalationAssignment(ctx context.Context, action EscalationAction, incident *Incident) {
	// Assign to senior analyst or management
	assignedTo := "senior-analyst@company.com"

	irm.mu.Lock()
	incident.AssignedTo = assignedTo
	incident.Updated = time.Now()
	irm.mu.Unlock()
}

// executeEscalationLevelIncrease executes escalation level increase
func (irm *IncidentResponseManager) executeEscalationLevelIncrease(ctx context.Context, action EscalationAction, incident *Incident) {
	// Further increase escalation level
	irm.mu.Lock()
	incident.EscalationLevel++
	incident.Updated = time.Now()
	irm.mu.Unlock()
}

// GetIncident returns an incident by ID
func (irm *IncidentResponseManager) GetIncident(incidentID string) (*Incident, bool) {
	irm.mu.RLock()
	defer irm.mu.RUnlock()

	incident, exists := irm.activeIncidents[incidentID]
	return incident, exists
}

// GetActiveIncidents returns all active incidents
func (irm *IncidentResponseManager) GetActiveIncidents() []*Incident {
	irm.mu.RLock()
	defer irm.mu.RUnlock()

	incidents := make([]*Incident, 0, len(irm.activeIncidents))
	for _, incident := range irm.activeIncidents {
		incidents = append(incidents, incident)
	}

	return incidents
}

// GetStats returns incident response statistics
func (irm *IncidentResponseManager) GetStats() map[string]interface{} {
	irm.mu.RLock()
	defer irm.mu.RUnlock()

	stats := map[string]interface{}{
		"total_incidents":     len(irm.activeIncidents),
		"active_workflows":    len(irm.workflows),
		"available_actions":   len(irm.actions),
		"active_engines":      len(irm.engines),
		"escalation_policies": irm.escalationManager.GetPolicyCount(),
	}

	// Count incidents by status
	statusCounts := make(map[string]int)
	severityCounts := make(map[string]int)

	for _, incident := range irm.activeIncidents {
		statusCounts[incident.Status]++
		severityCounts[incident.Severity]++
	}

	stats["incidents_by_status"] = statusCounts
	stats["incidents_by_severity"] = severityCounts

	return stats
}

// CreateIncident creates a new incident manually
func (irm *IncidentResponseManager) CreateIncident(inc *Incident) (*Incident, error) {
	irm.mu.Lock()
	defer irm.mu.Unlock()

	if inc.ID == "" {
		inc.ID = generateIncidentID()
	}
	if inc.Status == "" {
		inc.Status = "new"
	}
	if inc.Source == "" {
		inc.Source = "manual"
	}
	if inc.Created.IsZero() {
		inc.Created = time.Now()
	}
	if inc.Updated.IsZero() {
		inc.Updated = time.Now()
	}

	irm.activeIncidents[inc.ID] = inc

	log.Printf("Incident created: %s - %s", inc.ID, inc.Title)

	return inc, nil
}

// generateIncidentID generates a unique incident ID
func generateIncidentID() string {
	timestamp := time.Now().UnixNano()
	return fmt.Sprintf("INC-%d", timestamp)
}

// NewNotificationService creates a new notification service
func NewNotificationService() *NotificationService {
	return &NotificationService{
		channels: make(map[string]NotificationChannel),
	}
}

// Send sends a notification
func (ns *NotificationService) Send(ctx context.Context, notification *Notification) error {
	// For now, just log the notification
	log.Printf("Notification: %s to %s - %s", notification.Subject, notification.Recipient, notification.Message)
	return nil
}

// NewEscalationManager creates a new escalation manager
func NewEscalationManager() *EscalationManager {
	return &EscalationManager{
		policies: make(map[string]*EscalationPolicy),
	}
}

// GetMatchingPolicies returns policies that match an incident
func (em *EscalationManager) GetMatchingPolicies(incident *Incident) []*EscalationPolicy {
	em.mu.RLock()
	defer em.mu.RUnlock()

	var matching []*EscalationPolicy
	for _, policy := range em.policies {
		// Simple matching - in production, implement more sophisticated matching
		if policy.Enabled {
			matching = append(matching, policy)
		}
	}

	return matching
}

// GetPolicyCount returns the number of policies
func (em *EscalationManager) GetPolicyCount() int {
	em.mu.RLock()
	defer em.mu.RUnlock()

	return len(em.policies)
}

// Action executor implementations (simplified for demonstration)

type SystemIsolationExecutor struct{}

func (s *SystemIsolationExecutor) Name() string {
	return "System Isolation Executor"
}

func (s *SystemIsolationExecutor) Execute(ctx context.Context, action *ResponseAction, incident *Incident) (*ActionResult, error) {
	// Simulate system isolation
	time.Sleep(1 * time.Second)
	return &ActionResult{
		Success:    true,
		Message:    "System isolated successfully",
		ExecutedAt: time.Now(),
		Duration:   1 * time.Second,
	}, nil
}

func (s *SystemIsolationExecutor) CanExecute(action *ResponseAction, incident *Incident) bool {
	return true
}

func (s *SystemIsolationExecutor) Validate(action *ResponseAction) error {
	return nil
}

type NetworkBlockingExecutor struct{}

func (n *NetworkBlockingExecutor) Name() string {
	return "Network Blocking Executor"
}

func (n *NetworkBlockingExecutor) Execute(ctx context.Context, action *ResponseAction, incident *Incident) (*ActionResult, error) {
	// Simulate IP blocking
	time.Sleep(500 * time.Millisecond)
	return &ActionResult{
		Success:    true,
		Message:    "IP blocked successfully",
		ExecutedAt: time.Now(),
		Duration:   500 * time.Millisecond,
	}, nil
}

func (n *NetworkBlockingExecutor) CanExecute(action *ResponseAction, incident *Incident) bool {
	return incident.ThreatAlert != nil && incident.ThreatAlert.SourceIP != ""
}

func (n *NetworkBlockingExecutor) Validate(action *ResponseAction) error {
	return nil
}

type AccountManagementExecutor struct{}

func (a *AccountManagementExecutor) Name() string {
	return "Account Management Executor"
}

func (a *AccountManagementExecutor) Execute(ctx context.Context, action *ResponseAction, incident *Incident) (*ActionResult, error) {
	// Simulate account disable
	time.Sleep(300 * time.Millisecond)
	return &ActionResult{
		Success:    true,
		Message:    "Account disabled successfully",
		ExecutedAt: time.Now(),
		Duration:   300 * time.Millisecond,
	}, nil
}

func (a *AccountManagementExecutor) CanExecute(action *ResponseAction, incident *Incident) bool {
	return true
}

func (a *AccountManagementExecutor) Validate(action *ResponseAction) error {
	return nil
}

type SecurityScanningExecutor struct{}

func (s *SecurityScanningExecutor) Name() string {
	return "Security Scanning Executor"
}

func (s *SecurityScanningExecutor) Execute(ctx context.Context, action *ResponseAction, incident *Incident) (*ActionResult, error) {
	// Simulate security scan
	time.Sleep(2 * time.Second)
	return &ActionResult{
		Success:    true,
		Message:    "Security scan completed",
		ExecutedAt: time.Now(),
		Duration:   2 * time.Second,
		Data: map[string]interface{}{
			"threats_found": 3,
			"scan_duration": "2s",
		},
	}, nil
}

func (s *SecurityScanningExecutor) CanExecute(action *ResponseAction, incident *Incident) bool {
	return true
}

func (s *SecurityScanningExecutor) Validate(action *ResponseAction) error {
	return nil
}

type CredentialManagementExecutor struct{}

func (c *CredentialManagementExecutor) Name() string {
	return "Credential Management Executor"
}

func (c *CredentialManagementExecutor) Execute(ctx context.Context, action *ResponseAction, incident *Incident) (*ActionResult, error) {
	// Simulate credential rotation
	time.Sleep(1 * time.Second)
	return &ActionResult{
		Success:    true,
		Message:    "Credentials rotated successfully",
		ExecutedAt: time.Now(),
		Duration:   1 * time.Second,
	}, nil
}

func (c *CredentialManagementExecutor) CanExecute(action *ResponseAction, incident *Incident) bool {
	return true
}

func (c *CredentialManagementExecutor) Validate(action *ResponseAction) error {
	return nil
}

type NetworkProtectionExecutor struct{}

func (n *NetworkProtectionExecutor) Name() string {
	return "Network Protection Executor"
}

func (n *NetworkProtectionExecutor) Execute(ctx context.Context, action *ResponseAction, incident *Incident) (*ActionResult, error) {
	// Simulate DDoS protection activation
	time.Sleep(500 * time.Millisecond)
	return &ActionResult{
		Success:    true,
		Message:    "DDoS protection activated",
		ExecutedAt: time.Now(),
		Duration:   500 * time.Millisecond,
	}, nil
}

func (n *NetworkProtectionExecutor) CanExecute(action *ResponseAction, incident *Incident) bool {
	return incident.ThreatAlert != nil && incident.ThreatAlert.ThreatType == "ddos"
}

func (n *NetworkProtectionExecutor) Validate(action *ResponseAction) error {
	return nil
}
