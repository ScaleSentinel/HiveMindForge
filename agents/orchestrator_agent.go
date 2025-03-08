package agents

import (
	"fmt"
	"log"
)

// OrchestratorAgent é responsável por coordenar o fluxo de trabalho
type OrchestratorAgent struct {
	Agent           // Incorpora os campos básicos do Agent
	CognitiveAgents []*CognitiveAgent
	TotalScore      int
	HiveMind        *HiveMind // Referência para o HiveMind que gerencia as execuções
}

// NewOrchestratorAgent cria uma nova instância do OrchestratorAgent
func NewOrchestratorAgent() *OrchestratorAgent {
	return &OrchestratorAgent{
		Agent: Agent{
			Name:            "Orchestrator",
			Role:            "Coordenador do Fluxo de Trabalho",
			Goal:            "Coordenar a criação de conteúdo educacional e garantir a qualidade do material",
			AllowDelegation: true,
			Model:           "gpt-4o-mini",
			Backstory:       "Um especialista em gestão de projetos educacionais que coordena equipes multidisciplinares",
		},
		TotalScore: 0,
		HiveMind:   NewHiveMind(true), // Criar uma nova instância do HiveMind com verbose=true
	}
}

// AssignCognitiveAgents atribui os agentes cognitivos ao orquestrador
func (o *OrchestratorAgent) AssignCognitiveAgents(agents ...*CognitiveAgent) {
	o.CognitiveAgents = agents
}

// DelegateTask delega uma tarefa para um agente cognitivo específico
func (o *OrchestratorAgent) DelegateTask(agent *CognitiveAgent, task *Task) (string, error) {
	log.Printf("🎯 Delegando tarefa '%s' para o agente %s", task.Description, agent.Name)

	o.HiveMind.AssignAgents(&agent.Agent)
	o.HiveMind.AssignTasks(task)

	result, err := o.HiveMind.Execute()
	if err != nil {
		return "", fmt.Errorf("erro ao executar tarefa: %v", err)
	}

	return result, nil
}

// EvaluateContent avalia o conteúdo produzido e decide se deve ser aprovado
func (o *OrchestratorAgent) EvaluateContent(content string, score int) (bool, error) {
	task := &Task{
		Description:    fmt.Sprintf("Analise este conteúdo e decida se ele atende aos padrões de qualidade: %s", content),
		Agent:          &o.Agent,
		ExpectedOutput: "Uma decisão clara sobre a aprovação ou rejeição do conteúdo.",
	}

	o.HiveMind.AssignAgents(&o.Agent)
	o.HiveMind.AssignTasks(task)

	result, err := o.HiveMind.Execute()
	if err != nil {
		return false, fmt.Errorf("erro ao avaliar conteúdo: %v", err)
	}

	approved := result != "" && (result == "APROVADO" || result == "Aprovado" || result == "aprovado")
	if approved {
		o.TotalScore += score
		log.Printf("✅ Conteúdo aprovado! Nova pontuação total: %d", o.TotalScore)
	} else {
		log.Printf("❌ Conteúdo rejeitado")
	}

	return approved, nil
}

// IsWorkflowComplete verifica se o fluxo de trabalho está completo
func (o *OrchestratorAgent) IsWorkflowComplete() bool {
	return o.TotalScore >= 1000
}

// GetProgress retorna o progresso atual do fluxo de trabalho
func (o *OrchestratorAgent) GetProgress() float64 {
	return float64(o.TotalScore) / 1000.0 * 100
}
