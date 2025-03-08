package agents

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
)

// HiveMind gerencia um conjunto de agentes e tarefas
type HiveMind struct {
	Agents  []*Agent
	Tasks   []*Task
	Verbose bool
}

// GroqRequest representa a estrutura da requisição para a API da Groq
type GroqRequest struct {
	Model    string    `json:"model"`
	Messages []Message `json:"messages"`
}

// Message representa uma mensagem no formato da API da Groq
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// GroqResponse representa a resposta da API da Groq
type GroqResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
}

// NewHiveMind cria uma nova instância do HiveMind
func NewHiveMind(verbose bool) *HiveMind {
	return &HiveMind{
		Agents:  make([]*Agent, 0),
		Tasks:   make([]*Task, 0),
		Verbose: verbose,
	}
}

// AssignAgents atribui agentes ao HiveMind
func (h *HiveMind) AssignAgents(agents ...*Agent) {
	h.Agents = append(h.Agents, agents...)
}

// AssignTasks atribui tarefas ao HiveMind
func (h *HiveMind) AssignTasks(tasks ...*Task) {
	h.Tasks = append(h.Tasks, tasks...)
}

// Execute executa todas as tarefas usando os agentes apropriados
func (h *HiveMind) Execute() (string, error) {
	if len(h.Tasks) == 0 {
		return "", fmt.Errorf("nenhuma tarefa para executar")
	}

	// Por enquanto, vamos executar apenas a primeira tarefa
	task := h.Tasks[0]
	if task.Agent == nil && len(h.Agents) > 0 {
		task.Agent = h.Agents[0]
	}

	if task.Agent == nil {
		return "", fmt.Errorf("tarefa sem agente associado")
	}

	if h.Verbose {
		log.Printf("🤖 Agente %s executando tarefa: %s", task.Agent.Name, task.Description)
	}

	// Preparar a requisição para a Groq
	groqReq := GroqRequest{
		Model: "llama-3.3-70b-versatile",
		Messages: []Message{
			{
				Role: "system",
				Content: fmt.Sprintf("Você é %s. Seu papel é %s. Seu objetivo é %s. Backstory: %s",
					task.Agent.Name,
					task.Agent.Role,
					task.Agent.Goal,
					task.Agent.Backstory),
			},
			{
				Role:    "user",
				Content: task.Description,
			},
		},
	}

	// Converter para JSON
	jsonData, err := json.Marshal(groqReq)
	if err != nil {
		return "", fmt.Errorf("erro ao criar JSON: %v", err)
	}

	// Fazer a requisição para a Groq
	req, err := http.NewRequest("POST", "https://api.groq.com/openai/v1/chat/completions", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("erro ao criar requisição: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+os.Getenv("GROQ_API_KEY"))

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("erro ao fazer requisição: %v", err)
	}
	defer resp.Body.Close()

	// Decodificar a resposta
	var groqResp GroqResponse
	if err := json.NewDecoder(resp.Body).Decode(&groqResp); err != nil {
		return "", fmt.Errorf("erro ao decodificar resposta: %v", err)
	}

	if len(groqResp.Choices) == 0 {
		return "", fmt.Errorf("nenhuma resposta recebida da API")
	}

	result := groqResp.Choices[0].Message.Content

	if h.Verbose {
		log.Printf("✅ Tarefa concluída. Resultado: %s", result)
	}

	// Limpar as tarefas e agentes após a execução
	h.Tasks = make([]*Task, 0)
	h.Agents = make([]*Agent, 0)

	return result, nil
}
