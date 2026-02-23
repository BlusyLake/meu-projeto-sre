package main

import (
	"context"
	"fmt"
	"html/template"
	"io"
	"log"
	"math/rand"
	"net/http"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/redis/go-redis/v9"
)

var (
	httpRequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total de requisições HTTP processadas.",
		},
		[]string{"status", "container_id"},
	)
	httpDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "Latência das requisições HTTP.",
			Buckets: []float64{0.1, 0.5, 1, 2, 5},
		},
		[]string{"handler", "container_id"},
	)
)

func init() {
	prometheus.MustRegister(httpRequestsTotal)
	prometheus.MustRegister(httpDuration)
}

type PageData struct {
	Hostname       string
	Visits         string
	RedisStatus    string
	RedisIsDown    bool
	LastPingTime   string
	OSArchitecture string
	GoVersion      string
	CPUCores       int
	PipelineCode   string // NOVO
	AppAuthor      string // NOVO
}

var rdb *redis.Client

func main() {
	redisAddr := os.Getenv("REDIS_ADDR")
	if redisAddr == "" {
		redisAddr = "localhost:6379"
	}

	rdb = redis.NewClient(&redis.Options{
		Addr:        redisAddr,
		Password:    "",
		DB:          0,
		DialTimeout: 50 * time.Millisecond,
		ReadTimeout: 50 * time.Millisecond,
	})

	http.HandleFunc("/", homeHandler)
	http.HandleFunc("/chaos", chaosHandler)
	http.HandleFunc("/visual-dashboard", dashboardHandler)
	// proxy endpoint to allow browser JS to talk to Prometheus without CORS or path-prefix problems
	http.HandleFunc("/prometheus/", prometheusProxy)
	http.Handle("/metrics", promhttp.Handler())

	porta := "3000"
	fmt.Printf("🚀 [SRE App] Subindo na porta %s...\n", porta)
	log.Fatal(http.ListenAndServe(":"+porta, nil))
}

func dashboardHandler(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("templates/dashboard.html")
	if err != nil {
		http.Error(w, "Erro ao carregar o dashboard HTML.", http.StatusInternalServerError)
		return
	}
	tmpl.Execute(w, nil)
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	hostname, _ := os.Hostname()
	ctx := context.Background()

	visits := "0"
	redisStatus := "Online"
	redisIsDown := false

	val, err := rdb.Incr(ctx, "contador_visitas").Result()
	if err != nil {
		redisStatus = "Offline (Graceful Degradation Ativo)"
		visits = "Erro Cache"
		redisIsDown = true
	} else {
		visits = fmt.Sprintf("%d", val)
	}

	// Lendo a Pipeline do disco
	pipelineBytes, err := os.ReadFile("pipeline.yml")
	pipelineCode := "Código da pipeline em processamento..."
	if err == nil {
		pipelineCode = string(pipelineBytes)
	}

	// Lendo o Autor do Build
	author := os.Getenv("APP_AUTHOR")
	if author == "" {
		author = "Edi (Pipeline Local)"
	}

	data := PageData{
		Hostname:       hostname,
		Visits:         visits,
		RedisStatus:    redisStatus,
		RedisIsDown:    redisIsDown,
		LastPingTime:   time.Now().Format("15:04:05"),
		OSArchitecture: runtime.GOOS + " / " + runtime.GOARCH,
		GoVersion:      runtime.Version(),
		CPUCores:       runtime.NumCPU(),
		PipelineCode:   pipelineCode,
		AppAuthor:      author,
	}

	tmpl, _ := template.ParseFiles("templates/index.html")
	tmpl.Execute(w, data)

	httpRequestsTotal.WithLabelValues("200", hostname).Inc()
	httpDuration.WithLabelValues("home", hostname).Observe(time.Since(start).Seconds())
}

func chaosHandler(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	hostname, _ := os.Hostname()
	delay := time.Duration(rand.Intn(3)+1) * time.Second
	time.Sleep(delay)

	if rand.Float32() > 0.5 {
		httpRequestsTotal.WithLabelValues("500", hostname).Inc()
		httpDuration.WithLabelValues("chaos", hostname).Observe(time.Since(start).Seconds())
		http.Error(w, "💥 BOOM! Erro 500.", http.StatusInternalServerError)
		return
	}

	httpRequestsTotal.WithLabelValues("200", hostname).Inc()
	httpDuration.WithLabelValues("chaos", hostname).Observe(time.Since(start).Seconds())
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, `{"status": "success", "delay_injetado": "%v"}`, delay)
}

// prometheusProxy forwards any request from /prometheus/* to the actual Prometheus instance.
// The target address is taken from PROM_ADDR env var (default prom/prometheus container service).
func prometheusProxy(w http.ResponseWriter, r *http.Request) {
	target := os.Getenv("PROM_ADDR")
	if target == "" {
		// inside compose use service name
		target = "http://prometheus:9090"
	}
	// construct proxied URL
	proxied := target + strings.TrimPrefix(r.URL.Path, "/prometheus")
	if r.URL.RawQuery != "" {
		proxied += "?" + r.URL.RawQuery
	}
	resp, err := http.Get(proxied)
	if err != nil {
		http.Error(w, "Proxy error: "+err.Error(), http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()
	// copy status and headers
	for k, vals := range resp.Header {
		for _, v := range vals {
			w.Header().Add(k, v)
		}
	}
	w.WriteHeader(resp.StatusCode)
	io.Copy(w, resp.Body)
}
