package agents

import (
	"log"
	"runtime"
	"sync"
	"time"
)

// ObserverInfrastructureAgent monitora o estado do sistema e coleta métricas
type ObserverInfrastructureAgent struct {
	Agent
	metrics     *SystemMetrics
	metricsLock sync.RWMutex
	stopChan    chan struct{}
}

// NewObserverInfrastructureAgent cria uma nova instância do agente observador
func NewObserverInfrastructureAgent() *ObserverInfrastructureAgent {
	agent := &ObserverInfrastructureAgent{
		Agent: Agent{
			Name:            "System Observer",
			Role:            "Monitor do Sistema",
			Goal:            "Monitorar e coletar métricas do sistema",
			AllowDelegation: false,
			Model:           "gpt-4",
			Backstory:       "Um agente especializado em monitorar e analisar o desempenho do sistema",
		},
		metrics:  &SystemMetrics{},
		stopChan: make(chan struct{}),
	}

	go agent.startMonitoring()
	return agent
}

// startMonitoring inicia a coleta de métricas
func (o *ObserverInfrastructureAgent) startMonitoring() {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			o.collectMetrics()
		case <-o.stopChan:
			return
		}
	}
}

// collectMetrics coleta métricas do sistema
func (o *ObserverInfrastructureAgent) collectMetrics() {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	o.metricsLock.Lock()
	defer o.metricsLock.Unlock()

	// Calcular uso de CPU (exemplo simplificado)
	cpuUsage := float64(runtime.NumGoroutine()) / float64(runtime.NumCPU()) * 100

	// Calcular uso de memória
	memoryUsage := float64(m.Alloc) / float64(m.Sys) * 100

	o.metrics.CPUUsage = cpuUsage
	o.metrics.MemoryUsage = memoryUsage
	o.metrics.LastUpdate = time.Now().Unix()

	log.Printf("📊 Métricas atualizadas - CPU: %.2f%%, Memória: %.2f%%",
		cpuUsage, memoryUsage)
}

// GetSystemMetrics retorna as métricas atuais do sistema
func (o *ObserverInfrastructureAgent) GetSystemMetrics() *SystemMetrics {
	o.metricsLock.RLock()
	defer o.metricsLock.RUnlock()
	return o.metrics
}

// Stop interrompe o monitoramento
func (o *ObserverInfrastructureAgent) Stop() {
	close(o.stopChan)
}
