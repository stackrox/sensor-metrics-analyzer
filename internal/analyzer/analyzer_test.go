package analyzer

import (
	"bytes"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAnalyzeFile(t *testing.T) {
	t.Parallel()

	_, thisFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("AnalyzeFile() failed to resolve test file path")
	}
	repoRoot := filepath.Dir(filepath.Dir(filepath.Dir(thisFile)))
	metricsFile := filepath.Join(repoRoot, "testdata", "fixtures", "sample_metrics.txt")
	rulesDir := filepath.Join(repoRoot, "automated-rules")

	var logs bytes.Buffer
	report, err := AnalyzeFile(metricsFile, Options{
		RulesDir: rulesDir,
		Logger:   &logs,
	})
	assert.NoError(t, err)

	assert.NotEmpty(t, report.ClusterName, "AnalyzeFile() cluster name is empty")
	assert.False(t, report.Timestamp.IsZero(), "AnalyzeFile() timestamp is zero")
	assert.NotEmpty(t, report.LoadLevel, "AnalyzeFile() load level is empty")
	assert.NotEmpty(t, report.Results, "AnalyzeFile() returned no results")
	assert.Equal(t, report.Summary.TotalAnalyzed, len(report.Results), "AnalyzeFile() summary mismatch")
	statusTotal := report.Summary.RedCount + report.Summary.YellowCount + report.Summary.GreenCount
	assert.LessOrEqual(t, statusTotal, report.Summary.TotalAnalyzed, "AnalyzeFile() summary counts exceed total")
}
