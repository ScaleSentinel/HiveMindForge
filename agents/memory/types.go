package memory

import (
	"context"
	"time"
)

// MemoryType representa o tipo de memória (curto ou longo prazo)
type MemoryType string

const (
	ShortTerm MemoryType = "short_term"
	LongTerm  MemoryType = "long_term"
)

// Memory representa uma unidade de memória do agente
type Memory struct {
	ID          string                 `json:"id" bson:"_id"`
	AgentID     string                 `json:"agent_id" bson:"agent_id"`
	Type        MemoryType             `json:"type" bson:"type"`
	Content     map[string]interface{} `json:"content" bson:"content"`
	Importance  float64                `json:"importance" bson:"importance"`
	CreatedAt   time.Time              `json:"created_at" bson:"created_at"`
	AccessCount int                    `json:"access_count" bson:"access_count"`
	LastAccess  time.Time              `json:"last_access" bson:"last_access"`
	TTL         time.Duration          `json:"ttl" bson:"ttl"`
	Tags        []string               `json:"tags" bson:"tags"`
}

// MemoryManager gerencia o armazenamento e recuperação de memórias
type MemoryManager interface {
	// StoreMemory armazena uma nova memória
	StoreMemory(ctx context.Context, memory *Memory) error

	// GetMemory recupera uma memória pelo ID
	GetMemory(ctx context.Context, agentID, memoryID string) (*Memory, error)

	// SearchMemories busca memórias por tags
	SearchMemories(ctx context.Context, agentID string, tags []string) ([]*Memory, error)

	// UpdateMemory atualiza uma memória existente
	UpdateMemory(ctx context.Context, memory *Memory) error

	// DeleteMemory remove uma memória específica
	DeleteMemory(ctx context.Context, agentID, memoryID string) error

	// ConsolidateMemories move memórias importantes de curto prazo para longo prazo
	ConsolidateMemories(ctx context.Context, agentID string) error

	// PruneMemories remove memórias antigas ou irrelevantes
	PruneMemories(ctx context.Context, agentID string) error

	// Close fecha as conexões com os bancos de dados
	Close(ctx context.Context) error
}

// MemoryConfig contém as configurações para o gerenciador de memória
type MemoryConfig struct {
	RedisURL            string        // URL de conexão com o Redis
	MongoURL            string        // URL de conexão com o MongoDB
	MongoDB             string        // Nome do banco de dados MongoDB
	Collection          string        // Nome da coleção MongoDB
	ShortTermTTL        time.Duration // Tempo de vida padrão para memórias de curto prazo
	ImportanceThreshold float64       // Limiar de importância para consolidação
}

// DefaultMemoryConfig retorna uma configuração padrão para o gerenciador de memória
func DefaultMemoryConfig() *MemoryConfig {
	return &MemoryConfig{
		RedisURL:            "redis://localhost:6379",
		MongoURL:            "mongodb://localhost:27017",
		MongoDB:             "agent_memories",
		Collection:          "memories",
		ShortTermTTL:        24 * time.Hour,
		ImportanceThreshold: 0.7,
	}
}
