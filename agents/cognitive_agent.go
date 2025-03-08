package agents

import (
	"context"
	"fmt"
	"log"
	"runtime"
	"sync"
	"time"

	"hivemindforge/agents/telemetry"
)

// CognitiveAgent representa um agente cognitivo que pode executar tarefas específicas
type CognitiveAgent struct {
	Agent
	Type           string
	taskManager    *TaskManager
	currentTask    *Task
	taskLock       sync.RWMutex
	stopChan       chan struct{}
	healthTicker   *time.Ticker
	metricsTicker  *time.Ticker
	processingTime float64
	successCount   int
	totalTasks     int
	ctx            context.Context
}

// NewQuizAgent cria um novo agente especializado em criar quizzes
func NewQuizAgent(taskManager *TaskManager) *CognitiveAgent {
	agent := &CognitiveAgent{
		Agent: Agent{
			Name:            "Quiz Agent",
			Role:            "Especialista em Quizzes",
			Goal:            "Criar e avaliar quizzes educacionais",
			AllowDelegation: false,
			Model:           "gpt-4-mini",
			Backstory:       "Um especialista em educação focado em criar quizzes envolventes e educativos",
		},
		Type:          "quiz",
		taskManager:   taskManager,
		stopChan:      make(chan struct{}),
		healthTicker:  time.NewTicker(10 * time.Second),
		metricsTicker: time.NewTicker(3 * time.Second),
		ctx:           context.Background(),
	}

	go agent.startHealthMonitoring()
	go agent.startMetricsCollection()
	go agent.processTasksLoop()

	return agent
}

// NewChallengeAgent cria um novo agente especializado em criar desafios
func NewChallengeAgent(taskManager *TaskManager) *CognitiveAgent {
	agent := &CognitiveAgent{
		Agent: Agent{
			Name:            "Challenge Agent",
			Role:            "Especialista em Desafios",
			Goal:            "Criar e avaliar desafios práticos",
			AllowDelegation: false,
			Model:           "gpt-4-mini",
			Backstory:       "Um especialista em criar desafios práticos e envolventes para aprendizado",
		},
		Type:          "challenge",
		taskManager:   taskManager,
		stopChan:      make(chan struct{}),
		healthTicker:  time.NewTicker(10 * time.Second),
		metricsTicker: time.NewTicker(3 * time.Second),
		ctx:           context.Background(),
	}

	go agent.startHealthMonitoring()
	go agent.startMetricsCollection()
	go agent.processTasksLoop()

	return agent
}

// Clone cria uma nova instância do agente com um nome único
func (c *CognitiveAgent) Clone() *CognitiveAgent {
	timestamp := time.Now().Unix()
	newName := fmt.Sprintf("%s-%d", c.Name, timestamp)

	agent := &CognitiveAgent{
		Agent: Agent{
			Name:            newName,
			Role:            c.Role,
			Goal:            c.Goal,
			AllowDelegation: c.AllowDelegation,
			Model:           c.Model,
			Backstory:       c.Backstory,
		},
		Type:          c.Type,
		taskManager:   c.taskManager,
		stopChan:      make(chan struct{}),
		healthTicker:  time.NewTicker(10 * time.Second),
		metricsTicker: time.NewTicker(3 * time.Second),
		ctx:           context.Background(),
	}

	go agent.startHealthMonitoring()
	go agent.startMetricsCollection()
	go agent.processTasksLoop()

	return agent
}

// Stop interrompe o processamento de tarefas e monitoramento
func (c *CognitiveAgent) Stop() {
	close(c.stopChan)
	c.healthTicker.Stop()
	c.metricsTicker.Stop()
}

// startHealthMonitoring inicia o monitoramento de saúde do agente
func (c *CognitiveAgent) startHealthMonitoring() {
	for {
		select {
		case <-c.healthTicker.C:
			c.emitHealthSignal()
		case <-c.stopChan:
			return
		}
	}
}

// startMetricsCollection inicia a coleta de métricas do agente
func (c *CognitiveAgent) startMetricsCollection() {
	for {
		select {
		case <-c.metricsTicker.C:
			c.collectMetrics()
		case <-c.stopChan:
			return
		}
	}
}

// collectMetrics coleta métricas do agente
func (c *CognitiveAgent) collectMetrics() {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	// Calcular uso de CPU (exemplo simplificado)
	cpuUsage := float64(runtime.NumGoroutine()) / float64(runtime.NumCPU()) * 100

	// Registrar métricas no OpenTelemetry
	telemetry.RecordResourceUsage(c.ctx, c.Name, float64(m.Alloc), cpuUsage)
}

// emitHealthSignal emite um sinal de saúde do agente
func (c *CognitiveAgent) emitHealthSignal() {
	c.taskLock.RLock()
	currentTaskID := ""
	if c.currentTask != nil {
		currentTaskID = c.currentTask.ID
	}
	c.taskLock.RUnlock()

	health := &AgentHealth{
		AgentName:      c.Name,
		LastHeartbeat:  time.Now(),
		IsProcessing:   c.currentTask != nil,
		CurrentTaskID:  currentTaskID,
		ProcessingTime: c.processingTime,
		SuccessRate:    float64(c.successCount) / float64(c.totalTasks),
	}

	if err := c.taskManager.EmitHealthSignal(health); err != nil {
		log.Printf("❌ Erro ao emitir sinal de saúde: %v", err)
	}
}

// processTasksLoop processa tarefas continuamente
func (c *CognitiveAgent) processTasksLoop() {
	for {
		select {
		case <-c.stopChan:
			return
		default:
			c.processNextTask()
			time.Sleep(1 * time.Second) // Evitar consumo excessivo de CPU
		}
	}
}

// processNextTask processa a próxima tarefa disponível
func (c *CognitiveAgent) processNextTask() {
	task := c.taskManager.GetNextTask(c.Name)
	if task == nil {
		return
	}

	c.taskLock.Lock()
	c.currentTask = task
	c.taskLock.Unlock()

	// Registrar início da tarefa
	telemetry.RecordTaskStart(c.ctx, c.Name, task.ID, task.Type)

	startTime := time.Now()
	c.taskManager.UpdateTaskStatus(task.ID, TaskStatusRunning)

	// Processar a tarefa
	success := c.executeTask(task)

	// Registrar conclusão da tarefa
	duration := time.Since(startTime)
	telemetry.RecordTaskCompletion(c.ctx, c.Name, task.ID, task.Type, duration, success)

	// Atualizar métricas
	c.totalTasks++
	if success {
		c.successCount++
		c.taskManager.UpdateTaskStatus(task.ID, TaskStatusComplete)
	} else {
		c.taskManager.UpdateTaskStatus(task.ID, TaskStatusFailed)
	}

	// Atualizar tempo médio de processamento
	if c.processingTime == 0 {
		c.processingTime = duration.Seconds()
	} else {
		c.processingTime = (c.processingTime + duration.Seconds()) / 2
	}

	c.taskLock.Lock()
	c.currentTask = nil
	c.taskLock.Unlock()
}

// executeTask executa uma tarefa específica
func (c *CognitiveAgent) executeTask(task *Task) bool {
	// Implementar lógica específica de execução da tarefa
	// Por enquanto, apenas simular processamento
	time.Sleep(2 * time.Second)
	return true
}
