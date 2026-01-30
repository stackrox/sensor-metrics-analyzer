package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/stackrox/sensor-metrics-analyzer/internal/analyzer"
	"github.com/stackrox/sensor-metrics-analyzer/internal/reporter"
)

const (
	defaultListenAddr     = ":8080"
	defaultMaxFileSize    = 50 * 1024 * 1024 // 50MB
	defaultRequestTimeout = 60 * time.Second
	defaultRulesDir       = "./automated-rules"
	defaultLoadLevelDir   = "./automated-rules/load-level"
)

type Config struct {
	ListenAddr     string
	MaxFileSize    int64
	RequestTimeout time.Duration
	RulesDir       string
	LoadLevelDir   string
	TemplatePath   string
}

type AnalyzeResponse struct {
	Markdown string `json:"markdown"`
	Console  string `json:"console"`
	Error    string `json:"error,omitempty"`
}

type VersionResponse struct {
	Version    string `json:"version"`
	LastUpdate string `json:"lastUpdate"`
}

var (
	buildVersion = "dev"
	buildTime    = ""
)

func main() {
	cfg := parseFlags()

	log.Printf("Starting server on %s", cfg.ListenAddr)
	log.Printf("Rules directory: %s", cfg.RulesDir)
	log.Printf("Load level directory: %s", cfg.LoadLevelDir)
	log.Printf("Max file size: %d bytes", cfg.MaxFileSize)

	http.HandleFunc("/api/analyze/both", handleAnalyzeBoth(cfg))
	http.HandleFunc("/health", handleHealth)
	http.HandleFunc("/version", handleVersion())

	if err := http.ListenAndServe(cfg.ListenAddr, nil); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}

func parseFlags() *Config {
	cfg := &Config{
		ListenAddr:     defaultListenAddr,
		MaxFileSize:    defaultMaxFileSize,
		RequestTimeout: defaultRequestTimeout,
		RulesDir:       defaultRulesDir,
		LoadLevelDir:   defaultLoadLevelDir,
		TemplatePath:   "./templates/markdown.tmpl",
	}

	flag.StringVar(&cfg.ListenAddr, "listen", defaultListenAddr, "Listen address")
	flag.Int64Var(&cfg.MaxFileSize, "max-size", defaultMaxFileSize, "Max upload file size (bytes)")
	flag.DurationVar(&cfg.RequestTimeout, "timeout", defaultRequestTimeout, "Request timeout")
	flag.StringVar(&cfg.RulesDir, "rules", defaultRulesDir, "Rules directory")
	flag.StringVar(&cfg.LoadLevelDir, "load-level-dir", defaultLoadLevelDir, "Load level rules directory")
	flag.StringVar(&cfg.TemplatePath, "template", cfg.TemplatePath, "Path to markdown template")

	flag.Parse()

	// Override with environment variables if set
	if envAddr := os.Getenv("LISTEN_ADDR"); envAddr != "" {
		cfg.ListenAddr = envAddr
	}
	if envSize := os.Getenv("MAX_FILE_SIZE"); envSize != "" {
		var size int64
		if _, err := fmt.Sscanf(envSize, "%d", &size); err == nil {
			cfg.MaxFileSize = size
		}
	}
	if envRules := os.Getenv("RULES_DIR"); envRules != "" {
		cfg.RulesDir = envRules
	}
	if envLoadLevel := os.Getenv("LOAD_LEVEL_DIR"); envLoadLevel != "" {
		cfg.LoadLevelDir = envLoadLevel
	}
	if envTimeout := os.Getenv("REQUEST_TIMEOUT"); envTimeout != "" {
		if parsed, err := time.ParseDuration(envTimeout); err == nil {
			cfg.RequestTimeout = parsed
		}
	}
	if envTemplate := os.Getenv("TEMPLATE_PATH"); envTemplate != "" {
		cfg.TemplatePath = envTemplate
	}

	return cfg
}

func handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func handleVersion() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		version := "Unknown"
		lastUpdate := "Unknown"

		if buildVersion != "" {
			version = buildVersion
		}
		if buildTime != "" {
			lastUpdate = buildTime
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(VersionResponse{
			Version:    version,
			LastUpdate: lastUpdate,
		})
	}
}

func handleAnalyzeBoth(cfg *Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		ctx, cancel := context.WithTimeout(r.Context(), cfg.RequestTimeout)
		defer cancel()
		if err := ctx.Err(); err != nil {
			respondError(w, http.StatusRequestTimeout, "Request timed out")
			return
		}

		// Set max file size
		r.Body = http.MaxBytesReader(w, r.Body, cfg.MaxFileSize)

		// Parse multipart form
		if err := r.ParseMultipartForm(cfg.MaxFileSize); err != nil {
			respondError(w, http.StatusBadRequest, fmt.Sprintf("Failed to parse form: %v", err))
			return
		}

		// Get uploaded file
		file, header, err := r.FormFile("file")
		if err != nil {
			respondError(w, http.StatusBadRequest, fmt.Sprintf("No file uploaded: %v", err))
			return
		}
		defer file.Close()

		log.Printf("Processing upload (%d bytes)", header.Size)

		response := AnalyzeResponse{}
		report, err := analyzer.AnalyzeReader(file, analyzer.Options{
			RulesDir:     cfg.RulesDir,
			LoadLevelDir: cfg.LoadLevelDir,
			ClusterName:  analyzer.ExtractClusterName(header.Filename),
			Logger:       log.New(os.Stdout, "analyzer: ", log.LstdFlags).Writer(),
		})
		if err := ctx.Err(); err != nil {
			respondError(w, http.StatusRequestTimeout, "Request timed out")
			return
		}
		if err != nil {
			response.Error = fmt.Sprintf("Analysis failed: %v", err)
		} else {
			response.Console = reporter.GenerateConsole(report)
			markdown, mdErr := reporter.GenerateMarkdown(report, cfg.TemplatePath)
			if mdErr != nil {
				response.Error = fmt.Sprintf("Markdown generation failed: %v", mdErr)
			} else {
				response.Markdown = markdown
			}
		}

		// Return response
		w.Header().Set("Content-Type", "application/json")
		if response.Error != "" && response.Console == "" && response.Markdown == "" {
			w.WriteHeader(http.StatusInternalServerError)
		}
		json.NewEncoder(w).Encode(response)
	}
}

func respondError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(AnalyzeResponse{Error: message})
}
