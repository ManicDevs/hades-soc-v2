package ai

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os/exec"
	"strings"
	"sync"
	"time"
)

type AgentTask struct {
	ID          string                 `json:"id"`
	Description string                 `json:"description"`
	Command     string                 `json:"command,omitempty"`
	Context     map[string]interface{} `json:"context,omitempty"`
	Status      string                 `json:"status"`
	Result      string                 `json:"result,omitempty"`
	Error       string                 `json:"error,omitempty"`
	CreatedAt   time.Time              `json:"created_at"`
	CompletedAt time.Time              `json:"completed_at,omitempty"`
}

type Agent struct {
	ID           string      `json:"id"`
	Name         string      `json:"name"`
	Role         string      `json:"role"`
	Tasks        []AgentTask `json:"tasks"`
	mu           sync.RWMutex
	llm          *LLMService
	fileManager  *FileOperationManager `json:"-"`
	Autonomous   bool                  `json:"autonomous"`
	Capabilities []string              `json:"capabilities"`
}

type AgentSystem struct {
	agents      map[string]*Agent
	workflows   []Workflow
	mu          sync.RWMutex
	llm         *LLMService
	fileManager *FileOperationManager
	baseDir     string
}

type Workflow struct {
	ID      string         `json:"id"`
	Name    string         `json:"name"`
	Steps   []WorkflowStep `json:"steps"`
	Status  string         `json:"status"`
	LastRun time.Time      `json:"last_run,omitempty"`
	Results []string       `json:"results,omitempty"`
}

type WorkflowStep struct {
	Name      string `json:"name"`
	Action    string `json:"action"` // analyze, fix, alert, block
	Target    string `json:"target"`
	Condition string `json:"condition,omitempty"`
	OnSuccess string `json:"on_success,omitempty"`
	OnFailure string `json:"on_failure,omitempty"`
}

var agentSystem *AgentSystem

type AutonomousLoop struct {
	interval   time.Duration
	enabled    bool
	stopChan   chan struct{}
	running    bool
	mu         sync.RWMutex
	onDecision func(string, map[string]interface{})
}

func InitAgentSystem() *AgentSystem {
	return InitAgentSystemWithBaseDir(".")
}

func InitAgentSystemWithBaseDir(baseDir string) *AgentSystem {
	agentSystem = &AgentSystem{
		agents:      make(map[string]*Agent),
		workflows:   make([]Workflow, 0),
		llm:         GetLLMService(),
		fileManager: NewFileOperationManager(baseDir),
		baseDir:     baseDir,
	}
	agentSystem.createDefaultAgents()
	agentSystem.createDefaultWorkflows()
	return agentSystem
}

func GetAgentSystem() *AgentSystem {
	if agentSystem == nil {
		return InitAgentSystem()
	}
	return agentSystem
}

func (s *AgentSystem) createDefaultAgents() {
	agents := []struct {
		id, name, role string
		capabilities   []string
	}{
		{"analyst", "Security Analyst", "Analyzes security events and threats", []string{"analyze", "report", "read_file"}},
		{"responder", "Incident Responder", "Responds to security incidents", []string{"respond", "contain", "edit_file", "create_file"}},
		{"monitor", "System Monitor", "Monitors system health", []string{"monitor", "alert", "read_file"}},
		{"hunter", "Threat Hunter", "Proactively hunts for threats", []string{"hunt", "investigate", "edit_file", "create_file"}},
		{"autonomous", "Autonomous Agent", "Performs autonomous file operations", []string{"read", "write", "edit", "create", "delete", "list"}},
	}

	for _, a := range agents {
		s.agents[a.id] = &Agent{
			ID:           a.id,
			Name:         a.name,
			Role:         a.role,
			Tasks:        make([]AgentTask, 0),
			llm:          s.llm,
			fileManager:  s.fileManager,
			Autonomous:   a.id == "autonomous",
			Capabilities: a.capabilities,
		}
	}
}

func (s *AgentSystem) createDefaultWorkflows() {
	s.workflows = []Workflow{
		{
			ID:   "threat-response",
			Name: "Threat Response Workflow",
			Steps: []WorkflowStep{
				{Action: "analyze", Target: "threat", Condition: "severity >= high"},
				{Action: "contain", Target: "source"},
				{Action: "alert", Target: "admin"},
				{Action: "log", Target: "incident"},
			},
		},
		{
			ID:   "system-health",
			Name: "System Health Check",
			Steps: []WorkflowStep{
				{Action: "check", Target: "disk"},
				{Action: "check", Target: "memory"},
				{Action: "fix", Target: "auto", Condition: "status == unhealthy"},
				{Action: "report", Target: "admin"},
			},
		},
		{
			ID:   "breach-response",
			Name: "Breach Response",
			Steps: []WorkflowStep{
				{Action: "isolate", Target: "affected-systems"},
				{Action: "analyze", Target: "scope"},
				{Action: "notify", Target: "security-team"},
				{Action: "collect", Target: "evidence"},
				{Action: "remediate", Target: "threat"},
			},
		},
		{
			ID:   "autonomous-file-ops",
			Name: "Autonomous File Operations",
			Steps: []WorkflowStep{
				{Action: "read_file", Target: "config"},
				{Action: "analyze", Target: "content"},
				{Action: "edit_file", Target: "config", Condition: "needs_update"},
				{Action: "create_file", Target: "backup", Condition: "modified"},
				{Action: "log", Target: "operation"},
			},
		},
	}
}

func (s *AgentSystem) GetAgents() []*Agent {
	s.mu.RLock()
	defer s.mu.RUnlock()

	agents := make([]*Agent, 0, len(s.agents))
	for _, a := range s.agents {
		agents = append(agents, a)
	}
	return agents
}

func (s *AgentSystem) GetAgent(id string) *Agent {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.agents[id]
}

func (s *AgentSystem) GetWorkflows() []Workflow {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.workflows
}

func (s *AgentSystem) RunTask(agentID, description string) (*AgentTask, error) {
	agent := s.GetAgent(agentID)
	if agent == nil {
		return nil, fmt.Errorf("agent not found: %s", agentID)
	}

	task := &AgentTask{
		ID:          fmt.Sprintf("task-%d", time.Now().UnixNano()),
		Description: description,
		Status:      "running",
		CreatedAt:   time.Now(),
	}

	agent.mu.Lock()
	agent.Tasks = append(agent.Tasks, *task)
	agent.mu.Unlock()

	go s.executeTask(agent, task)

	return task, nil
}

func (s *AgentSystem) executeTask(agent *Agent, task *AgentTask) {
	defer func() {
		task.CompletedAt = time.Now()
	}()

	resp, err := s.llm.Query(task.Description)
	if err != nil {
		task.Status = "failed"
		task.Result = err.Error()
		return
	}

	task.Result = resp.Response
	task.Status = "completed"
}

func (s *AgentSystem) RunWorkflow(workflowID string, context map[string]interface{}) (string, error) {
	s.mu.RLock()
	var workflow *Workflow
	for i := range s.workflows {
		if s.workflows[i].ID == workflowID {
			w := s.workflows[i]
			workflow = &w
			break
		}
	}
	s.mu.RUnlock()

	if workflow == nil {
		return "", fmt.Errorf("workflow not found: %s", workflowID)
	}

	var results []string
	for _, step := range workflow.Steps {
		result, err := s.executeStep(step, context)
		results = append(results, fmt.Sprintf("[%s] %s: %v", step.Name, step.Action, result))
		if err != nil {
			if step.OnFailure != "" {
				results = append(results, fmt.Sprintf("[FAILURE HANDLER] %s", step.OnFailure))
			}
			break
		}
	}

	workflow.LastRun = time.Now()
	workflow.Results = results
	workflow.Status = "completed"

	return strings.Join(results, "\n"), nil
}

func (s *AgentSystem) executeStep(step WorkflowStep, context map[string]interface{}) (string, error) {
	switch step.Action {
	case "analyze":
		target := step.Target
		if v, ok := context["data"].(string); ok {
			target = v
		}
		resp, err := s.llm.Diagnose(target)
		if err != nil {
			return "", err
		}
		return resp, nil

	case "fix":
		target := step.Target
		if v, ok := context["issue"].(string); ok {
			target = v
		}
		resp, err := s.llm.Diagnose(target)
		if err != nil {
			return "", err
		}

		var fix struct {
			Fix string `json:"fix"`
		}
		json.Unmarshal([]byte(resp), &fix)
		if fix.Fix != "" {
			exec.Command("bash", "-c", fix.Fix).CombinedOutput()
			return fmt.Sprintf("Executed: %s", fix.Fix), nil
		}
		return "No auto-fix available", nil

	case "check":
		cmd := exec.Command("bash", "-c", fmt.Sprintf("echo 'Checking %s'", step.Target))
		out, _ := cmd.Output()
		return string(out), nil

	case "alert":
		msg := fmt.Sprintf("Alert triggered for %s", step.Target)
		return msg, nil

	case "contain":
		return fmt.Sprintf("Contained: %s", step.Target), nil

	case "block":
		return fmt.Sprintf("Blocked: %s", step.Target), nil

	case "log":
		return fmt.Sprintf("Logged: %s", step.Target), nil

	case "notify":
		return fmt.Sprintf("Notified: %s", step.Target), nil

	case "isolate":
		return fmt.Sprintf("Isolated: %s", step.Target), nil

	case "collect":
		return fmt.Sprintf("Evidence collected: %s", step.Target), nil

	case "remediate":
		return fmt.Sprintf("Remediated: %s", step.Target), nil

	case "report":
		data, _ := json.Marshal(context)
		resp, err := s.llm.GenerateReport(map[string]interface{}{
			"data": string(data),
		})
		if err != nil {
			return "", err
		}
		return resp, nil

	case "read_file":
		if s.fileManager == nil {
			return "File manager not available", nil
		}
		path := step.Target
		if v, ok := context["file_path"].(string); ok {
			path = v
		}
		op, err := s.fileManager.ReadFile(path)
		if err != nil {
			return fmt.Sprintf("Failed to read file: %v", err), nil
		}
		context["file_content"] = op.Result
		return fmt.Sprintf("Read file: %s (%d bytes)", path, len(op.Result)), nil

	case "edit_file":
		if s.fileManager == nil {
			return "File manager not available", nil
		}
		path := step.Target
		instruction := "Update file based on analysis"
		if v, ok := context["edit_instruction"].(string); ok {
			instruction = v
		}
		req := FileEditRequest{
			Path:        path,
			Instruction: instruction,
			ManualACK:   true, // Autonomous mode allows auto-ACK
		}
		op, err := s.fileManager.EditFile(req)
		if err != nil {
			return fmt.Sprintf("Failed to edit file: %v", err), nil
		}
		context["file_modified"] = true
		return op.Result, nil

	case "create_file":
		if s.fileManager == nil {
			return "File manager not available", nil
		}
		path := step.Target
		content := ""
		if v, ok := context["file_content"].(string); ok {
			content = v
		}
		op, err := s.fileManager.CreateFile(path, content, true)
		if err != nil {
			return fmt.Sprintf("Failed to create file: %v", err), nil
		}
		return op.Result, nil

	case "delete_file":
		if s.fileManager == nil {
			return "File manager not available", nil
		}
		path := step.Target
		op, err := s.fileManager.DeleteFile(path, true)
		if err != nil {
			return fmt.Sprintf("Failed to delete file: %v", err), nil
		}
		return op.Result, nil

	case "list_files":
		if s.fileManager == nil {
			return "File manager not available", nil
		}
		path := step.Target
		if path == "" {
			path = s.baseDir
		}
		_, err := s.fileManager.ListDirectory(path)
		if err != nil {
			return fmt.Sprintf("Failed to list directory: %v", err), nil
		}
		return fmt.Sprintf("Listed directory: %s", path), nil

	default:
		return fmt.Sprintf("Unknown action: %s", step.Action), nil
	}
}

func (s *AgentSystem) AnalyzeThreat(threatData string) (map[string]interface{}, error) {
	prompt := fmt.Sprintf(`Analyze this threat and provide structured JSON response:

Threat Data: %s

Provide JSON:
{
  "threat_level": "low/medium/high/critical",
  "category": "malware/intrusion/phishing/dos/data_breach/unknown",
  "indicators": ["IOCs"],
  "recommended_actions": ["actions to take"],
  "confidence": 0-100
}`, threatData)

	resp, err := s.llm.Query(prompt)
	if err != nil {
		return nil, err
	}

	var result map[string]interface{}
	json.Unmarshal([]byte(resp.Response), &result)

	if result == nil {
		result = map[string]interface{}{
			"threat_level":        "unknown",
			"category":            "unknown",
			"indicators":          []string{},
			"recommended_actions": []string{resp.Response},
			"confidence":          50,
		}
	}

	return result, nil
}

func (s *AgentSystem) AutonomousResponse(incident string) (map[string]interface{}, error) {
	prompt := fmt.Sprintf(`This security incident requires autonomous response:

Incident: %s

Based on the severity and type, determine:
1. Immediate containment actions
2. Investigation steps
3. Communication requirements
4. Recovery steps

Provide JSON:
{
  "containment": ["immediate actions"],
  "investigation": ["investigation steps"],
  "communication": ["who to notify"],
  "recovery": ["recovery steps"],
  "estimated_time": "time to resolve"
}`, incident)

	resp, err := s.llm.Query(prompt)
	if err != nil {
		return nil, err
	}

	var result map[string]interface{}
	json.Unmarshal([]byte(resp.Response), &result)

	if result == nil {
		result = map[string]interface{}{
			"containment":    []string{},
			"investigation":  []string{},
			"communication":  []string{},
			"recovery":       []string{},
			"estimated_time": "unknown",
		}
	}

	return result, nil
}

func (s *AgentSystem) HuntThreats() ([]string, error) {
	prompt := `Based on current threat intelligence and common attack patterns, what should I hunt for?

Consider:
- Recent CVE disclosures
- Common attack vectors
- Your environment (web servers, databases, user endpoints)

Provide a prioritized list of hunt hypotheses.`

	resp, err := s.llm.Query(prompt)
	if err != nil {
		return nil, err
	}

	lines := strings.Split(resp.Response, "\n")
	var hypotheses []string
	for _, line := range lines {
		if strings.TrimSpace(line) != "" {
			hypotheses = append(hypotheses, strings.TrimSpace(line))
		}
	}

	return hypotheses, nil
}

// AutonomousFileOperation performs an autonomous file operation using AI
func (a *Agent) AutonomousFileOperation(instruction string) (*AgentTask, error) {
	if a.fileManager == nil {
		return nil, fmt.Errorf("file manager not available for agent %s", a.ID)
	}

	// Use LLM to determine the operation
	prompt := fmt.Sprintf(`Analyze this instruction and determine the file operation needed:

Instruction: %s

Provide JSON response:
{
  "operation": "read|write|edit|create|delete|list",
  "path": "file path",
  "content": "content (for write/create)",
  "instruction": "edit instruction (for edit)",
  "reasoning": "why this operation"
}`, instruction)

	resp, err := a.llm.Query(prompt)
	if err != nil {
		return nil, err
	}

	var operation struct {
		Operation   string `json:"operation"`
		Path        string `json:"path"`
		Content     string `json:"content"`
		Instruction string `json:"instruction"`
		Reasoning   string `json:"reasoning"`
	}

	if err := json.Unmarshal([]byte(resp.Response), &operation); err != nil {
		return nil, fmt.Errorf("failed to parse AI response: %w", err)
	}

	task := &AgentTask{
		ID:          fmt.Sprintf("task-%d", time.Now().UnixNano()),
		Description: fmt.Sprintf("Autonomous file operation: %s", instruction),
		Context: map[string]interface{}{
			"operation":  operation.Operation,
			"path":       operation.Path,
			"reasoning":  operation.Reasoning,
			"autonomous": true,
		},
		Status:    "running",
		CreatedAt: time.Now(),
	}

	a.mu.Lock()
	a.Tasks = append(a.Tasks, *task)
	a.mu.Unlock()

	// Execute the operation
	var result string
	switch operation.Operation {
	case "read":
		op, err := a.fileManager.ReadFile(operation.Path)
		if err != nil {
			task.Status = "failed"
			task.Error = err.Error()
			task.CompletedAt = time.Now()
			return task, err
		}
		result = op.Result
		task.Context["file_content"] = op.Result

	case "write":
		op, err := a.fileManager.WriteFile(operation.Path, operation.Content, true)
		if err != nil {
			task.Status = "failed"
			task.Error = err.Error()
			task.CompletedAt = time.Now()
			return task, err
		}
		result = op.Result

	case "edit":
		req := FileEditRequest{
			Path:        operation.Path,
			Instruction: operation.Instruction,
			ManualACK:   true,
		}
		op, err := a.fileManager.EditFile(req)
		if err != nil {
			task.Status = "failed"
			task.Error = err.Error()
			task.CompletedAt = time.Now()
			return task, err
		}
		result = op.Result

	case "create":
		op, err := a.fileManager.CreateFile(operation.Path, operation.Content, true)
		if err != nil {
			task.Status = "failed"
			task.Error = err.Error()
			task.CompletedAt = time.Now()
			return task, err
		}
		result = op.Result

	case "delete":
		op, err := a.fileManager.DeleteFile(operation.Path, true)
		if err != nil {
			task.Status = "failed"
			task.Error = err.Error()
			task.CompletedAt = time.Now()
			return task, err
		}
		result = op.Result

	case "list":
		op, err := a.fileManager.ListDirectory(operation.Path)
		if err != nil {
			task.Status = "failed"
			task.Error = err.Error()
			task.CompletedAt = time.Now()
			return task, err
		}
		result = op.Result

	default:
		task.Status = "failed"
		task.Error = fmt.Sprintf("unknown operation: %s", operation.Operation)
		task.CompletedAt = time.Now()
		return task, errors.New(task.Error)
	}

	task.Status = "completed"
	task.Result = result
	task.CompletedAt = time.Now()

	a.mu.Lock()
	a.Tasks[len(a.Tasks)-1] = *task
	a.mu.Unlock()

	return task, nil
}

// InteractiveFileEdit performs an interactive AI-assisted file edit
func (a *Agent) InteractiveFileEdit(path, instruction string) (*AgentTask, error) {
	if a.fileManager == nil {
		return nil, fmt.Errorf("file manager not available for agent %s", a.ID)
	}

	task := &AgentTask{
		ID:          fmt.Sprintf("task-%d", time.Now().UnixNano()),
		Description: fmt.Sprintf("Interactive file edit: %s", path),
		Context: map[string]interface{}{
			"path":        path,
			"instruction": instruction,
			"interactive": true,
		},
		Status:    "running",
		CreatedAt: time.Now(),
	}

	a.mu.Lock()
	a.Tasks = append(a.Tasks, *task)
	a.mu.Unlock()

	req := FileEditRequest{
		Path:        path,
		Instruction: instruction,
		ManualACK:   true,
	}

	op, err := a.fileManager.EditFile(req)
	if err != nil {
		task.Status = "failed"
		task.Error = err.Error()
		task.CompletedAt = time.Now()
		a.mu.Lock()
		a.Tasks[len(a.Tasks)-1] = *task
		a.mu.Unlock()
		return task, err
	}

	task.Status = "completed"
	task.Result = op.Result
	task.CompletedAt = time.Now()

	a.mu.Lock()
	a.Tasks[len(a.Tasks)-1] = *task
	a.mu.Unlock()

	return task, nil
}

func (s *AgentSystem) StartAutonomousLoop(intervalSec int, onDecision func(string, map[string]interface{})) {
	go func() {
		interval := time.Duration(intervalSec) * time.Second
		if interval == 0 {
			interval = 30 * time.Second
		}

		log.Printf("Starting autonomous agent loop (interval: %v)", interval)

		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				s.runAutonomousCycle(onDecision)
			}
		}
	}()
}

func (s *AgentSystem) runAutonomousCycle(onDecision func(string, map[string]interface{})) {
	log.Println("Agent: Running autonomous cycle...")

	threatData := "Routine system health check - checking for anomalies"
	result, err := s.AnalyzeThreat(threatData)
	if err != nil {
		log.Printf("Agent: Analysis error: %v", err)
		return
	}

	decision := map[string]interface{}{
		"threat_analysis": result,
		"timestamp":       time.Now().Format(time.RFC3339),
		"actions_taken":   []string{},
	}

	if level, ok := result["threat_level"].(string); ok {
		if level == "critical" || level == "high" {
			log.Printf("Agent: Detected %s threat, initiating autonomous response", level)

			incident := fmt.Sprintf("Detected %s threat: %v", level, result)
			resp, err := s.AutonomousResponse(incident)
			if err == nil {
				decision["response"] = resp
				decision["actions_taken"] = append(decision["actions_taken"].([]string), "autonomous_response_triggered")
			}
		}
	}

	log.Printf("Agent: Cycle complete - %d actions taken", len(decision["actions_taken"].([]string)))

	if onDecision != nil {
		onDecision("autonomous_cycle", decision)
	}
}

func (s *AgentSystem) EnableAutonomousMode(enabled bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	log.Printf("Agent: Autonomous mode %s", map[bool]string{true: "enabled", false: "disabled"}[enabled])
}

func (s *AgentSystem) GetAutonomousStatus() map[string]interface{} {
	agents := s.GetAgents()
	workflows := s.GetWorkflows()

	autonomousAgent := s.GetAgent("autonomous")
	var hasActiveTasks int
	if autonomousAgent != nil {
		autonomousAgent.mu.RLock()
		for _, task := range autonomousAgent.Tasks {
			if task.Status == "running" {
				hasActiveTasks++
			}
		}
		autonomousAgent.mu.RUnlock()
	}

	return map[string]interface{}{
		"enabled":         true,
		"agents_count":    len(agents),
		"workflows_count": len(workflows),
		"active_tasks":    hasActiveTasks,
		"last_cycle":      time.Now().Format(time.RFC3339),
		"llm_available":   s.llm != nil,
		"file_manager":    s.fileManager != nil,
	}
}
