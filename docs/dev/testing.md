# Testing

```bash
# Unit tests
make test

# Integration test (deprecated)
# This compares the results with the output of the Python script, 
# which was historically the initial form of this entire project.
python3 analyze_metrics_full.py metrics.txt > /tmp/python-output.txt
./bin/metrics-analyzer analyze --format markdown --output /tmp/go-report.md metrics.txt
go run testdata/compare_outputs.go /tmp/python-output.txt /tmp/go-report.md
```

