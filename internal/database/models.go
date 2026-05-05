package database

import (
	"time"
)

// User model for database
type User struct {
	ID          int       `json:"id" db:"id"`
	Username    string    `json:"username" db:"username"`
	Email       string    `json:"email" db:"email"`
	Password    string    `json:"-" db:"password_hash"` // Never return password in JSON
	Role        string    `json:"role" db:"role"`
	Status      string    `json:"status" db:"status"`
	LastLogin   time.Time `json:"last_login" db:"last_login"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
	Permissions []string  `json:"permissions" db:"permissions"`
}

// Threat model for database
type Threat struct {
	ID          int        `json:"id" db:"id"`
	Title       string     `json:"title" db:"title"`
	Description string     `json:"description" db:"description"`
	Severity    string     `json:"severity" db:"severity"`
	Status      string     `json:"status" db:"status"`
	Source      string     `json:"source" db:"source"`
	Target      string     `json:"target" db:"target"`
	DetectedAt  time.Time  `json:"detected_at" db:"detected_at"`
	ResolvedAt  *time.Time `json:"resolved_at" db:"resolved_at"`
	CreatedBy   int        `json:"created_by" db:"created_by"`
	ResolvedBy  *int       `json:"resolved_by" db:"resolved_by"`
	CreatedAt   time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at" db:"updated_at"`
}

// SecurityPolicy model for database
type SecurityPolicy struct {
	ID          int       `json:"id" db:"id"`
	Name        string    `json:"name" db:"name"`
	Description string    `json:"description" db:"description"`
	Category    string    `json:"category" db:"category"`
	Rules       string    `json:"rules" db:"rules"` // JSON string of rules
	Severity    string    `json:"severity" db:"severity"`
	Status      string    `json:"status" db:"status"`
	Enabled     bool      `json:"enabled" db:"enabled"`
	CreatedBy   int       `json:"created_by" db:"created_by"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

// AuditLog model for database
type AuditLog struct {
	ID        int       `json:"id" db:"id"`
	UserID    int       `json:"user_id" db:"user_id"`
	Action    string    `json:"action" db:"action"`
	Resource  string    `json:"resource" db:"resource"`
	Details   string    `json:"details" db:"details"`
	IPAddress string    `json:"ip_address" db:"ip_address"`
	UserAgent string    `json:"user_agent" db:"user_agent"`
	Timestamp time.Time `json:"timestamp" db:"timestamp"`
}

// Webhook model for database
type Webhook struct {
	ID          int        `json:"id" db:"id"`
	Name        string     `json:"name" db:"name"`
	URL         string     `json:"url" db:"url"`
	Events      []string   `json:"events" db:"events"` // JSON array of events
	Secret      string     `json:"secret" db:"secret"`
	Active      bool       `json:"active" db:"active"`
	LastTrigger *time.Time `json:"last_triggered" db:"last_triggered"`
	CreatedBy   int        `json:"created_by" db:"created_by"`
	CreatedAt   time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at" db:"updated_at"`
}

// SystemMetrics model for database
type SystemMetrics struct {
	ID            int       `json:"id" db:"id"`
	CPUUsage      float64   `json:"cpu_usage" db:"cpu_usage"`
	MemoryUsage   float64   `json:"memory_usage" db:"memory_usage"`
	DiskUsage     float64   `json:"disk_usage" db:"disk_usage"`
	NetworkIn     int64     `json:"network_in" db:"network_in"`
	NetworkOut    int64     `json:"network_out" db:"network_out"`
	ActiveUsers   int       `json:"active_users" db:"active_users"`
	TotalRequests int       `json:"total_requests" db:"total_requests"`
	ErrorRate     float64   `json:"error_rate" db:"error_rate"`
	Timestamp     time.Time `json:"timestamp" db:"timestamp"`
}

// Notification model for database
type Notification struct {
	ID        int       `json:"id" db:"id"`
	UserID    int       `json:"user_id" db:"user_id"`
	Title     string    `json:"title" db:"title"`
	Message   string    `json:"message" db:"message"`
	Type      string    `json:"type" db:"type"`
	Severity  string    `json:"severity" db:"severity"`
	Read      bool      `json:"read" db:"read"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// Migration model for database
type Migration struct {
	ID        int       `json:"id" db:"id"`
	Version   string    `json:"version" db:"version"`
	Name      string    `json:"name" db:"name"`
	AppliedAt time.Time `json:"applied_at" db:"applied_at"`
}

// WorkerTask model for distributed processing
type WorkerTask struct {
	ID          int                    `json:"id" db:"id"`
	Type        string                 `json:"type" db:"type"`
	Status      string                 `json:"status" db:"status"`
	Priority    int                    `json:"priority" db:"priority"`
	Payload     map[string]interface{} `json:"payload" db:"payload"` // JSON object
	Attempted   int                    `json:"attempted" db:"attempted"`
	MaxAttempts int                    `json:"max_attempts" db:"max_attempts"`
	LastError   string                 `json:"last_error" db:"last_error"`
	AssignedTo  *int                   `json:"assigned_to" db:"assigned_to"`
	CreatedBy   int                    `json:"created_by" db:"created_by"`
	CreatedAt   time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at" db:"updated_at"`
	CompletedAt *time.Time             `json:"completed_at" db:"completed_at"`
}

// TaskType represents the type of task being tracked
type TaskType string

const (
	TaskTypeScan     TaskType = "scan"
	TaskTypeIncident TaskType = "incident"
)

// TaskStatus represents the status of a tracked task
type TaskStatus string

const (
	TaskStatusPending   TaskStatus = "pending"
	TaskStatusRunning   TaskStatus = "running"
	TaskStatusCompleted TaskStatus = "completed"
	TaskStatusFailed    TaskStatus = "failed"
	TaskStatusCancelled TaskStatus = "cancelled"
)

// GlobalState tracks the status of every scan and incident to prevent redundant agent tasks
type GlobalState struct {
	ID            int                    `json:"id" db:"id"`
	TaskID        string                 `json:"task_id" db:"task_id"`
	TaskType      TaskType               `json:"task_type" db:"task_type"` // "scan" or "incident"
	Status        TaskStatus             `json:"status" db:"status"`
	Target        string                 `json:"target" db:"target"` // Target identifier (IP, hostname, etc.)
	TargetType    string                 `json:"target_type" db:"target_type"`
	AgentID       string                 `json:"agent_id" db:"agent_id"` // Agent currently handling this task
	ModuleName    string                 `json:"module_name" db:"module_name"`
	PolicyID      string                 `json:"policy_id" db:"policy_id"`     // For scans
	WorkflowID    string                 `json:"workflow_id" db:"workflow_id"` // For incidents
	Severity      string                 `json:"severity" db:"severity"`
	ErrorMessage  string                 `json:"error_message" db:"error_message"`
	ResultSummary string                 `json:"result_summary" db:"result_summary"` // JSON summary of results
	Metadata      map[string]interface{} `json:"metadata" db:"metadata"`
	StartedAt     time.Time              `json:"started_at" db:"started_at"`
	CompletedAt   *time.Time             `json:"completed_at" db:"completed_at"`
	CreatedAt     time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time              `json:"updated_at" db:"updated_at"`
}

// DecisionType represents the type of agent decision
type DecisionType string

const (
	DecisionTypeLaunchModule DecisionType = "launch_module"
	DecisionTypeBlockIP      DecisionType = "block_ip"
	DecisionTypeIsolateNode  DecisionType = "isolate_node"
	DecisionTypeEscalate     DecisionType = "escalate"
	DecisionTypeDismiss      DecisionType = "dismiss"
	DecisionTypeRetry        DecisionType = "retry"
	DecisionTypeRollback     DecisionType = "rollback"
)

// DecisionReason represents the reasoning behind an agent decision
type DecisionReason struct {
	TriggerEvent    string                 `json:"trigger_event"`
	Analysis        string                 `json:"analysis"`
	Confidence      float64                `json:"confidence"`
	ThreatLevel     string                 `json:"threat_level"`
	Recommendations []string               `json:"recommendations"`
	Factors         map[string]interface{} `json:"factors"`
}

// AgentDecision tracks autonomous agent decisions for audit purposes
type AgentDecision struct {
	ID            int          `json:"id" db:"id"`
	DecisionID    string       `json:"decision_id" db:"decision_id"`
	AgentID       string       `json:"agent_id" db:"agent_id"`
	AgentType     string       `json:"agent_type" db:"agent_type"` // orchestrator, threat_engine, etc.
	DecisionType  DecisionType `json:"decision_type" db:"decision_type"`
	Target        string       `json:"target" db:"target"`
	TargetType    string       `json:"target_type" db:"target_type"` // ip, domain, user, etc.
	Action        string       `json:"action" db:"action"`           // module name, block, isolate, etc.
	Reason        string       `json:"reason" db:"reason"`           // JSON of DecisionReason
	Confidence    float64      `json:"confidence" db:"confidence"`
	ThreatLevel   string       `json:"threat_level" db:"threat_level"`
	TriggerEvent  string       `json:"trigger_event" db:"trigger_event"`
	Status        string       `json:"status" db:"status"` // pending, executed, failed, rolled_back
	Result        string       `json:"result" db:"result"`
	ErrorMessage  string       `json:"error_message" db:"error_message"`
	ExecutionTime int64        `json:"execution_time" db:"execution_time"` // milliseconds
	CreatedAt     time.Time    `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time    `json:"updated_at" db:"updated_at"`
}

// CloudProvider represents a cloud provider with known IP ranges
type CloudProvider struct {
	ID          int       `json:"id" db:"id"`
	Name        string    `json:"name" db:"name"`       // AWS, GCP, Azure, etc.
	CIDR        string    `json:"cidr" db:"cidr"`       // IP range in CIDR notation
	Region      string    `json:"region" db:"region"`   // e.g., us-east-1
	Service     string    `json:"service" db:"service"` // e.g., EC2, S3, Compute
	Description string    `json:"description" db:"description"`
	Priority    int       `json:"priority" db:"priority"` // Scan priority (higher = more important)
	Enabled     bool      `json:"enabled" db:"enabled"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

// SafeList represents IPs/domains that should be excluded from scanning
type SafeList struct {
	ID         int        `json:"id" db:"id"`
	Target     string     `json:"target" db:"target"`           // IP, CIDR, or domain
	TargetType string     `json:"target_type" db:"target_type"` // ip, cidr, domain
	Reason     string     `json:"reason" db:"reason"`           // Why it's safe
	Category   string     `json:"category" db:"category"`       // internal, partner, critical_infrastructure
	CreatedBy  int        `json:"created_by" db:"created_by"`
	ExpiresAt  *time.Time `json:"expires_at" db:"expires_at"` // Optional expiration
	CreatedAt  time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at" db:"updated_at"`
}

// AttackVectorStatus represents the status of an attack vector
type AttackVectorStatus string

const (
	AttackVectorStatusActive   AttackVectorStatus = "active"
	AttackVectorStatusBurned   AttackVectorStatus = "burned"
	AttackVectorStatusCooldown AttackVectorStatus = "cooldown"
)

// AttackVector tracks the status of exploit vectors per target
type AttackVector struct {
	ID          int                    `json:"id" db:"id"`
	Target      string                 `json:"target" db:"target"`
	VectorType  string                 `json:"vector_type" db:"vector_type"` // e.g., "sqli", "xss", "rce"
	Status      AttackVectorStatus     `json:"status" db:"status"`
	Reason      string                 `json:"reason" db:"reason"`
	LastAttempt time.Time              `json:"last_attempt" db:"last_attempt"`
	Attempts    int                    `json:"attempts" db:"attempts"`
	Metadata    map[string]interface{} `json:"metadata" db:"metadata"`
	CreatedAt   time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at" db:"updated_at"`
}

// ScanMode represents the scanning mode for a target
type ScanMode string

const (
	ScanModeStandard   ScanMode = "standard"
	ScanModeObfuscated ScanMode = "obfuscated"
	ScanModeStealth    ScanMode = "stealth"
)

// ScanModeConfig tracks scanning mode per target
type ScanModeConfig struct {
	ID          int                    `json:"id" db:"id"`
	Target      string                 `json:"target" db:"target"`
	Mode        ScanMode               `json:"mode" db:"mode"`
	Reason      string                 `json:"reason" db:"reason"`
	TriggeredBy string                 `json:"triggered_by" db:"triggered_by"`
	Settings    map[string]interface{} `json:"settings" db:"settings"`
	ExpiresAt   *time.Time             `json:"expires_at" db:"expires_at"`
	CreatedAt   time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at" db:"updated_at"`
}

// GovernorAction tracks automated actions for the Safety Governor
type GovernorAction struct {
	ID                int       `json:"id" db:"id"`
	ActionID          string    `json:"action_id" db:"action_id"`
	ActionName        string    `json:"action_name" db:"action_name"`
	Target            string    `json:"target" db:"target"`
	Reasoning         string    `json:"reasoning" db:"reasoning"`
	Requester         string    `json:"requester" db:"requester"`
	Status            string    `json:"status" db:"status"` // approved, blocked, manual_ack_required
	RequiresApproval  bool      `json:"requires_approval" db:"requires_approval"`
	Approved          bool      `json:"approved" db:"approved"`
	RequiresManualAck bool      `json:"requires_manual_ack" db:"requires_manual_ack"`
	BlockReason       string    `json:"block_reason" db:"block_reason"`
	ExecutionTime     int64     `json:"execution_time" db:"execution_time"` // milliseconds
	Metadata          string    `json:"metadata" db:"metadata"`             // JSON string
	CreatedAt         time.Time `json:"created_at" db:"created_at"`
	UpdatedAt         time.Time `json:"updated_at" db:"updated_at"`
}
