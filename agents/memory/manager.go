package memory

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// HybridMemoryManager implementa a interface MemoryManager usando Redis e MongoDB
type HybridMemoryManager struct {
	config      *MemoryConfig
	redisClient *redis.Client
	mongoClient *mongo.Client
	collection  *mongo.Collection
}

// NewHybridMemoryManager cria uma nova instância do gerenciador de memória híbrido
func NewHybridMemoryManager(ctx context.Context, config *MemoryConfig) (*HybridMemoryManager, error) {
	// Conecta ao Redis
	opt, err := redis.ParseURL(config.RedisURL)
	if err != nil {
		return nil, fmt.Errorf("erro ao parsear URL do Redis: %v", err)
	}
	redisClient := redis.NewClient(opt)

	// Testa conexão com Redis
	if err := redisClient.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("erro ao conectar ao Redis: %v", err)
	}

	// Conecta ao MongoDB
	mongoClient, err := mongo.Connect(ctx, options.Client().ApplyURI(config.MongoURL))
	if err != nil {
		return nil, fmt.Errorf("erro ao conectar ao MongoDB: %v", err)
	}

	// Testa conexão com MongoDB
	if err := mongoClient.Ping(ctx, nil); err != nil {
		return nil, fmt.Errorf("erro ao conectar ao MongoDB: %v", err)
	}

	collection := mongoClient.Database(config.MongoDB).Collection(config.Collection)

	// Cria índices no MongoDB
	indexes := []mongo.IndexModel{
		{
			Keys: bson.D{
				{Key: "agent_id", Value: 1},
				{Key: "type", Value: 1},
			},
		},
		{
			Keys: bson.D{
				{Key: "tags", Value: 1},
			},
		},
		{
			Keys: bson.D{
				{Key: "created_at", Value: 1},
			},
		},
	}

	if _, err := collection.Indexes().CreateMany(ctx, indexes); err != nil {
		return nil, fmt.Errorf("erro ao criar índices no MongoDB: %v", err)
	}

	return &HybridMemoryManager{
		config:      config,
		redisClient: redisClient,
		mongoClient: mongoClient,
		collection:  collection,
	}, nil
}

// Close fecha as conexões com Redis e MongoDB
func (m *HybridMemoryManager) Close(ctx context.Context) error {
	if err := m.redisClient.Close(); err != nil {
		return fmt.Errorf("erro ao fechar conexão com Redis: %v", err)
	}

	if err := m.mongoClient.Disconnect(ctx); err != nil {
		return fmt.Errorf("erro ao fechar conexão com MongoDB: %v", err)
	}

	return nil
}

// StoreMemory armazena uma memória no Redis (curto prazo) ou MongoDB (longo prazo)
func (m *HybridMemoryManager) StoreMemory(ctx context.Context, memory *Memory) error {
	memory.CreatedAt = time.Now()

	if memory.Type == ShortTerm {
		// Armazena no Redis com TTL
		data, err := json.Marshal(memory)
		if err != nil {
			return fmt.Errorf("erro ao serializar memória: %v", err)
		}

		key := fmt.Sprintf("memory:%s:%s", memory.AgentID, memory.ID)
		ttl := memory.TTL
		if ttl == 0 {
			ttl = m.config.ShortTermTTL
		}

		if err := m.redisClient.Set(ctx, key, data, ttl).Err(); err != nil {
			return fmt.Errorf("erro ao armazenar memória no Redis: %v", err)
		}
	} else {
		// Armazena no MongoDB
		if _, err := m.collection.InsertOne(ctx, memory); err != nil {
			return fmt.Errorf("erro ao armazenar memória no MongoDB: %v", err)
		}
	}

	return nil
}

// GetMemory recupera uma memória do Redis ou MongoDB
func (m *HybridMemoryManager) GetMemory(ctx context.Context, agentID, memoryID string) (*Memory, error) {
	// Tenta primeiro no Redis
	key := fmt.Sprintf("memory:%s:%s", agentID, memoryID)
	data, err := m.redisClient.Get(ctx, key).Bytes()
	if err == nil {
		var memory Memory
		if err := json.Unmarshal(data, &memory); err != nil {
			return nil, fmt.Errorf("erro ao deserializar memória do Redis: %v", err)
		}
		return &memory, nil
	}

	// Se não encontrou no Redis, busca no MongoDB
	filter := bson.M{"agent_id": agentID, "_id": memoryID}
	var memory Memory
	if err := m.collection.FindOne(ctx, filter).Decode(&memory); err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("memória não encontrada")
		}
		return nil, fmt.Errorf("erro ao buscar memória no MongoDB: %v", err)
	}

	return &memory, nil
}

// SearchMemories busca memórias por tags no Redis e MongoDB
func (m *HybridMemoryManager) SearchMemories(ctx context.Context, agentID string, tags []string) ([]*Memory, error) {
	var memories []*Memory

	// Busca no Redis
	pattern := fmt.Sprintf("memory:%s:*", agentID)
	iter := m.redisClient.Scan(ctx, 0, pattern, 0).Iterator()
	for iter.Next(ctx) {
		data, err := m.redisClient.Get(ctx, iter.Val()).Bytes()
		if err != nil {
			continue
		}

		var memory Memory
		if err := json.Unmarshal(data, &memory); err != nil {
			continue
		}

		// Verifica se a memória tem todas as tags buscadas
		hasAllTags := true
		for _, tag := range tags {
			found := false
			for _, memTag := range memory.Tags {
				if memTag == tag {
					found = true
					break
				}
			}
			if !found {
				hasAllTags = false
				break
			}
		}

		if hasAllTags {
			memories = append(memories, &memory)
		}
	}

	// Busca no MongoDB
	filter := bson.M{
		"agent_id": agentID,
		"tags":     bson.M{"$all": tags},
	}
	cursor, err := m.collection.Find(ctx, filter)
	if err != nil {
		return memories, fmt.Errorf("erro ao buscar memórias no MongoDB: %v", err)
	}
	defer cursor.Close(ctx)

	for cursor.Next(ctx) {
		var memory Memory
		if err := cursor.Decode(&memory); err != nil {
			continue
		}
		memories = append(memories, &memory)
	}

	return memories, nil
}

// ConsolidateMemories move memórias importantes de curto prazo para longo prazo
func (m *HybridMemoryManager) ConsolidateMemories(ctx context.Context, agentID string) error {
	pattern := fmt.Sprintf("memory:%s:*", agentID)
	iter := m.redisClient.Scan(ctx, 0, pattern, 0).Iterator()

	for iter.Next(ctx) {
		data, err := m.redisClient.Get(ctx, iter.Val()).Bytes()
		if err != nil {
			continue
		}

		var memory Memory
		if err := json.Unmarshal(data, &memory); err != nil {
			continue
		}

		// Se a memória é importante o suficiente, move para longo prazo
		if memory.Importance >= m.config.ImportanceThreshold {
			memory.Type = LongTerm
			if err := m.StoreMemory(ctx, &memory); err != nil {
				continue
			}
			m.redisClient.Del(ctx, iter.Val())
		}
	}

	return nil
}

// PruneMemories remove memórias antigas ou irrelevantes
func (m *HybridMemoryManager) PruneMemories(ctx context.Context, agentID string) error {
	// Remove memórias antigas do MongoDB
	threshold := time.Now().Add(-30 * 24 * time.Hour) // 30 dias
	filter := bson.M{
		"agent_id":   agentID,
		"created_at": bson.M{"$lt": threshold},
		"importance": bson.M{"$lt": m.config.ImportanceThreshold},
	}

	if _, err := m.collection.DeleteMany(ctx, filter); err != nil {
		return fmt.Errorf("erro ao remover memórias antigas do MongoDB: %v", err)
	}

	return nil
}

// DeleteMemory remove uma memória específica
func (m *HybridMemoryManager) DeleteMemory(ctx context.Context, agentID, memoryID string) error {
	// Remove do Redis
	key := fmt.Sprintf("memory:%s:%s", agentID, memoryID)
	if err := m.redisClient.Del(ctx, key).Err(); err != nil && err != redis.Nil {
		return fmt.Errorf("erro ao remover memória do Redis: %v", err)
	}

	// Remove do MongoDB
	filter := bson.M{
		"agent_id": agentID,
		"_id":      memoryID,
	}
	if _, err := m.collection.DeleteOne(ctx, filter); err != nil {
		return fmt.Errorf("erro ao remover memória do MongoDB: %v", err)
	}

	return nil
}

// UpdateMemory atualiza uma memória existente
func (m *HybridMemoryManager) UpdateMemory(ctx context.Context, memory *Memory) error {
	// Atualiza no Redis se for memória de curto prazo
	if memory.Type == ShortTerm {
		data, err := json.Marshal(memory)
		if err != nil {
			return fmt.Errorf("erro ao serializar memória: %v", err)
		}

		key := fmt.Sprintf("memory:%s:%s", memory.AgentID, memory.ID)
		ttl := memory.TTL
		if ttl == 0 {
			ttl = m.config.ShortTermTTL
		}

		if err := m.redisClient.Set(ctx, key, data, ttl).Err(); err != nil {
			return fmt.Errorf("erro ao atualizar memória no Redis: %v", err)
		}
	} else {
		// Atualiza no MongoDB
		filter := bson.M{
			"agent_id": memory.AgentID,
			"_id":      memory.ID,
		}
		update := bson.M{"$set": memory}

		if _, err := m.collection.UpdateOne(ctx, filter, update); err != nil {
			return fmt.Errorf("erro ao atualizar memória no MongoDB: %v", err)
		}
	}

	return nil
}
