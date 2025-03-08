package agents

import (
	"context"
	"fmt"
	"time"

	"HiveMindForge/agents/memory"
)

// CognitiveAgent representa um agente cognitivo que pode executar tarefas específicas
type CognitiveAgent struct {
	*BaseAgent
	Model            string                 // Modelo de IA usado pelo agente
	Temperature      float64                // Temperatura para geração de respostas
	MaxTokens        int                    // Número máximo de tokens por resposta
	ContextWindow    int                    // Tamanho da janela de contexto
	KnowledgeBase    map[string]interface{} // Base de conhecimento do agente
	LearningRate     float64                // Taxa de aprendizado para ajustes
	PromptTemplates  map[string]string      // Templates de prompts
	ResponseHistory  []string               // Histórico de respostas
	PerformanceStats map[string]float64     // Estatísticas de performance
	Role             string                 // Papel do agente no sistema
	Goal             string                 // Objetivo principal do agente
	AllowDelegation  bool                   // Se o agente pode delegar tarefas
	Backstory        string                 // História/contexto do agente

	// Campos específicos para execução de tarefas
	taskManager   *TaskManager
	memoryManager memory.MemoryManager
	stopChan      chan struct{}
	healthTicker  *time.Ticker
	metricsTicker *time.Ticker
	ctx           context.Context
}

// NewCognitiveAgent cria uma nova instância de CognitiveAgent
func NewCognitiveAgent(id, name, description string, maxRounds int, model string, role string, goal string, memoryManager memory.MemoryManager) *CognitiveAgent {
	return &CognitiveAgent{
		BaseAgent:       NewBaseAgent(id, name, description, maxRounds),
		Model:           model,
		Temperature:     0.7,
		MaxTokens:       2048,
		ContextWindow:   4096,
		KnowledgeBase:   make(map[string]interface{}),
		LearningRate:    0.001,
		PromptTemplates: make(map[string]string),
		ResponseHistory: make([]string, 0),
		Role:            role,
		Goal:            goal,
		AllowDelegation: true,
		PerformanceStats: map[string]float64{
			"accuracy":       0.8, // Inicializa com 80% de acurácia
			"response_time":  0.0,
			"success_rate":   0.8, // Inicializa com 80% de taxa de sucesso
			"token_usage":    0.0,
			"context_hits":   0.0,
			"learning_score": 0.0,
		},
		memoryManager: memoryManager,
		stopChan:      make(chan struct{}),
	}
}

// Train implementa o treinamento específico para o agente cognitivo
func (a *CognitiveAgent) Train(ctx context.Context, config TrainingConfig) (*TrainingMetrics, error) {
	// Primeiro executa o treinamento base
	metrics, err := a.BaseAgent.Train(ctx, config)
	if err != nil {
		return metrics, err
	}

	// Ajusta parâmetros baseado no histórico
	if config.UseHistorical && len(a.ResponseHistory) > 0 {
		a.adjustParameters()
	}

	// Atualiza métricas com dados específicos do agente cognitivo
	metrics.Accuracy = a.PerformanceStats["accuracy"]
	metrics.Loss = 1.0 - a.PerformanceStats["success_rate"]

	// Simula processo de aprendizado
	time.Sleep(500 * time.Millisecond)

	// Atualiza estatísticas
	a.updatePerformanceStats(metrics)

	// Armazena métricas na memória de longo prazo
	memory := &memory.Memory{
		ID:      fmt.Sprintf("training_%s_%d", a.ID, time.Now().Unix()),
		AgentID: a.ID,
		Type:    memory.LongTerm,
		Content: map[string]interface{}{
			"metrics": metrics,
			"parameters": map[string]interface{}{
				"temperature":   a.Temperature,
				"learning_rate": a.LearningRate,
			},
		},
		Importance: metrics.Accuracy,
		Tags:       []string{"training", "metrics", "parameters"},
	}

	if err := a.memoryManager.StoreMemory(ctx, memory); err != nil {
		return metrics, fmt.Errorf("erro ao armazenar métricas de treinamento: %v", err)
	}

	return metrics, nil
}

// Remember busca memórias relacionadas a um conjunto de tags
func (a *CognitiveAgent) Remember(ctx context.Context, tags []string) ([]*memory.Memory, error) {
	return a.memoryManager.SearchMemories(ctx, a.ID, tags)
}

// Memorize armazena uma nova memória
func (a *CognitiveAgent) Memorize(ctx context.Context, content map[string]interface{}, importance float64, tags []string, isLongTerm bool) error {
	memType := memory.ShortTerm
	var ttl time.Duration

	if isLongTerm {
		memType = memory.LongTerm
	} else {
		ttl = 24 * time.Hour // Memórias de curto prazo expiram em 24 horas
	}

	memory := &memory.Memory{
		ID:         fmt.Sprintf("memory_%s_%d", a.ID, time.Now().Unix()),
		AgentID:    a.ID,
		Type:       memType,
		Content:    content,
		Importance: importance,
		TTL:        ttl,
		Tags:       tags,
	}

	return a.memoryManager.StoreMemory(ctx, memory)
}

// ConsolidateMemories move memórias importantes de curto prazo para longo prazo
func (a *CognitiveAgent) ConsolidateMemories(ctx context.Context) error {
	return a.memoryManager.ConsolidateMemories(ctx, a.ID)
}

// ForgetOldMemories remove memórias antigas ou irrelevantes
func (a *CognitiveAgent) ForgetOldMemories(ctx context.Context) error {
	return a.memoryManager.PruneMemories(ctx, a.ID)
}

// adjustParameters ajusta os parâmetros do agente baseado no histórico
func (a *CognitiveAgent) adjustParameters() {
	// Ajusta temperatura baseado no sucesso das respostas
	successRate := a.PerformanceStats["success_rate"]
	if successRate < 0.5 {
		a.Temperature *= 0.9 // Reduz temperatura para respostas mais conservadoras
	} else {
		a.Temperature *= 1.1 // Aumenta temperatura para mais criatividade
	}

	// Limita temperatura entre 0.1 e 1.0
	if a.Temperature < 0.1 {
		a.Temperature = 0.1
	} else if a.Temperature > 1.0 {
		a.Temperature = 1.0
	}

	// Ajusta taxa de aprendizado
	a.LearningRate *= 0.95 // Diminui gradualmente
	if a.LearningRate < 0.0001 {
		a.LearningRate = 0.0001
	}
}

// updatePerformanceStats atualiza as estatísticas de performance
func (a *CognitiveAgent) updatePerformanceStats(metrics *TrainingMetrics) {
	// Calcula tempo médio de resposta
	responseTime := metrics.EndTime.Sub(metrics.StartTime).Seconds()
	a.PerformanceStats["response_time"] = (a.PerformanceStats["response_time"]*0.9 + responseTime*0.1)

	// Atualiza taxa de sucesso
	if len(metrics.Errors) == 0 {
		a.PerformanceStats["success_rate"] = (a.PerformanceStats["success_rate"]*0.9 + 1.0*0.1)
	} else {
		a.PerformanceStats["success_rate"] = (a.PerformanceStats["success_rate"] * 0.9)
	}

	// Atualiza score de aprendizado
	learningProgress := float64(metrics.RoundsExecuted) / float64(a.MaxRounds)
	a.PerformanceStats["learning_score"] = learningProgress
}

// Validate implementa validação específica para o agente cognitivo
func (a *CognitiveAgent) Validate(ctx context.Context) error {
	// Primeiro executa validação base
	if err := a.BaseAgent.Validate(ctx); err != nil {
		return err
	}

	// Validações específicas do agente cognitivo
	if a.Temperature <= 0 {
		return fmt.Errorf("temperatura inválida: %v", a.Temperature)
	}

	if a.MaxTokens <= 0 {
		return fmt.Errorf("número máximo de tokens inválido: %v", a.MaxTokens)
	}

	if a.ContextWindow <= 0 {
		return fmt.Errorf("tamanho da janela de contexto inválido: %v", a.ContextWindow)
	}

	// Verifica performance mínima
	if a.PerformanceStats["success_rate"] < 0.5 {
		return fmt.Errorf("taxa de sucesso muito baixa: %v", a.PerformanceStats["success_rate"])
	}

	return nil
}

// GetPerformanceStats retorna as estatísticas de performance
func (a *CognitiveAgent) GetPerformanceStats() map[string]float64 {
	stats := make(map[string]float64)
	for k, v := range a.PerformanceStats {
		stats[k] = v
	}
	return stats
}

// AddPromptTemplate adiciona um template de prompt
func (a *CognitiveAgent) AddPromptTemplate(name, template string) {
	a.PromptTemplates[name] = template
}

// GetPromptTemplate retorna um template de prompt
func (a *CognitiveAgent) GetPromptTemplate(name string) (string, bool) {
	template, ok := a.PromptTemplates[name]
	return template, ok
}

// AddToKnowledgeBase adiciona informação à base de conhecimento
func (a *CognitiveAgent) AddToKnowledgeBase(key string, value interface{}) {
	a.KnowledgeBase[key] = value
}

// GetFromKnowledgeBase recupera informação da base de conhecimento
func (a *CognitiveAgent) GetFromKnowledgeBase(key string) (interface{}, bool) {
	value, ok := a.KnowledgeBase[key]
	return value, ok
}

// AddResponse adiciona uma resposta ao histórico
func (a *CognitiveAgent) AddResponse(response string) {
	a.ResponseHistory = append(a.ResponseHistory, response)
}

// GetResponseHistory retorna o histórico de respostas
func (a *CognitiveAgent) GetResponseHistory() []string {
	history := make([]string, len(a.ResponseHistory))
	copy(history, a.ResponseHistory)
	return history
}

// GetRole retorna o papel do agente
func (a *CognitiveAgent) GetRole() string {
	return a.Role
}

// GetGoal retorna o objetivo do agente
func (a *CognitiveAgent) GetGoal() string {
	return a.Goal
}

// GetAllowDelegation retorna se o agente pode delegar tarefas
func (a *CognitiveAgent) GetAllowDelegation() bool {
	return a.AllowDelegation
}

// GetBackstory retorna a história/contexto do agente
func (a *CognitiveAgent) GetBackstory() string {
	return a.Backstory
}

// SetBackstory define a história/contexto do agente
func (a *CognitiveAgent) SetBackstory(backstory string) {
	a.Backstory = backstory
}
