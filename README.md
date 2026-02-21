# SRE com CI/CD e Auto-Deploy

Projeto de estudos para entender na prática como funciona um workflow completo de CI/CD com auto-deployment via webhooks.

## O que faz

Basicamente você faz um commit no código, a pipeline testa e builda automaticamente, dispara um webhook pra sua VPS, e a aplicação fica online com a nova versão. Tudo automático.

```
git push → GitLab Pipeline → Webhook → VPS → Deploy
```

## Arquitetura

A aplicação é uma API em Go que roda em 3 instâncias (replicas) em produção. Traefik faz o load balancing entre elas, Redis fica como cache compartilhado.

- **Traefik**: Load balancer + Reverse proxy
- **App Go**: 3 replicas da aplicação
- **Redis**: Cache/session store
- **Prometheus**: Métricas de performance

Tudo rodando em Docker Compose pra fácil gestão.

## Como funciona o deploy

1. Você faz commit e push pro main
2. GitLab roda a pipeline (build, test, segurança)
3. Após sucesso, dispara webhook HTTP pra VPS
4. Servidor recebe webhook e valida assinatura (HMAC-SHA256)
5. Faz git pull da nova versão
6. Docker-compose pull e inicia nova versão
7. Health check verifica se tá tudo ok
8. Pronto, app atualizada com zero downtime

Tudo isso leva uns 5-10 minutos dependendo do tamanho.

## Estrutura

```
.
├── README.md                  (este arquivo)
├── docker-compose.yml         (orquestra tudo)
├── .env                       (variáveis de ambiente)
│
└── app/
    ├── main.go                (aplicação Go)
    ├── go.mod                 (dependências)
    ├── Dockerfile             (build da imagem)
    └── templates/
        └── index.html
```

## Executar localmente

```bash
git clone https://github.com/seu-usuario/meu-projeto-sre.git
cd meu-projeto-sre

docker-compose up -d
```

Acesso:
- Aplicação: http://localhost
- Traefik Dashboard: http://localhost:8080
- Métricas: http://localhost/metrics

```bash
# Ver logs
docker-compose logs -f sre-app

# Parar
docker-compose down
```

## Monitoramento

A app expõe métricas em `/metrics`:

```
http_requests_total           (total de requisições)
http_request_duration_seconds (latência)
redis_connections            (conexões ativas)
```

Pode integrar com Grafana pra visualizar melhor.

## Segurança

- Webhook é assinado com HMAC-SHA256
- Secrets em variáveis de ambiente
- Health checks verificam se app tá saudável
- Containers isolados em rede Docker

## Por que fiz isso

Pra entender na prática:

- Containerização com Docker
- Orquestração com Docker Compose
- CI/CD automático
- Auto-deployment via webhooks
- Load balancing
- Cache com Redis
- Métricas com Prometheus
- Alta disponibilidade

Basicamente como as coisas funcionam em produção real.

## Licença

MIT
