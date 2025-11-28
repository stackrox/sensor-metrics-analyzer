# StackRox Sensor Metrics Analyzer

A standalone Go binary that analyzes StackRox Sensor Prometheus metrics using declarative TOML rule files.

## Features

- **Load-Aware Analysis**: Automatically detects cluster load level (low/medium/high) and adjusts thresholds accordingly
- **Correlation Rules**: Rules can reference other metrics for intelligent status evaluation
- **ACS Versioning**: Rules specify supported ACS versions and are filtered automatically
- **Template-Based Reports**: Markdown reports generated from templates
- **Console Output**: Default colorful console output with tables

## Installation

```bash
make build
```

## Usage

```bash
# Analyze metrics (console output)
./bin/metrics-analyzer analyze metrics.txt

# Analyze with custom rules directory
./bin/metrics-analyzer analyze metrics.txt --rules ./automated-rules

# Generate markdown report
./bin/metrics-analyzer analyze metrics.txt --format markdown --output report.md

# Override load level
./bin/metrics-analyzer analyze metrics.txt --load-level high

# Specify ACS version
./bin/metrics-analyzer analyze metrics.txt --acs-version 4.8

# Validate rules (defaults to current directory)
./bin/metrics-analyzer validate

# Validate rules in specific directory
./bin/metrics-analyzer validate ./automated-rules

# List all rules
./bin/metrics-analyzer list-rules
```

## Project Structure

- `cmd/metrics-analyzer/` - CLI entry point
- `internal/parser/` - Prometheus metrics parser
- `internal/rules/` - TOML rule loader and validator
- `internal/loadlevel/` - Load level detection engine
- `internal/evaluator/` - Rule evaluation logic
- `internal/reporter/` - Report generation (markdown/console)
- `automated-rules/` - TOML rule definitions
- `templates/` - Report templates

## Testing

```bash
# Unit tests
make test

# Integration test (compare with Python output)
python3 analyze_metrics_full.py metrics.txt > /tmp/python-output.txt
./bin/metrics-analyzer analyze metrics.txt --format markdown --output /tmp/go-report.md
go run testdata/compare_outputs.go /tmp/python-output.txt /tmp/go-report.md
```

