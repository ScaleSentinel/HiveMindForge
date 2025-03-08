package agents

// SystemMetrics armazena métricas do sistema
type SystemMetrics struct {
	CPUUsage      float64 // Porcentagem de uso de CPU
	MemoryUsage   float64 // Porcentagem de uso de memória
	TasksPerAgent int     // Número médio de tarefas por agente
	ErrorCount    int     // Número de erros no período
	LastUpdate    int64   // Timestamp da última atualização
}

// AgentMetrics armazena métricas específicas de um agente
type AgentMetrics struct {
	AgentName    string  // Nome do agente
	CPU          float64 // Uso de CPU em porcentagem
	Memory       uint64  // Uso de memória em bytes
	TasksInQueue int     // Número de tarefas na fila
	ErrorRate    float64 // Taxa de erros (0-1)
	LastUpdate   int64   // Timestamp da última atualização
}
