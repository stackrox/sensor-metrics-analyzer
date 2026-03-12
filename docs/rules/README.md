# Rules Wiki

This project evaluates Prometheus metrics using declarative TOML rules from `automated-rules/`.

Use this wiki to understand what rule types exist and how to write new rules.

## Start Here

- [Rule Types and Examples](./rule-types.md)
- [Advanced Rule Features](./advanced-features.md)
- [Load Detection Rules](./load-detection.md)

## Minimal Rule Skeleton

```toml
rule_type = "gauge_threshold"
metric_name = "my_metric_name"
display_name = "my_metric_name"
description = "What this metric means"

[thresholds]
low = 100
high = 1000
higher_is_worse = true

[messages]
green = "{value:.0f} healthy"
yellow = "{value:.0f} elevated"
red = "{value:.0f} critical"
```

## Validate Rules

```bash
# Build CLI
make build

# Validate all rules
./bin/metrics-analyzer validate ./automated-rules
```

## Quick Test With Metrics

```bash
./bin/metrics-analyzer analyze --rules ./automated-rules testdata/fixtures/sample_metrics.txt
```

