package agents

import (
	"time"
)

// TrainingConfig contém configurações para treinamento de agentes
type TrainingConfig struct {
	UseHistorical bool    // Se deve usar dados históricos no treinamento
	BatchSize     int     // Tamanho do lote de treinamento
	Epochs        int     // Número de épocas de treinamento
	LearningRate  float64 // Taxa de aprendizado
}

// TrainingMetrics contém métricas do treinamento
type TrainingMetrics struct {
	StartTime      time.Time          // Quando o treinamento começou
	EndTime        time.Time          // Quando o treinamento terminou
	RoundsExecuted int                // Número de rounds executados
	Accuracy       float64            // Acurácia do modelo
	Loss           float64            // Perda do modelo
	Errors         []error            // Erros encontrados durante o treinamento
	Metrics        map[string]float64 // Métricas adicionais
}
