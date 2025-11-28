# StackRox Sensor Metrics Analyzer

A standalone Go binary that analyzes StackRox Sensor Prometheus metrics using declarative TOML rule files.

## ✨ Features

- **🎮 Interactive TUI**: Beautiful terminal UI with keyboard navigation (powered by [Bubble Tea](https://github.com/charmbracelet/bubbletea))
- **📊 Load-Aware Analysis**: Automatically detects cluster load level (low/medium/high) and adjusts thresholds accordingly
- **🔗 Correlation Rules**: Rules can reference other metrics for intelligent status evaluation
- **🏷️ ACS Versioning**: Rules specify supported ACS versions and are filtered automatically
- **📝 Template-Based Reports**: Markdown reports generated from templates
- **🖥️ Console Output**: Default colorful console output with tables

## Installation

```bash
make build
```

## Usage

### Interactive TUI Mode (Recommended)

```bash
# Launch interactive terminal UI
./bin/metrics-analyzer analyze metrics.txt --format tui
```

**TUI Features:**
- Navigate results with `↑`/`↓` or `j`/`k` keys
- Press `Enter` to view detailed information
- Filter by status with `1-4` keys (All/Red/Yellow/Green)
- Search with `/` key
- Press `?` for help

### Console & Markdown Output

```bash
# Analyze metrics (console output - default)
./bin/metrics-analyzer analyze metrics.txt

# Analyze with custom rules directory
./bin/metrics-analyzer analyze metrics.txt --rules ./automated-rules

# Generate markdown report
./bin/metrics-analyzer analyze metrics.txt --format markdown --output report.md

# Override load level
./bin/metrics-analyzer analyze metrics.txt --load-level high

# Specify ACS version
./bin/metrics-analyzer analyze metrics.txt --acs-version 4.8
```

### Utility Commands

```bash
# Validate rules (defaults to current directory)
./bin/metrics-analyzer validate

# Validate rules in specific directory
./bin/metrics-analyzer validate ./automated-rules

# List all rules
./bin/metrics-analyzer list-rules
```

## TUI Keyboard Shortcuts

| Key | Action |
|-----|--------|
| `↑`/`k`, `↓`/`j` | Navigate up/down |
| `Enter`/`→` | View details |
| `←`/`Esc` | Go back |
| `g`/`Home` | Go to top |
| `G`/`End` | Go to bottom |
| `PgUp`/`PgDn` | Page up/down |
| `/` | Search/filter |
| `1-4` | Filter by status (All/Red/Yellow/Green) |
| `?` | Toggle help |
| `q` | Quit |

## Project Structure

```
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

## Testing

```bash
# Unit tests
make test

# Integration test (compare with Python output)
python3 analyze_metrics_full.py metrics.txt > /tmp/python-output.txt
./bin/metrics-analyzer analyze metrics.txt --format markdown --output /tmp/go-report.md
go run testdata/compare_outputs.go /tmp/python-output.txt /tmp/go-report.md
```

## Dependencies

Built with:
- [Bubble Tea](https://github.com/charmbracelet/bubbletea) - TUI framework
- [Lip Gloss](https://github.com/charmbracelet/lipgloss) - Style definitions
- [Bubbles](https://github.com/charmbracelet/bubbles) - TUI components
- [go-pretty](https://github.com/jedib0t/go-pretty) - Table formatting

## License

Apache 2.0 - See [LICENSE](LICENSE) for details.
