package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"HiveMindForge/agents"
)

func main() {
	// Cria configuração de treinamento
	config := agents.TrainingConfig{
		MaxRounds:       5,
		TrainingTimeout: 2 * time.Second,
		ValidationRatio: 0.2,
		MinAccuracy:     0.8,
		BatchSize:       32,
		LearningRate:    0.001,
		UseHistorical:   true,
	}

	// Cria o gerenciador de treinamento
	trainer := agents.NewAgentTrainer(config)

	// Cria alguns agentes para teste
	agent1 := agents.NewBaseAgent("agent1", "Agente 1", "Agente de teste 1", 3)
	agent2 := agents.NewBaseAgent("agent2", "Agente 2", "Agente de teste 2", 5)
	agent3 := agents.NewBaseAgent("agent3", "Agente 3", "Agente de teste 3", 4)

	// Adiciona os agentes ao trainer
	trainer.AddAgent(agent1)
	trainer.AddAgent(agent2)
	trainer.AddAgent(agent3)

	// Cria contexto com timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Executa o treinamento
	fmt.Println("Iniciando treinamento dos agentes...")
	if err := trainer.Train(ctx); err != nil {
		log.Fatalf("Erro durante treinamento: %v", err)
	}

	// Exibe métricas de cada agente
	fmt.Println("\nMétricas de treinamento:")
	metrics := trainer.GetAllMetrics()
	for agent, metric := range metrics {
		var baseAgent *agents.BaseAgent
		switch a := agent.(type) {
		case *agents.BaseAgent:
			baseAgent = a
		}

		if baseAgent != nil {
			fmt.Printf("\nAgente: %s\n", baseAgent.Name)
			fmt.Printf("Rounds executados: %d/%d\n", metric.RoundsExecuted, baseAgent.GetMaxRounds())
			fmt.Printf("Tempo de treinamento: %v\n", metric.EndTime.Sub(metric.StartTime))
			if len(metric.Errors) > 0 {
				fmt.Printf("Erros: %v\n", metric.Errors)
			} else {
				fmt.Println("Sem erros")
			}
		}
	}

	// Tenta salvar o estado dos agentes
	fmt.Println("\nSalvando estado dos agentes...")
	for agent := range metrics {
		var baseAgent *agents.BaseAgent
		switch a := agent.(type) {
		case *agents.BaseAgent:
			baseAgent = a
		}

		if baseAgent != nil {
			path := fmt.Sprintf("agent_%s_state.json", baseAgent.ID)
			if err := baseAgent.SaveState(path); err != nil {
				fmt.Printf("Erro ao salvar estado do agente %s: %v\n", baseAgent.Name, err)
			} else {
				fmt.Printf("Estado do agente %s salvo em %s\n", baseAgent.Name, path)
			}
		}
	}
}
