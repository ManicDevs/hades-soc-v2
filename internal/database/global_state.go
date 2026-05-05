package database

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"
)

type GlobalStateRepository struct {
	db *sql.DB
}

func NewGlobalStateRepository(db *sql.DB) *GlobalStateRepository {
	return &GlobalStateRepository{db: db}
}

func (r *GlobalStateRepository) Create(state *GlobalState) error {
	metadataJSON, err := json.Marshal(state.Metadata)
	if err != nil {
		metadataJSON = []byte("{}")
	}

	query := `
		INSERT INTO global_states (
			task_id, task_type, status, target, target_type, agent_id,
			module_name, policy_id, workflow_id, severity, error_message,
			result_summary, metadata, started_at, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16)
		RETURNING id
	`

	now := time.Now()
	state.CreatedAt = now
	state.UpdatedAt = now

	err = r.db.QueryRow(
		query,
		state.TaskID, state.TaskType, state.Status, state.Target, state.TargetType,
		state.AgentID, state.ModuleName, state.PolicyID, state.WorkflowID,
		state.Severity, state.ErrorMessage, state.ResultSummary, metadataJSON,
		state.StartedAt, state.CreatedAt, state.UpdatedAt,
	).Scan(&state.ID)

	if err != nil {
		return fmt.Errorf("failed to create global state: %w", err)
	}

	return nil
}

func (r *GlobalStateRepository) Update(state *GlobalState) error {
	metadataJSON, err := json.Marshal(state.Metadata)
	if err != nil {
		metadataJSON = []byte("{}")
	}

	query := `
		UPDATE global_states SET
			status = $1, agent_id = $2, error_message = $3,
			result_summary = $4, metadata = $5, completed_at = $6, updated_at = $7
		WHERE id = $8
	`

	state.UpdatedAt = time.Now()

	_, err = r.db.Exec(
		query,
		state.Status, state.AgentID, state.ErrorMessage,
		state.ResultSummary, metadataJSON, state.CompletedAt, state.UpdatedAt, state.ID,
	)

	if err != nil {
		return fmt.Errorf("failed to update global state: %w", err)
	}

	return nil
}

func (r *GlobalStateRepository) GetByID(id int) (*GlobalState, error) {
	query := `
		SELECT id, task_id, task_type, status, target, target_type, agent_id,
			module_name, policy_id, workflow_id, severity, error_message,
			result_summary, metadata, started_at, completed_at, created_at, updated_at
		FROM global_states WHERE id = $1
	`

	state := &GlobalState{}
	var metadataJSON []byte

	err := r.db.QueryRow(query, id).Scan(
		&state.ID, &state.TaskID, &state.TaskType, &state.Status, &state.Target,
		&state.TargetType, &state.AgentID, &state.ModuleName, &state.PolicyID,
		&state.WorkflowID, &state.Severity, &state.ErrorMessage, &state.ResultSummary,
		&metadataJSON, &state.StartedAt, &state.CompletedAt, &state.CreatedAt, &state.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get global state: %w", err)
	}

	json.Unmarshal(metadataJSON, &state.Metadata)
	return state, nil
}

func (r *GlobalStateRepository) GetByTaskID(taskID string) (*GlobalState, error) {
	query := `
		SELECT id, task_id, task_type, status, target, target_type, agent_id,
			module_name, policy_id, workflow_id, severity, error_message,
			result_summary, metadata, started_at, completed_at, created_at, updated_at
		FROM global_states WHERE task_id = $1
	`

	state := &GlobalState{}
	var metadataJSON []byte

	err := r.db.QueryRow(query, taskID).Scan(
		&state.ID, &state.TaskID, &state.TaskType, &state.Status, &state.Target,
		&state.TargetType, &state.AgentID, &state.ModuleName, &state.PolicyID,
		&state.WorkflowID, &state.Severity, &state.ErrorMessage, &state.ResultSummary,
		&metadataJSON, &state.StartedAt, &state.CompletedAt, &state.CreatedAt, &state.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get global state by task_id: %w", err)
	}

	json.Unmarshal(metadataJSON, &state.Metadata)
	return state, nil
}

func (r *GlobalStateRepository) FindRunningByTarget(taskType TaskType, target string) (*GlobalState, error) {
	query := `
		SELECT id, task_id, task_type, status, target, target_type, agent_id,
			module_name, policy_id, workflow_id, severity, error_message,
			result_summary, metadata, started_at, completed_at, created_at, updated_at
		FROM global_states
		WHERE task_type = $1 AND target = $2 AND status = 'running'
		ORDER BY created_at DESC
		LIMIT 1
	`

	state := &GlobalState{}
	var metadataJSON []byte

	err := r.db.QueryRow(query, taskType, target).Scan(
		&state.ID, &state.TaskID, &state.TaskType, &state.Status, &state.Target,
		&state.TargetType, &state.AgentID, &state.ModuleName, &state.PolicyID,
		&state.WorkflowID, &state.Severity, &state.ErrorMessage, &state.ResultSummary,
		&metadataJSON, &state.StartedAt, &state.CompletedAt, &state.CreatedAt, &state.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to find running state by target: %w", err)
	}

	json.Unmarshal(metadataJSON, &state.Metadata)
	return state, nil
}

func (r *GlobalStateRepository) FindRunningByModule(taskType TaskType, moduleName string) (*GlobalState, error) {
	query := `
		SELECT id, task_id, task_type, status, target, target_type, agent_id,
			module_name, policy_id, workflow_id, severity, error_message,
			result_summary, metadata, started_at, completed_at, created_at, updated_at
		FROM global_states
		WHERE task_type = $1 AND module_name = $2 AND status = 'running'
		ORDER BY created_at DESC
		LIMIT 1
	`

	state := &GlobalState{}
	var metadataJSON []byte

	err := r.db.QueryRow(query, taskType, moduleName).Scan(
		&state.ID, &state.TaskID, &state.TaskType, &state.Status, &state.Target,
		&state.TargetType, &state.AgentID, &state.ModuleName, &state.PolicyID,
		&state.WorkflowID, &state.Severity, &state.ErrorMessage, &state.ResultSummary,
		&metadataJSON, &state.StartedAt, &state.CompletedAt, &state.CreatedAt, &state.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to find running state by module: %w", err)
	}

	json.Unmarshal(metadataJSON, &state.Metadata)
	return state, nil
}

func (r *GlobalStateRepository) ListByStatus(taskType TaskType, status TaskStatus) ([]*GlobalState, error) {
	query := `
		SELECT id, task_id, task_type, status, target, target_type, agent_id,
			module_name, policy_id, workflow_id, severity, error_message,
			result_summary, metadata, started_at, completed_at, created_at, updated_at
		FROM global_states
		WHERE task_type = $1 AND status = $2
		ORDER BY created_at DESC
	`

	rows, err := r.db.Query(query, taskType, status)
	if err != nil {
		return nil, fmt.Errorf("failed to list global states by status: %w", err)
	}
	defer rows.Close()

	var states []*GlobalState
	for rows.Next() {
		state := &GlobalState{}
		var metadataJSON []byte

		err := rows.Scan(
			&state.ID, &state.TaskID, &state.TaskType, &state.Status, &state.Target,
			&state.TargetType, &state.AgentID, &state.ModuleName, &state.PolicyID,
			&state.WorkflowID, &state.Severity, &state.ErrorMessage, &state.ResultSummary,
			&metadataJSON, &state.StartedAt, &state.CompletedAt, &state.CreatedAt, &state.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan global state: %w", err)
		}

		json.Unmarshal(metadataJSON, &state.Metadata)
		states = append(states, state)
	}

	return states, nil
}

func (r *GlobalStateRepository) Delete(id int) error {
	query := `DELETE FROM global_states WHERE id = $1`
	_, err := r.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to delete global state: %w", err)
	}
	return nil
}

func (r *GlobalStateRepository) DeleteOlderThan(olderThan time.Duration) (int64, error) {
	query := `DELETE FROM global_states WHERE created_at < $1`
	result, err := r.db.Exec(query, time.Now().Add(-olderThan))
	if err != nil {
		return 0, fmt.Errorf("failed to delete old global states: %w", err)
	}
	return result.RowsAffected()
}

func (r *GlobalStateRepository) IsTaskRunning(taskType TaskType, target string) (bool, *GlobalState, error) {
	state, err := r.FindRunningByTarget(taskType, target)
	if err != nil {
		return false, nil, err
	}
	if state != nil {
		return true, state, nil
	}
	return false, nil, nil
}

func (r *GlobalStateRepository) IsModuleRunning(taskType TaskType, moduleName string) (bool, *GlobalState, error) {
	state, err := r.FindRunningByModule(taskType, moduleName)
	if err != nil {
		return false, nil, err
	}
	if state != nil {
		return true, state, nil
	}
	return false, nil, nil
}

// FindByTimeRange queries GlobalState entries within a time range
func (r *GlobalStateRepository) FindByTimeRange(start, end time.Time) ([]*GlobalState, error) {
	query := `
		SELECT id, task_id, task_type, status, target, target_type, agent_id,
			module_name, policy_id, workflow_id, severity, error_message,
			result_summary, metadata, started_at, completed_at, created_at, updated_at
		FROM global_states
		WHERE created_at BETWEEN $1 AND $2
		ORDER BY created_at DESC
	`

	rows, err := r.db.Query(query, start, end)
	if err != nil {
		return nil, fmt.Errorf("failed to query global states by time range: %w", err)
	}
	defer rows.Close()

	return r.scanGlobalStates(rows)
}

// FindByTypeAndStatus queries GlobalState entries by task type and status
func (r *GlobalStateRepository) FindByTypeAndStatus(taskType TaskType, status TaskStatus) ([]*GlobalState, error) {
	query := `
		SELECT id, task_id, task_type, status, target, target_type, agent_id,
			module_name, policy_id, workflow_id, severity, error_message,
			result_summary, metadata, started_at, completed_at, created_at, updated_at
		FROM global_states
		WHERE task_type = $1 AND status = $2
		ORDER BY created_at DESC
	`

	rows, err := r.db.Query(query, taskType, status)
	if err != nil {
		return nil, fmt.Errorf("failed to query global states by type and status: %w", err)
	}
	defer rows.Close()

	return r.scanGlobalStates(rows)
}

// scanGlobalStates scans SQL rows into GlobalState structs
func (r *GlobalStateRepository) scanGlobalStates(rows *sql.Rows) ([]*GlobalState, error) {
	var states []*GlobalState

	for rows.Next() {
		state := &GlobalState{}
		var metadataJSON []byte
		var completedAt sql.NullTime

		err := rows.Scan(
			&state.ID, &state.TaskID, &state.TaskType, &state.Status, &state.Target, &state.TargetType,
			&state.AgentID, &state.ModuleName, &state.PolicyID, &state.WorkflowID,
			&state.Severity, &state.ErrorMessage, &state.ResultSummary,
			&metadataJSON, &state.StartedAt, &completedAt, &state.CreatedAt, &state.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan global state: %w", err)
		}

		if completedAt.Valid {
			state.CompletedAt = &completedAt.Time
		}

		if len(metadataJSON) > 0 {
			json.Unmarshal(metadataJSON, &state.Metadata)
		}

		states = append(states, state)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating global states: %w", err)
	}

	return states, nil
}
