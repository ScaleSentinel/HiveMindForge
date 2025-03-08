package agents

import (
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/streadway/amqp"
)

// TaskManager gerencia a distribuição e monitoramento de tarefas
type TaskManager struct {
	sync.RWMutex
	tasks        map[string]*Task
	healthStatus map[string]*AgentHealth
	rabbitmqConn *amqp.Connection
	rabbitmqCh   *amqp.Channel
}

// NewTaskManager cria uma nova instância do gerenciador de tarefas
func NewTaskManager() (*TaskManager, error) {
	conn, err := amqp.Dial(fmt.Sprintf("amqp://guest:guest@%s:%d/", RABBITMQ_HOST, RABBITMQ_PORT))
	if err != nil {
		return nil, fmt.Errorf("falha ao conectar ao RabbitMQ: %v", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("falha ao criar canal: %v", err)
	}

	// Declarar exchanges
	err = ch.ExchangeDeclare(
		EXCHANGE_HEALTH, // nome
		"topic",         // tipo
		QUEUE_DURABLE,   // durable
		false,           // auto-deleted
		false,           // internal
		false,           // no-wait
		nil,             // arguments
	)
	if err != nil {
		ch.Close()
		conn.Close()
		return nil, fmt.Errorf("falha ao declarar exchange de saúde: %v", err)
	}

	err = ch.ExchangeDeclare(
		EXCHANGE_TASK, // nome
		"topic",       // tipo
		QUEUE_DURABLE, // durable
		false,         // auto-deleted
		false,         // internal
		false,         // no-wait
		nil,           // arguments
	)
	if err != nil {
		ch.Close()
		conn.Close()
		return nil, fmt.Errorf("falha ao declarar exchange de tarefas: %v", err)
	}

	tm := &TaskManager{
		tasks:        make(map[string]*Task),
		healthStatus: make(map[string]*AgentHealth),
		rabbitmqConn: conn,
		rabbitmqCh:   ch,
	}

	// Iniciar monitoramento de saúde
	go tm.monitorHealthEvents()

	return tm, nil
}

// AddTask adiciona uma nova tarefa à fila
func (tm *TaskManager) AddTask(task *Task) {
	tm.Lock()
	defer tm.Unlock()

	task.CreatedAt = time.Now()
	task.Status = TaskStatusPending
	tm.tasks[task.ID] = task

	// Publicar evento de nova tarefa
	body, err := json.Marshal(task)
	if err != nil {
		log.Printf("Erro ao converter tarefa para JSON: %v", err)
		return
	}

	err = tm.rabbitmqCh.Publish(
		EXCHANGE_TASK,                     // exchange
		fmt.Sprintf("task.%s", task.Type), // routing key
		false,                             // mandatory
		false,                             // immediate
		amqp.Publishing{
			ContentType:  "application/json",
			Body:         body,
			DeliveryMode: MESSAGE_PERSISTENT, // persistente
		})

	if err != nil {
		log.Printf("Erro ao publicar tarefa: %v", err)
		return
	}

	log.Printf("📋 Nova tarefa adicionada: %s (Tipo: %s, Prioridade: %d)",
		task.ID, task.Type, task.Priority)
}

// monitorHealthEvents monitora os eventos de saúde dos agentes
func (tm *TaskManager) monitorHealthEvents() {
	q, err := tm.rabbitmqCh.QueueDeclare(
		QUEUE_HEALTH_MONITOR, // nome
		QUEUE_DURABLE,        // durable
		QUEUE_AUTO_DELETE,    // delete when unused
		QUEUE_EXCLUSIVE,      // exclusive
		QUEUE_NO_WAIT,        // no-wait
		nil,                  // arguments
	)
	if err != nil {
		log.Printf("Erro ao declarar fila de saúde: %v", err)
		return
	}

	err = tm.rabbitmqCh.QueueBind(
		q.Name,          // queue name
		"health.*",      // routing key
		EXCHANGE_HEALTH, // exchange
		false,
		nil,
	)
	if err != nil {
		log.Printf("Erro ao fazer binding da fila de saúde: %v", err)
		return
	}

	msgs, err := tm.rabbitmqCh.Consume(
		q.Name, // queue
		"",     // consumer
		true,   // auto-ack
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
	)
	if err != nil {
		log.Printf("Erro ao consumir eventos de saúde: %v", err)
		return
	}

	for msg := range msgs {
		var health AgentHealth
		if err := json.Unmarshal(msg.Body, &health); err != nil {
			log.Printf("Erro ao decodificar evento de saúde: %v", err)
			continue
		}

		tm.Lock()
		tm.healthStatus[health.AgentName] = &health
		tm.Unlock()

		log.Printf("❤️ Heartbeat recebido do agente %s (Processando: %v)",
			health.AgentName, health.IsProcessing)
	}
}

// EmitHealthSignal emite um sinal de saúde para um agente
func (tm *TaskManager) EmitHealthSignal(health *AgentHealth) error {
	body, err := json.Marshal(health)
	if err != nil {
		return fmt.Errorf("erro ao codificar evento de saúde: %v", err)
	}

	err = tm.rabbitmqCh.Publish(
		EXCHANGE_HEALTH, // exchange
		fmt.Sprintf("health.%s", health.AgentName), // routing key
		false, // mandatory
		false, // immediate
		amqp.Publishing{
			ContentType:  "application/json",
			Body:         body,
			DeliveryMode: MESSAGE_PERSISTENT, // persistente
		})

	if err != nil {
		return fmt.Errorf("erro ao publicar evento de saúde: %v", err)
	}

	return nil
}

// GetNextTask retorna a próxima tarefa mais adequada para um agente
func (tm *TaskManager) GetNextTask(agentName string) *Task {
	tm.Lock()
	defer tm.Unlock()

	// Verificar saúde do agente
	health, exists := tm.healthStatus[agentName]
	if !exists || time.Since(health.LastHeartbeat) > 30*time.Second {
		log.Printf("⚠️ Agente %s não está saudável ou não registrado", agentName)
		return nil
	}

	// Se o agente já está processando, não atribuir nova tarefa
	if health.IsProcessing {
		return nil
	}

	var bestTask *Task
	highestPriority := PriorityLow - 1

	// Encontrar a tarefa pendente com maior prioridade
	for _, task := range tm.tasks {
		if task.Status != TaskStatusPending {
			continue
		}

		if task.Type == health.AgentName && task.Priority > highestPriority {
			bestTask = task
			highestPriority = task.Priority
		}
	}

	if bestTask != nil {
		bestTask.Status = TaskStatusAssigned
		bestTask.AssignedTo = agentName
		log.Printf("✅ Tarefa %s atribuída ao agente %s", bestTask.ID, agentName)
	}

	return bestTask
}

// UpdateTaskStatus atualiza o status de uma tarefa
func (tm *TaskManager) UpdateTaskStatus(taskID string, status TaskStatus) {
	tm.Lock()
	defer tm.Unlock()

	if task, exists := tm.tasks[taskID]; exists {
		task.Status = status
		if status == TaskStatusComplete {
			now := time.Now()
			task.CompletedAt = &now
		}
		log.Printf("🔄 Status da tarefa %s atualizado para: %s", taskID, status)
	}
}

// GetAgentHealth retorna o estado de saúde de um agente
func (tm *TaskManager) GetAgentHealth(agentName string) *AgentHealth {
	tm.RLock()
	defer tm.RUnlock()
	return tm.healthStatus[agentName]
}
