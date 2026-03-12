# Rule Types and Examples

Supported rule types are defined in `internal/rules/types.go` and evaluated in `internal/evaluator/`.

Examples below are intentionally simple and domain-neutral.

## 1) `gauge_threshold`

Think: **speed limit monitor** - for this example assume speed limit is 50 km/h.

- Below 51 km/h is GREEN
- 51-55 km/h is YELLOW
- Over 55 km/h is RED

```toml
rule_type = "gauge_threshold"
metric_name = "car_speed_kmh"
display_name = "car_speed_kmh"
description = "Current car speed"

[thresholds]
low = 51
high = 56
higher_is_worse = true

[messages]
green = "{value:.0f} km/h (within limit)"
yellow = "{value:.0f} km/h (slightly above limit)"
red = "{value:.0f} km/h (too fast)"
```

Example metric input:
```text
car_speed_kmh 54
```

Notes:
- With `higher_is_worse = true`: `< low => GREEN`, `< high => YELLOW`, else `RED`.
- With `higher_is_worse = false`, logic is inverted.

## 2) `percentage`

Think: **exam score** = correct answers / total answers.

```toml
rule_type = "percentage"
display_name = "exam_score"
description = "Percent of correct answers"

[percentage_config]
numerator = "answers_correct"
denominator = "answers_total"

[thresholds]
low = 70.0
high = 90.0
higher_is_worse = false

[messages]
green = "Score {value:.1f}% (great)"
yellow = "Score {value:.1f}% (okay)"
red = "Score {value:.1f}% (needs work)"
zero_activity = "No answers yet"
```

Example metric input:
```text
answers_correct 42
answers_total 50
```

Notes:
- Engine computes `(numerator / denominator) * 100`.
- If denominator is zero, it uses `zero_activity` (or a default fallback).

## 3) `queue_operations`

Think: **people entering vs leaving a store queue**.

```toml
rule_type = "queue_operations"
metric_name = "store_queue_operations_total"
display_name = "store_queue_balance"
description = "Queue growth over time"

[queue_config]
operation_label = "Operation"
add_value = "Enter"
remove_value = "Leave"

[thresholds]
low = 5
high = 20

[messages]
green = "Queue stable: Enter={add:.0f}, Leave={remove:.0f}, Diff={diff:.0f}"
yellow = "Queue growing: Diff={diff:.0f}"
red = "Queue overloaded: Diff={diff:.0f}"
```

Example metric input:
```text
store_queue_operations_total{Operation="Enter"} 120
store_queue_operations_total{Operation="Leave"} 115
```

Notes:
- Engine computes `diff = Add - Remove`.
- Thresholds are checked against `diff`.

## 4) `histogram`

Think: **delivery time distribution**.

```toml
rule_type = "histogram"
metric_name = "delivery_time_seconds"
display_name = "delivery_time_seconds"
description = "How long deliveries take"

[histogram_config]
unit = "seconds"

[thresholds]
p95_good = 30
p95_warn = 60

[messages]
green = "Delivery time healthy (p95={p95:.0f}s, p99={p99:.0f}s)"
yellow = "Delivery time elevated (p95={p95:.0f}s)"
red = "Delivery time too high (p95={p95:.0f}s)"
```

Example metric input:
```text
delivery_time_seconds_bucket{le="10"} 20
delivery_time_seconds_bucket{le="30"} 80
delivery_time_seconds_bucket{le="60"} 98
delivery_time_seconds_bucket{le="+Inf"} 100
```

Notes:
- Reads buckets from `<metric_name>_bucket`.
- Computes p50/p75/p95/p99 and evaluates based on p95.
- There is also a global automatic histogram `+Inf` overflow check - this is a separate rule built into the code.

## 5) `cache_hit_rate`

Think: **library lookup cache efficiency**.

```toml
rule_type = "cache_hit_rate"
display_name = "library_cache_hit_rate"
description = "Cache efficiency for book lookups"

[cache_config]
hits_metric = "cache_hits_total"
misses_metric = "cache_misses_total"

[thresholds]
low = 50.0
high = 80.0
higher_is_worse = false

[messages]
green = "{value:.1f}% hit rate (excellent)"
yellow = "{value:.1f}% hit rate (acceptable)"
red = "{value:.1f}% hit rate (poor)"
zero_activity = "No cache activity yet"
```

Example metric input:
```text
cache_hits_total 90
cache_misses_total 10
```

Notes:
- Hit rate is `(hits / (hits + misses)) * 100`.
- If both hits and misses are zero, `zero_activity` is used.

## 6) `composite`

Think: **coffee shop health check** using multiple metrics together.

```toml
rule_type = "composite"
display_name = "coffee_shop_health"
description = "Combined check of key shop metrics"

[composite_config]

[[composite_config.metrics]]
name = "orders"
source = "orders_total"

[[composite_config.metrics]]
name = "baristas"
source = "baristas_on_shift"

[[composite_config.metrics]]
name = "machines"
source = "working_machines"

[[composite_config.checks]]
check_type = "not_zero"
metrics = ["baristas", "machines"]
status = "red"
message = "No baristas or no machines available"

[[composite_config.checks]]
check_type = "ratio"
numerator = "orders"
denominator = "baristas"
min_ratio = 2.0
status = "yellow"
message = "Low orders per barista ratio"

[messages]
green = "Coffee shop looks healthy"
```

Example metric input:
```text
orders_total 40
baristas_on_shift 8
working_machines 3
```

Notes:
- Checks run in order; first matching check decides the status/message.
- Supported `check_type`: `not_zero`, `ratio`.

