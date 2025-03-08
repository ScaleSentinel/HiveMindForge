package agents

import "time"

// Configurações de escalabilidade
const (
	// Limites para escalabilidade
	CPU_THRESHOLD    = 80.0 // Porcentagem de uso de CPU
	MEMORY_THRESHOLD = 85.0 // Porcentagem de uso de memória
	TASKS_THRESHOLD  = 10   // Número máximo de tarefas por agente
	ERROR_THRESHOLD  = 5    // Número máximo de erros antes de escalar

	// Tempo de espera entre operações de escala
	SCALE_COOLDOWN = 5 * time.Minute
)
