package agents

// Configurações de escalabilidade
const (
	CPU_THRESHOLD    = 80.0 // 80% de uso de CPU
	MEMORY_THRESHOLD = 85.0 // 85% de uso de memória
	TASKS_THRESHOLD  = 100  // 100 tarefas na fila
	ERROR_THRESHOLD  = 0.05 // 5% de taxa de erro
	SCALE_COOLDOWN   = 300  // 5 minutos de cooldown entre escalas
)
