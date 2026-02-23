# Projeto SRE - Estudo de Caso DevOps com Observabilidade

Este repositório contém um **caso de estudo** focado em ciclos de desenvolvimento DevOps, com ênfase na observabilidade. O objetivo é demonstrar práticas de integração entre aplicação, infraestrutura, monitoramento e dashboards, utilizando uma arquitetura simples para facilitar o aprendizado.


---

## 🔍 Visão Geral

O projeto é composto por:

- **app/**: código da aplicação Go (serviço HTTP simples) junto com templates para páginas web.
- **infra/**: configuração de infraestrutura para monitoramento (Prometheus, Grafana) em `docker-compose` separado.
- **docker-compose.yml**: orquestração geral para executar a aplicação e dependências.
- **prometheus.yml**: configuração principal do Prometheus.

A intenção é executar os dois ambientes (aplicação e monitoramento) localmente, observando métricas, logs e criando dashboards para análises.

---

## 🧱 Estrutura de Diretórios

```
/ (root do repositório)
├── docker-compose.yml          # Compose principal (app + serviços)
├── prometheus.yml              # Configuração de Prometheus
├── README.md                   # Este documento
├── app/                        # Código da aplicação
    ├── Dockerfile
    ├── go.mod
    ├── main.go
    └── templates/              # Views HTML
        ├── dashboard.html
        └── index.html

```

---

## 🚀 Executando o Projeto

O projeto foi pensado para rodar com **Docker Compose**. Siga os passos abaixo:

1. **Pré-requisitos**
   - Docker e Docker Compose instalados.
   - Portas 8080, 9090 e 3000 livres.

2. **Iniciar a aplicação + serviços**
   ```bash
   docker-compose up --build
   ```

3. **Catalogar endpoints**
   - Aplicação: `http://localhost:8080`
   - Prometheus: `http://localhost:9090`

4. Para iniciar apenas o ambiente de monitoramento:
   ```bash
   cd infra && docker-compose up
   ```

---

## 🛠 Tecnologias Utilizadas

- **Go** – serviço web simples com métricas expostas para Prometheus.
- **Docker / Docker Compose** – containerização e orquestração.
- **Prometheus** – coleta de métricas.

---

## 📈 Observabilidade & DevOps

Este caso de estudo aborda os seguintes pontos:

1. **Instrumentação** da aplicação para expor métricas de desempenho.
2. **Configuração** de coleta de métricas via Prometheus.
3. **Criação de dashboards** no Grafana usando dados reais.
4. **Ciclos de desenvolvimento** típicos:
   - Build local via `docker-compose`.
   - Adição de novas métricas ou alterações de código.
   - Teste de integração com o stack de monitoramento.
   - Feedback através de dashboards e logs.

A ideia é permitir que engenheiros pratiquem commits, builds e deploys locais enquanto observam impactos em métricas e visualizações.

---

## 📄 Licença

Este projeto está disponível sob a **licença MIT**. Sinta-se à vontade para clonar, modificar e aprender.

---

## 💡 Como usar este repositório

1. **Clone** o repositório publicamente.
2. Explore a estrutura de diretórios e arquivos.
3. Modifique ou estenda a aplicação para adicionar métricas ou funcionalidades.
4. Ajuste as configurações de Prometheus/Grafana para ganhar mais insights.

> Este repositório serve como base para workshops, estudos e demonstrações sobre práticas modernas de DevOps e observabilidade.

---

Se tiver dúvidas ou quiser contribuir, sinta-se livre para abrir issues ou pull requests.

> _Boa exploração!_ 👩‍💻👨‍💻
