package main

import (
	"context"
	"fmt"
	"html/template"
	"log"
	"math/rand"
	"net/http"
	"os"
	"runtime"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/redis/go-redis/v9"
)

// --- 1. DEFINIÇÃO DAS MÉTRICAS (PROMETHEUS) ---
// SRE sem métrica é só um cara com achismos.
var (
	httpRequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total de requisições HTTP processadas, separadas por status.",
		},
		[]string{"status"},
	)
	httpDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "Latência das requisições HTTP.",
			Buckets: []float64{0.1, 0.5, 1, 2, 5}, // Buckets de tempo em segundos
		},
		[]string{"handler"},
	)
)

func init() {
	// Registra as métricas no Prometheus local
	prometheus.MustRegister(httpRequestsTotal)
	prometheus.MustRegister(httpDuration)
}

// --- 2. ESTRUTURA DE DADOS PARA O FRONTEND ---
// O que vamos mandar para o nosso HTML renderizar
type PageData struct {
	Hostname       string
	Visits         string
	RedisStatus    string
	RedisIsDown    bool
	LastPingTime   string
	OSArchitecture string
	GoVersion      string
	CPUCores       int
}

// Cliente global do Redis
var rdb *redis.Client

func main() {
	// Pega a URL do Redis via variável de ambiente (Docker injeta isso depois)
	// Se não tiver, tenta o localhost (para seus testes locais)
	redisAddr := os.Getenv("REDIS_ADDR")
	if redisAddr == "" {
		redisAddr = "localhost:6379"
	}

	rdb = redis.NewClient(&redis.Options{
		Addr:        redisAddr,
		Password:    "", // Sem senha para o laboratório
		DB:          0,
		DialTimeout: 50 * time.Millisecond, // SRE Rule: Fail Fast! Se demorar > 50ms para conectar, desiste.
		ReadTimeout: 50 * time.Millisecond,
	})

	// Roteamento
	http.HandleFunc("/", homeHandler)
	http.HandleFunc("/chaos", chaosHandler)
	http.Handle("/metrics", promhttp.Handler()) // Endpoint sagrado do Prometheus

	porta := "3000"
	fmt.Printf("🚀 [SRE App] Subindo na porta %s...\n", porta)
	log.Fatal(http.ListenAndServe(":"+porta, nil))
}

// --- 3. HANDLERS (A Lógica de Negócio e Falha) ---

func homeHandler(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	// Ignora requisições de favicon para não sujar nossas métricas
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	hostname, _ := os.Hostname()
	ctx := context.Background()

	// Tenta incrementar no Redis (Nosso ponto de falha potencial)
	visits := "0"
	redisStatus := "Online"
	redisIsDown := false

	val, err := rdb.Incr(ctx, "contador_visitas").Result()
	if err != nil {
		// Graceful Degradation: O banco caiu? O app continua de pé.
		redisStatus = "Offline (Graceful Degradation Ativo)"
		visits = "Erro ao ler cache"
		redisIsDown = true
		fmt.Printf("⚠️ [Alerta] Falha no Redis: %v\n", err)
	} else {
		visits = fmt.Sprintf("%d", val)
	}

	// Prepara os dados pro HTML
	data := PageData{
		Hostname:       hostname,
		Visits:         visits,
		RedisStatus:    redisStatus,
		RedisIsDown:    redisIsDown,
		LastPingTime:   time.Now().Format("15:04:05.000"),
		OSArchitecture: runtime.GOOS + " / " + runtime.GOARCH,
		GoVersion:      runtime.Version(),
		CPUCores:       runtime.NumCPU(),
	}

	// Renderiza o HTML
	tmpl, err := template.ParseFiles("templates/index.html")
	if err != nil {
		http.Error(w, "Erro ao carregar o template HTML. Arquivo templates/index.html existe?", http.StatusInternalServerError)
		return
	}

	tmpl.Execute(w, data)

	// Registra métrica de sucesso (HTTP 200) e latência
	httpRequestsTotal.WithLabelValues("200").Inc()
	httpDuration.WithLabelValues("home").Observe(time.Since(start).Seconds())
}

func chaosHandler(w http.ResponseWriter, r *http.Request) {
	start := time.Now()

	// Simula um "Gargalo de Processamento" ou I/O travado
	delay := time.Duration(rand.Intn(3)+1) * time.Second
	fmt.Printf("🔥 [CAOS] Injetando latência de %v...\n", delay)
	time.Sleep(delay)

	// Simula um erro 500 aleatório (Bug de compatibilidade no código)
	if rand.Float32() > 0.5 {
		httpRequestsTotal.WithLabelValues("500").Inc()
		httpDuration.WithLabelValues("chaos").Observe(time.Since(start).Seconds())
		http.Error(w, "💥 BOOM! Erro 500 Injetado pelo Engenheiro do Caos.", http.StatusInternalServerError)
		return
	}

	httpRequestsTotal.WithLabelValues("200").Inc()
	httpDuration.WithLabelValues("chaos").Observe(time.Since(start).Seconds())
	fmt.Fprintf(w, "Latência de %v injetada com sucesso, mas sobrevivemos (HTTP 200). Volte para a home.", delay)
}
