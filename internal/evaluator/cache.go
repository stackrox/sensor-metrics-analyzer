package evaluator

import (
	"fmt"
	"time"

	"github.com/stackrox/sensor-metrics-analyzer/internal/parser"
	"github.com/stackrox/sensor-metrics-analyzer/internal/rules"
)

// EvaluateCacheHit evaluates a cache hit rate rule
func EvaluateCacheHit(rule rules.Rule, metrics parser.MetricsData, loadLevel rules.LoadLevel) rules.EvaluationResult {
	result := rules.EvaluationResult{
		RuleName:  rule.DisplayName,
		Status:    rules.StatusGreen,
		Details:   []string{},
		Timestamp: time.Now(),
	}

	if rule.CacheConfig == nil {
		result.Message = "Cache config not specified"
		return result
	}

	// Get hits metric
	hitsMetric, exists := metrics.GetMetric(rule.CacheConfig.HitsMetric)
	if !exists || len(hitsMetric.Values) == 0 {
		result.Message = fmt.Sprintf("Hits metric %s not found", rule.CacheConfig.HitsMetric)
		return result
	}

	// Get misses metric
	missesMetric, exists := metrics.GetMetric(rule.CacheConfig.MissesMetric)
	if !exists || len(missesMetric.Values) == 0 {
		result.Message = fmt.Sprintf("Misses metric %s not found", rule.CacheConfig.MissesMetric)
		return result
	}

	hits, _ := hitsMetric.GetSingleValue()
	misses, _ := missesMetric.GetSingleValue()
	total := hits + misses

	if total == 0 {
		result.Status = rules.StatusGreen
		result.Message = rule.Messages.ZeroActivity
		if result.Message == "" {
			result.Message = "No cache activity yet (0 hits, 0 misses)"
		}
		return result
	}

	hitRate := (hits / total) * 100
	result.Value = hitRate

	// Select thresholds based on load level
	thresholds := selectThresholds(rule, loadLevel)

	// For cache hit rate, higher is better (higher_is_worse = false)
	if thresholds.HigherIsWorse {
		// If misconfigured, treat as higher is better
		if hitRate >= thresholds.High {
			result.Status = rules.StatusGreen
		} else if hitRate >= thresholds.Low {
			result.Status = rules.StatusYellow
		} else {
			result.Status = rules.StatusRed
		}
	} else {
		// Higher is better
		if hitRate >= thresholds.High {
			result.Status = rules.StatusGreen
		} else if hitRate >= thresholds.Low {
			result.Status = rules.StatusYellow
		} else {
			result.Status = rules.StatusRed
		}
	}

	result.Message = interpolate(rule.Messages.Green, hitRate, map[string]interface{}{
		"hits":   hits,
		"misses": misses,
	})
	switch result.Status {
	case rules.StatusYellow:
		result.Message = interpolate(rule.Messages.Yellow, hitRate, map[string]interface{}{
			"hits":   hits,
			"misses": misses,
		})
	case rules.StatusRed:
		result.Message = interpolate(rule.Messages.Red, hitRate, map[string]interface{}{
			"hits":   hits,
			"misses": misses,
		})
	}

	return result
}
