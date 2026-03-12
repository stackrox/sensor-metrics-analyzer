# Advanced Rule Features

This page covers optional fields supported by the rules engine.

## 1) Load-Aware Thresholds

Rules can override thresholds per load level (`low`, `medium`, `high`).
If a field is missing in a load-level block, defaults from `[thresholds]` are used.

```toml
[thresholds]
low = 100
high = 1000
higher_is_worse = true

[load_level_thresholds.low]
low = 50
high = 500
higher_is_worse = true

[load_level_thresholds.high]
low = 200
high = 2000
higher_is_worse = true
```

Interpretation:
- Default thresholds are `low = 100` and `high = 1000` (`100/1000`).
- If detector says load is `low`, the rule becomes stricter (`50/500`).
- If detector says load is `high`, the rule becomes more tolerant (`200/2000`).
- If `medium` is not defined, it falls back to default thresholds.

Quick example (queue drops):
- Imagine the metric is `dropped_events = 100`.
- Under `low` load (thresholds `50/500`): `100` is between low and high -> `YELLOW` (unusual under light traffic).
- Under `high` load (thresholds `200/2000`): `100 < 200` -> `GREEN` (acceptable under heavy traffic).

Another way to tune this:
- If your team wants any drops under low load to be more severe, you can set low-load thresholds tighter (for example `low = 1`, `high = 20`), so `100` becomes `RED` at low load.

## 2) Correlation Rules

A rule can be suppressed or elevated based on other metrics.
It is similar to the load-awareness, but here the parameter influencing the
evaluation thresholds can be any other metric.

```toml
[correlation]

[[correlation.suppress_if]]
metric_name = "some_related_metric"
operator = "lt"     # gt, lt, eq, gte, lte
value = 10
status = "GREEN"    # optional explicit status override

[[correlation.elevate_if]]
metric_name = "error_rate"
operator = "gt"
value = 5
status = "RED"
```

Interpretation:
- First, the rule gets a base status from its own thresholds.
- Then correlation checks can adjust that status:
  - `suppress_if` lowers severity (or forces `status` if set).
  - `elevate_if` raises severity (or forces `status` if set).

Concrete example (home thermostat):
- Primary rule watches `room_temperature`.
- If temperature is too high, base status is `RED`.
- Suppress rule: if `window_is_open = 1`, set status to `GREEN` (open window explains temporary temperature change).
- Elevate rule: if `smoke_detected = 1`, set status to `RED` (always critical).

Metric snapshot example:
- `room_temperature = 35` (base `RED` as 35 Celcius is way too much in a room)
- `window_is_open = 1` (`suppress_if` matches -> `GREEN`)
- `smoke_detected = 1` (`elevate_if` matches -> `RED`)
- Final result: `RED`

Behavior summary:
- `suppress_if`: downgrades severity (`RED -> YELLOW -> GREEN`) unless `status` is set.
- `elevate_if`: upgrades severity (`GREEN -> YELLOW -> RED`) unless `status` is set.

## 3) ACS Version Constraints

You can scope rules by ACS version using one of these methods:

```toml
acs_versions = ["4.9+", "4.10+"]
```

or:

```toml
min_acs_version = "4.9.0"
max_acs_version = "4.10.99"
```

Interpretation:
- Version constraints decide whether a rule is evaluated at all.
- If target ACS version does not match, the rule is skipped.
- This is useful when metric behavior changes between ACS releases.

Quick examples:
- `acs_versions = ["4.9+"]` matches `4.9.2` and `4.10.0`, but not `4.8.7`.
- `min_acs_version = "4.9.0"` and `max_acs_version = "4.10.99"` matches `4.9.x` and `4.10.x`.

Supported `acs_versions` patterns:
- exact: `4.9`
- plus: `4.9+`
- range: `4.9-4.10`
- comparator: `>=4.9`

Pattern examples:

```toml
# exact: only one version train (major.minor)
acs_versions = ["4.9"]
```
- Matches: `4.9.0`, `4.9.2`
- Does not match: `4.10.0`

```toml
# plus: this minor and all newer minors
acs_versions = ["4.9+"]
```
- Matches: `4.9.0`, `4.10.1`, `4.11.0`
- Does not match: `4.8.9`

```toml
# range: bounded set of minors
acs_versions = ["4.9-4.10"]
```
- Matches: `4.9.x`, `4.10.x`
- Does not match: `4.8.x`, `4.11.x`

```toml
# comparator: same idea as plus, but explicit operator syntax
acs_versions = [">=4.9"]
```
- Matches: `4.9.0` and newer
- Does not match: `4.8.x`

When to use `4.9+` vs `>=4.9`:
- Use `4.9+` when you want concise, reader-friendly config for "4.9 and newer".
- Use `>=4.9` when your team prefers explicit comparator style (often easier to scan when mixed with other comparator forms).
- Functionally, they are equivalent in this analyzer.

## 4) Review and Remediation Metadata

Why this matters:
- This project started as an AI hackathon tool, and the first wave of rules was AI-generated.
- Over time, many rules were reviewed and improved by humans, which generally increases reliability.
- Some older AI-generated rules may still exist, so reports should make rule provenance visible.
- The review metadata helps users understand confidence level during evaluation (for example, fully reviewed vs not yet reviewed).

```toml
reviewed = "Yes, human-reviewed"
last_review_by = "Alice"
last_review_on = "2026-01-30"

[remediation]
yellow = "Monitor trend for the next hour."
red = "Investigate immediately and check downstream symptoms."
```

Interpretation:
- `reviewed`, `last_review_by`, and `last_review_on` show confidence and ownership.
- `remediation` gives human-facing actions per status.
- These fields do not change status calculation; they improve report usability.

Quick example:
- Rule evaluates to `YELLOW` -> report includes a hint what to do next: `Monitor trend for the next hour.`
- Rule evaluates to `RED` -> report includes hint: `Investigate X and Y immediately...`

