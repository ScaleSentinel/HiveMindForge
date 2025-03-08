package agents

// Configurações do RabbitMQ
const (
	RABBITMQ_HOST = "localhost"
	RABBITMQ_PORT = 5672

	// Exchanges
	EXCHANGE_HEALTH = "health_events"
	EXCHANGE_TASK   = "task_events"

	// Filas
	QUEUE_HEALTH_MONITOR = "health_monitor"
	QUEUE_TASK_QUIZ      = "quiz_tasks"
	QUEUE_TASK_CHALLENGE = "challenge_tasks"

	// Configurações de durabilidade
	QUEUE_DURABLE     = true
	QUEUE_AUTO_DELETE = false
	QUEUE_EXCLUSIVE   = false
	QUEUE_NO_WAIT     = false

	// Configurações de mensagem
	MESSAGE_PERSISTENT = 2 // DeliveryMode 2 = persistente
)

// Configurações de escalabilidade
const (
	CPU_THRESHOLD    = 80.0 // 80% de uso de CPU
	MEMORY_THRESHOLD = 85.0 // 85% de uso de memória
	TASKS_THRESHOLD  = 100  // 100 tarefas na fila
	ERROR_THRESHOLD  = 0.05 // 5% de taxa de erro
	SCALE_COOLDOWN   = 300  // 5 minutos de cooldown entre escalas
)
