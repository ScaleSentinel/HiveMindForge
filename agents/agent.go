package agents

// Agent representa a estrutura base de um agente
type Agent struct {
	ID              string
	Name            string
	Role            string
	Goal            string
	AllowDelegation bool
	Model           string
	Backstory       string
}

// Clone cria uma cópia do agente
func (a *Agent) Clone() *Agent {
	return &Agent{
		ID:              a.ID,
		Name:            a.Name,
		Role:            a.Role,
		Goal:            a.Goal,
		AllowDelegation: a.AllowDelegation,
		Model:           a.Model,
		Backstory:       a.Backstory,
	}
}
