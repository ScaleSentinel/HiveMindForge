# AIAgentsForge

AIAgentsForge é uma plataforma para criação e gerenciamento de agentes cognitivos autônomos, com foco em tarefas específicas como marketing, análise de dados e geração de conteúdo.

## Funcionalidades

- Sistema de memória híbrido (Redis + MongoDB) para armazenamento de curto e longo prazo
- Agentes cognitivos especializados com diferentes papéis e objetivos
- Configuração flexível via arquivos YAML
- Ferramentas integradas para pesquisa, análise e criação de conteúdo
- Sistema de tarefas com dependências e contexto
- Monitoramento em tempo real do estado dos agentes

## Requisitos

- Go 1.21 ou superior
- Redis 6.0 ou superior
- MongoDB 4.4 ou superior
- OpenTelemetry para métricas e tracing

## Instalação

1. Clone o repositório:
```bash
git clone https://github.com/seu-usuario/AIAgentsForge.git
cd AIAgentsForge
```

2. Instale as dependências:
```bash
go mod tidy
```

3. Configure as variáveis de ambiente:
```bash
export REDIS_URL="redis://localhost:6379"
export MONGO_URL="mongodb://localhost:27017"
export SERPER_API_KEY="sua-chave-api"
```

## Configuração

O projeto usa arquivos YAML para configuração dos agentes, tarefas e ferramentas:

- `config/agents.yaml`: Configuração dos agentes cognitivos
- `config/tasks.yaml`: Configuração das tarefas e seus fluxos
- `config/tools.yaml`: Configuração das ferramentas disponíveis

## Uso

### Marketing Posts

O exemplo de Marketing Posts demonstra como criar uma campanha de marketing usando agentes cognitivos:

```bash
cd examples/marketing
go run main.go
```

O exemplo inclui:
- Análise de mercado
- Desenvolvimento de estratégia
- Criação de campanha
- Geração de conteúdo

### Estrutura do Projeto

```
.
├── agents/
│   ├── cognitive_agent.go
│   ├── base_agent.go
│   ├── memory/
│   │   ├── types.go
│   │   └── manager.go
│   └── marketing/
│       ├── marketing_posts.go
│       ├── config.go
│       └── tools.go
├── config/
│   ├── agents.yaml
│   ├── tasks.yaml
│   └── tools.yaml
└── examples/
    ├── marketing/
    │   └── main.go
    └── memory/
        └── main.go
```

## Desenvolvimento

Para contribuir com o projeto:

1. Fork o repositório
2. Crie uma branch para sua feature (`git checkout -b feature/nova-feature`)
3. Commit suas mudanças (`git commit -am 'Adiciona nova feature'`)
4. Push para a branch (`git push origin feature/nova-feature`)
5. Crie um Pull Request

## Licença

Este projeto está licenciado sob a MIT License - veja o arquivo [LICENSE](LICENSE) para detalhes.

## Contato

- Email: seu-email@exemplo.com
- Twitter: @seu-usuario
- LinkedIn: linkedin.com/in/seu-usuario 