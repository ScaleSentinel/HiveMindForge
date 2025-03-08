package agents

import (
	"log"
	"sync"
	"time"
)

// OrchestratorInfrastructureAgent gerencia a infraestrutura e escalabilidade do sistema
type OrchestratorInfrastructureAgent struct {
	Agent
	taskManager   *TaskManager
	agents        map[string]*CognitiveAgent
	agentLock     sync.RWMutex
	lastScaleTime time.Time
	observerAgent *ObserverInfrastructureAgent
}

// NewOrchestratorInfrastructureAgent cria uma nova instância do agente orquestrador de infraestrutura
func NewOrchestratorInfrastructureAgent() (*OrchestratorInfrastructureAgent, error) {
	taskManager, err := NewTaskManager()
	if err != nil {
		return nil, err
	}

	observer := NewObserverInfrastructureAgent()

	agent := &OrchestratorInfrastructureAgent{
		Agent: Agent{
			Name:            "Infrastructure Orchestrator",
			Role:            "Gerenciador de Infraestrutura",
			Goal:            "Gerenciar e escalar a infraestrutura do sistema",
			AllowDelegation: true,
			Model:           "gpt-4",
			Backstory:       "Um agente especializado em gerenciar e otimizar a infraestrutura do sistema",
		},
		taskManager:   taskManager,
		agents:        make(map[string]*CognitiveAgent),
		lastScaleTime: time.Now(),
		observerAgent: observer,
	}

	return agent, nil
}

// RegisterAgent registra um novo agente cognitivo
func (o *OrchestratorInfrastructureAgent) RegisterAgent(agent *CognitiveAgent) {
	o.agentLock.Lock()
	defer o.agentLock.Unlock()

	o.agents[agent.Name] = agent
	log.Printf("✨ Agente registrado: %s", agent.Name)
}

// GetTaskManager retorna o gerenciador de tarefas
func (o *OrchestratorInfrastructureAgent) GetTaskManager() *TaskManager {
	return o.taskManager
}

// AddTask adiciona uma nova tarefa ao sistema
func (o *OrchestratorInfrastructureAgent) AddTask(task *Task) {
	o.taskManager.AddTask(task)
}

// ScaleSystem avalia e ajusta a escala do sistema
func (o *OrchestratorInfrastructureAgent) ScaleSystem() {
	// Verificar cooldown de escala
	if time.Since(o.lastScaleTime) < SCALE_COOLDOWN {
		return
	}

	metrics := o.observerAgent.GetSystemMetrics()

	// Verificar condições para escala
	needsScaling := false

	if metrics.CPUUsage > CPU_THRESHOLD {
		log.Printf("⚠️ Alto uso de CPU: %.2f%%", metrics.CPUUsage)
		needsScaling = true
	}

	if metrics.MemoryUsage > MEMORY_THRESHOLD {
		log.Printf("⚠️ Alto uso de memória: %.2f%%", metrics.MemoryUsage)
		needsScaling = true
	}

	if metrics.TasksPerAgent > TASKS_THRESHOLD {
		log.Printf("⚠️ Muitas tarefas por agente: %d", metrics.TasksPerAgent)
		needsScaling = true
	}

	if metrics.ErrorCount > ERROR_THRESHOLD {
		log.Printf("⚠️ Alto número de erros: %d", metrics.ErrorCount)
		needsScaling = true
	}

	if needsScaling {
		o.scaleOut()
		o.lastScaleTime = time.Now()
	}
}

// scaleOut aumenta a capacidade do sistema
func (o *OrchestratorInfrastructureAgent) scaleOut() {
	o.agentLock.RLock()
	defer o.agentLock.RUnlock()

	// Identificar agentes mais sobrecarregados
	for _, agent := range o.agents {
		health := o.taskManager.GetAgentHealth(agent.Name)
		if health == nil {
			continue
		}

		// Se o agente está sobrecarregado, criar um clone
		if health.IsProcessing && health.ProcessingTime > 5.0 { // mais de 5 segundos por tarefa
			newAgent := agent.Clone()
			o.RegisterAgent(newAgent)
			log.Printf("🔄 Novo agente criado: %s", newAgent.Name)
		}
	}
}

// Stop interrompe o agente orquestrador
func (o *OrchestratorInfrastructureAgent) Stop() {
	o.agentLock.Lock()
	defer o.agentLock.Unlock()

	for _, agent := range o.agents {
		agent.Stop()
	}
}
