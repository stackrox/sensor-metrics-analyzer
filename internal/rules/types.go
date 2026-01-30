package rules

import "time"

// RuleType represents the type of analysis to perform
type RuleType string

const (
	RuleTypeGauge         RuleType = "gauge_threshold"
	RuleTypePercentage    RuleType = "percentage"
	RuleTypeQueue         RuleType = "queue_operations"
	RuleTypeHistogram     RuleType = "histogram"
	RuleTypeCacheHit      RuleType = "cache_hit_rate"
	RuleTypeComposite     RuleType = "composite"
	RuleTypeLoadDetection RuleType = "load_detection"
)

// Status represents the health status of a metric
type Status string

const (
	StatusGreen  Status = "GREEN"
	StatusYellow Status = "YELLOW"
	StatusRed    Status = "RED"
)

// LoadLevel represents the detected cluster load level
type LoadLevel string

const (
	LoadLevelLow    LoadLevel = "low"
	LoadLevelMedium LoadLevel = "medium"
	LoadLevelHigh   LoadLevel = "high"
)

// Thresholds contains threshold values for evaluation
type Thresholds struct {
	Low           float64 `toml:"low"`
	High          float64 `toml:"high"`
	HigherIsWorse bool    `toml:"higher_is_worse"`
	P95Good       float64 `toml:"p95_good"`
	P95Warn       float64 `toml:"p95_warn"`
	MinRatio      float64 `toml:"min_ratio"`
}

// LoadLevelThresholds contains thresholds for each load level
type LoadLevelThresholds struct {
	Low    *Thresholds `toml:"low"`
	Medium *Thresholds `toml:"medium"`
	High   *Thresholds `toml:"high"`
}

// Messages contains status message templates
type Messages struct {
	Green        string `toml:"green"`
	Yellow       string `toml:"yellow"`
	Red          string `toml:"red"`
	ZeroActivity string `toml:"zero_activity"`
}

// Remediation contains suggested actions for each status level
type Remediation struct {
	Red    string `toml:"red"`    // Suggested action when status is RED
	Yellow string `toml:"yellow"` // Suggested action when status is YELLOW (optional)
	Green  string `toml:"green"`  // Informational message when status is GREEN (optional)
}

// CorrelationConfig allows rules to reference other metrics
type CorrelationConfig struct {
	// If referenced metrics meet conditions, suppress or modify this rule's status
	SuppressIf []CorrelationCondition `toml:"suppress_if"`

	// If referenced metrics meet conditions, elevate this rule's status
	ElevateIf []CorrelationCondition `toml:"elevate_if"`
}

// CorrelationCondition defines a condition for correlation evaluation
type CorrelationCondition struct {
	MetricName string  `toml:"metric_name"`
	Operator   string  `toml:"operator"` // "gt", "lt", "eq", "gte", "lte"
	Value      float64 `toml:"value"`
	Status     Status  `toml:"status"` // Status to apply if condition met
}

// Rule represents a loaded TOML rule
type Rule struct {
	RuleType    RuleType `toml:"rule_type"`
	MetricName  string   `toml:"metric_name"`
	DisplayName string   `toml:"display_name"`
	Description string   `toml:"description"`

	// Review metadata (optional)
	Reviewed     string `toml:"reviewed"`
	LastReviewBy string `toml:"last_review_by"`
	LastReviewOn string `toml:"last_review_on"`

	// Type-specific configurations
	GaugeConfig      *GaugeConfig      `toml:"gauge_config"`
	PercentageConfig *PercentageConfig `toml:"percentage_config"`
	QueueConfig      *QueueConfig      `toml:"queue_config"`
	HistogramConfig  *HistogramConfig  `toml:"histogram_config"`
	CacheConfig      *CacheConfig      `toml:"cache_config"`
	CompositeConfig  *CompositeConfig  `toml:"composite_config"`

	Thresholds  Thresholds   `toml:"thresholds"`
	Messages    Messages     `toml:"messages"`
	Remediation *Remediation `toml:"remediation"` // Optional remediation actions

	// Load-aware thresholds (optional, falls back to thresholds if not set)
	LoadLevelThresholds *LoadLevelThresholds `toml:"load_level_thresholds"`

	// Correlation: reference other metrics for conditional evaluation
	Correlation *CorrelationConfig `toml:"correlation"`

	// ACS version support
	ACSVersions []string `toml:"acs_versions"` // e.g., ["4.7+", "4.8+", "4.9+"]

	// Minimum ACS version (inclusive)
	MinACSVersion string `toml:"min_acs_version"`

	// Maximum ACS version (inclusive)
	MaxACSVersion string `toml:"max_acs_version"`
}

// GaugeConfig for simple threshold-based gauge metrics
type GaugeConfig struct {
	// All in Thresholds
}

// PercentageConfig for ratio/percentage calculations
type PercentageConfig struct {
	Numerator   string `toml:"numerator"`
	Denominator string `toml:"denominator"`
}

// QueueConfig for queue operation balance
type QueueConfig struct {
	OperationLabel string `toml:"operation_label"`
	AddValue       string `toml:"add_value"`
	RemoveValue    string `toml:"remove_value"`
}

// HistogramConfig for latency/duration histograms
type HistogramConfig struct {
	Unit string `toml:"unit"`
}

// CacheConfig for cache hit rate calculation
type CacheConfig struct {
	HitsMetric   string `toml:"hits_metric"`
	MissesMetric string `toml:"misses_metric"`
}

// CompositeConfig for multi-metric rules
type CompositeConfig struct {
	Metrics []CompositeMetric `toml:"metrics"`
	Checks  []CompositeCheck  `toml:"checks"`
}

type CompositeMetric struct {
	Name   string `toml:"name"`
	Source string `toml:"source"`
}

type CompositeCheck struct {
	CheckType   string   `toml:"check_type"`
	Metrics     []string `toml:"metrics"`
	Numerator   string   `toml:"numerator"`
	Denominator string   `toml:"denominator"`
	MinRatio    float64  `toml:"min_ratio"`
	Status      string   `toml:"status"`
	Message     string   `toml:"message"`
}

// LoadDetectionRule represents a load detection rule
type LoadDetectionRule struct {
	RuleType    RuleType                 `toml:"rule_type"`
	DisplayName string                   `toml:"display_name"`
	Metrics     []LoadDetectionMetric    `toml:"metrics"`
	Thresholds  []LoadDetectionThreshold `toml:"thresholds"`
}

type LoadDetectionMetric struct {
	Name   string  `toml:"name"`
	Source string  `toml:"source"`
	Weight float64 `toml:"weight"`
}

type LoadDetectionThreshold struct {
	Level    LoadLevel `toml:"level"`
	MinValue float64   `toml:"min_value"`
	MaxValue float64   `toml:"max_value"`
}

// EvaluationResult represents the result of evaluating a rule
type EvaluationResult struct {
	RuleName                 string
	Status                   Status
	Message                  string
	Value                    float64
	Details                  []string
	ReviewStatus             string
	Remediation              string // Legacy field (use PotentialActionUser/Developer)
	PotentialActionUser      string
	PotentialActionDeveloper string
	Timestamp                time.Time
}

// AnalysisReport contains all evaluation results
type AnalysisReport struct {
	ClusterName string
	ACSVersion  string    // Detected or user-specified
	LoadLevel   LoadLevel // Detected or user-specified
	Timestamp   time.Time
	Results     []EvaluationResult
	Summary     Summary
}

// Summary contains aggregate statistics
type Summary struct {
	TotalAnalyzed int
	RedCount      int
	YellowCount   int
	GreenCount    int
}
