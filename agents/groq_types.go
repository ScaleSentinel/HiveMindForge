package agents

// GroqRequest representa a estrutura da requisição para a API da Groq
type GroqRequest struct {
	Messages []Message `json:"messages"`
}

// Message representa uma mensagem no formato da API da Groq
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// GroqResponse representa a estrutura da resposta da API da Groq
type GroqResponse struct {
	ID                string   `json:"id"`
	Object            string   `json:"object"`
	Created           int64    `json:"created"`
	Model             string   `json:"model"`
	SystemFingerprint string   `json:"system_fingerprint"`
	Choices           []Choice `json:"choices"`
	Usage             Usage    `json:"usage"`
}

// Choice representa uma escolha na resposta da API da Groq
type Choice struct {
	Index        int     `json:"index"`
	Message      Message `json:"message"`
	FinishReason string  `json:"finish_reason"`
}

// Usage representa informações de uso da API da Groq
type Usage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}
