# Load Detection Rules

Load detection rules live in `automated-rules/load-level/` and define how cluster load is classified as:
- `low`
- `medium`
- `high`

This load level is then used by normal rules that define `load_level_thresholds`.

## Example

```toml
rule_type = "load_detection"
display_name = "cluster_volume"

[[metrics]]
name = "containers"
source = "rox_sensor_num_containers_in_clusterentities_store"
weight = 1.0

[[metrics]]
name = "pods"
source = "rox_sensor_num_pods_in_store"
weight = 0.5

[[thresholds]]
level = "low"
max_value = 100

[[thresholds]]
level = "medium"
min_value = 100
max_value = 500

[[thresholds]]
level = "high"
min_value = 500
```

## Worked Example (Step by Step)

Assume the current metrics are:

```text
rox_sensor_num_containers_in_clusterentities_store 240
rox_sensor_num_pods_in_store 120
```

Using the weights from the rule:

- `containers` weight = `1.0`
- `pods` weight = `0.5`

Now compute:

1. `weighted_sum = (240 * 1.0) + (120 * 0.5) = 240 + 60 = 300`
2. `total_weight = 1.0 + 0.5 = 1.5`
3. `normalized_value = weighted_sum / total_weight = 300 / 1.5 = 200`

Compare `normalized_value = 200` to thresholds:

- `low`: `< 100`
- `medium`: `>= 100` and `< 500`
- `high`: `>= 500`

Result: **`medium` load level**.

## How Score Calculation Works

The detector computes:

1. `weighted_sum = Σ(metric_value * weight)`
2. `total_weight = Σ(weight)`
3. `normalized_value = weighted_sum / total_weight`
4. Match `normalized_value` against threshold ranges

## Practical Guidance

- Start with 1-3 metrics that represent cluster size/pressure.
- Use larger weights for more trusted indicators.
- Keep thresholds simple first, then tune from production observations.
- Re-test normal rules after load detection changes, because rule statuses may shift with new load levels.

