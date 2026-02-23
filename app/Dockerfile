FROM golang:1.25.7-alpine AS builder
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /bin/sre-app main.go


FROM alpine:3.19
WORKDIR /app


ARG CI_USER="Sysadmin"
ENV APP_AUTHOR=$CI_USER


COPY --from=builder /bin/sre-app /app/sre-app
COPY --from=builder /src/templates /app/templates
COPY --from=builder /src/pipeline-copy.yml /app/pipeline.yml


RUN adduser -D sreuser
USER sreuser
EXPOSE 3000

ENTRYPOINT ["/app/sre-app"]