package agents

import "time"

// Task representa uma tarefa a ser executada por um agente
type Task struct {
	ID             string       `json:"id"`
	Type           string       `json:"type"` // quiz, challenge, etc.
	Priority       TaskPriority `json:"priority"`
	Status         TaskStatus   `json:"status"`
	Data           interface{}  `json:"data"`
	ExpectedOutput string       `json:"expected_output"`
	FormatOutput   string       `json:"format_output"`
	CreatedAt      time.Time    `json:"created_at"`
	AssignedTo     string       `json:"assigned_to,omitempty"`
	CompletedAt    *time.Time   `json:"completed_at,omitempty"`
}

// Agent representa um agente base com características comuns
type Agent struct {
	Name            string `json:"name"`
	Role            string `json:"role"`
	Goal            string `json:"goal"`
	AllowDelegation bool   `json:"allow_delegation"`
	Model           string `json:"model"`
	Backstory       string `json:"backstory"`
}

// TaskPriority define a prioridade da tarefa
type TaskPriority int

const (
	PriorityLow    TaskPriority = 1
	PriorityNormal TaskPriority = 2
	PriorityHigh   TaskPriority = 3
)

// TaskStatus representa o estado atual da tarefa
type TaskStatus string

const (
	TaskStatusPending  TaskStatus = "pending"
	TaskStatusAssigned TaskStatus = "assigned"
	TaskStatusRunning  TaskStatus = "running"
	TaskStatusComplete TaskStatus = "complete"
	TaskStatusFailed   TaskStatus = "failed"
)

// AgentHealth representa o estado de saúde de um agente
type AgentHealth struct {
	AgentName      string    `json:"agent_name"`
	LastHeartbeat  time.Time `json:"last_heartbeat"`
	IsProcessing   bool      `json:"is_processing"`
	CurrentTaskID  string    `json:"current_task_id,omitempty"`
	ProcessingTime float64   `json:"processing_time"` // tempo médio de processamento em segundos
	SuccessRate    float64   `json:"success_rate"`    // taxa de sucesso (0-1)
	LastError      string    `json:"last_error,omitempty"`
	LastErrorTime  time.Time `json:"last_error_time,omitempty"`
}
