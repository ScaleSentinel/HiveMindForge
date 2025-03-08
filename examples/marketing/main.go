package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"HiveMindForge/agents/marketing"
	"HiveMindForge/agents/memory"
)

func main() {
	ctx := context.Background()

	// Obtém o diretório base do projeto
	baseDir, err := os.Getwd()
	if err != nil {
		log.Fatalf("Erro ao obter diretório atual: %v", err)
	}

	// Carrega as configurações
	agentsConfig, err := marketing.LoadAgentsConfig(filepath.Join(baseDir, "config", "agents.yaml"))
	if err != nil {
		log.Fatalf("Erro ao carregar configuração dos agentes: %v", err)
	}

	tasksConfig, err := marketing.LoadTasksConfig(filepath.Join(baseDir, "config", "tasks.yaml"))
	if err != nil {
		log.Fatalf("Erro ao carregar configuração das tarefas: %v", err)
	}

	toolsConfig, err := marketing.LoadToolsConfig(filepath.Join(baseDir, "config", "tools.yaml"))
	if err != nil {
		log.Fatalf("Erro ao carregar configuração das ferramentas: %v", err)
	}

	// Configuração do gerenciador de memória
	memConfig := memory.DefaultMemoryConfig()
	memManager, err := memory.NewHybridMemoryManager(ctx, memConfig)
	if err != nil {
		log.Fatalf("Erro ao criar gerenciador de memória: %v", err)
	}
	defer memManager.Close(ctx)

	// Exibe configuração dos agentes
	log.Printf("\nAgentes configurados:")
	log.Printf("- Analista: %s (%s)", agentsConfig.LeadMarketAnalyst.Name, agentsConfig.LeadMarketAnalyst.Role)
	log.Printf("  - Objetivo: %s", agentsConfig.LeadMarketAnalyst.Goal)
	log.Printf("  - História: %s", agentsConfig.LeadMarketAnalyst.Backstory)
	log.Printf("- Estrategista: %s (%s)", agentsConfig.ChiefMarketingStrategist.Name, agentsConfig.ChiefMarketingStrategist.Role)
	log.Printf("  - Objetivo: %s", agentsConfig.ChiefMarketingStrategist.Goal)
	log.Printf("  - História: %s", agentsConfig.ChiefMarketingStrategist.Backstory)
	log.Printf("- Criador: %s (%s)", agentsConfig.CreativeContentCreator.Name, agentsConfig.CreativeContentCreator.Role)
	log.Printf("  - Objetivo: %s", agentsConfig.CreativeContentCreator.Goal)
	log.Printf("  - História: %s", agentsConfig.CreativeContentCreator.Backstory)

	// Cria a equipe de marketing
	crew := marketing.NewMarketingPostsCrew(ctx, memManager)

	// Define os detalhes do projeto
	projectDetails := map[string]interface{}{
		"name":        "Campanha de Marketing Digital",
		"objective":   "Aumentar engajamento nas redes sociais",
		"target":      "Profissionais de marketing digital",
		"budget":      50000,
		"duration":    "3 meses",
		"channels":    []string{"LinkedIn", "Twitter", "Email"},
		"constraints": []string{"ROI positivo em 6 meses", "Foco em conteúdo educativo"},
		"tasks": map[string]interface{}{
			"research":      tasksConfig.ResearchTask,
			"understanding": tasksConfig.ProjectUnderstandingTask,
			"strategy":      tasksConfig.MarketingStrategyTask,
			"campaign":      tasksConfig.CampaignIdeaTask,
			"copy":          tasksConfig.CopyCreationTask,
		},
		"tools": map[string]interface{}{
			"research": toolsConfig.GetToolsByCategory("research"),
			"planning": toolsConfig.GetToolsByCategory("planning"),
			"creative": toolsConfig.GetToolsByCategory("creative"),
			"analysis": toolsConfig.GetToolsByCategory("analysis"),
		},
	}

	// Exibe as ferramentas disponíveis
	log.Printf("\nFerramentas disponíveis por categoria:")
	categories := []string{"research", "planning", "creative", "analysis"}
	for _, category := range categories {
		tools := toolsConfig.GetToolsByCategory(category)
		log.Printf("\n%s:", category)
		for _, tool := range tools {
			log.Printf("- %s: %s", tool.Name, tool.Description)
		}
	}

	// Executa o fluxo de trabalho
	result, err := crew.ExecuteWorkflow(projectDetails)
	if err != nil {
		log.Fatalf("Erro ao executar workflow: %v", err)
	}

	// Exibe os resultados
	log.Printf("\nResultados do workflow:\n%s", result)

	// Aguarda sinais de interrupção
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Monitora o processo
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			// Exibe estatísticas periódicas
			log.Printf("\nStatus do projeto:")
			log.Printf("- Estratégia: %s", result.Strategy.Name)
			log.Printf("  - Táticas: %v", result.Strategy.Tactics)
			log.Printf("  - Canais: %v", result.Strategy.Channels)
			log.Printf("  - KPIs: %v", result.Strategy.KPIs)
			log.Printf("- Campanha: %s", result.Campaign.Name)
			log.Printf("  - Descrição: %s", result.Campaign.Description)
			log.Printf("  - Público: %s", result.Campaign.Audience)
			log.Printf("  - Canal: %s", result.Campaign.Channel)
			log.Printf("- Copy:")
			log.Printf("  - Título: %s", result.Copy.Title)
			log.Printf("  - Texto: %s", result.Copy.Body)

		case sig := <-sigChan:
			log.Printf("\nRecebido sinal %v, encerrando...", sig)
			return
		}
	}
}
