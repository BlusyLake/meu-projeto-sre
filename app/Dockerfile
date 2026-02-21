# STAGE 1: O Compilador (Pesado)
FROM golang:1.25.7-alpine AS builder

WORKDIR /src

# Traz os manifestos de dependência e baixa (cache layer)
COPY go.mod go.sum ./
RUN go mod download

# Traz o código e compila estaticamente (sem dependência de CGO)
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /bin/sre-app main.go

# STAGE 2: O Runtime (Leve e Seguro)
FROM alpine:3.19

WORKDIR /app

# Copia APENAS o binário e a pasta de templates do estágio anterior
COPY --from=builder /bin/sre-app /app/sre-app
COPY --from=builder /src/templates /app/templates

# Nunca rode como root.
RUN adduser -D sreuser
USER sreuser

EXPOSE 3000

ENTRYPOINT ["/app/sre-app"]