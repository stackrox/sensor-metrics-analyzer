# Project Structure

```text
sensor-metrics-analyzer-go/
├── cmd/metrics-analyzer/    # CLI entry point
├── internal/
│   ├── parser/              # Prometheus metrics parser
│   ├── rules/               # TOML rule loader and validator
│   ├── loadlevel/           # Load level detection engine
│   ├── evaluator/           # Rule evaluation logic
│   ├── reporter/            # Report generation (markdown/console)
│   └── tui/                 # Interactive terminal UI (Bubble Tea)
├── automated-rules/         # TOML rule definitions
└── templates/               # Report templates
```

