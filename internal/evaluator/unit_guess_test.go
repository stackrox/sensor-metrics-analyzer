package evaluator

import "testing"

func TestGuessUnitFromMetricName(t *testing.T) {
	tests := []struct {
		name       string
		metricName string
		want       string
	}{
		{name: "seconds suffix", metricName: "http_request_duration_seconds", want: "seconds"},
		{name: "bytes suffix", metricName: "container_memory_usage_bytes", want: "bytes"},
		{name: "seconds total suffix", metricName: "process_cpu_seconds_total", want: "seconds"},
		{name: "timestamp seconds suffix", metricName: "last_seen_timestamp_seconds", want: "seconds"},
		{name: "real metric histogram bytes", metricName: "http_incoming_request_size_histogram_bytes", want: "bytes"},
		{name: "real metric bytes total", metricName: "go_memstats_alloc_bytes_total", want: "bytes"},
		{name: "real metric duration milliseconds", metricName: "rox_sensor_scan_call_duration_milliseconds", want: "milliseconds"},
		{name: "real metric purger duration seconds", metricName: "rox_sensor_network_flow_manager_purger_duration_seconds", want: "seconds"},
		{name: "real metric vm report duration milliseconds", metricName: "rox_sensor_virtual_machine_index_report_processing_duration_milliseconds", want: "milliseconds"},
		{name: "sec alias suffix", metricName: "request_duration_sec", want: "seconds"},
		{name: "msec alias suffix", metricName: "request_duration_msec", want: "milliseconds"},
		{name: "no known suffix", metricName: "rox_sensor_events", want: ""},
		{name: "empty metric name", metricName: "", want: ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := guessUnitFromMetricName(tt.metricName)
			if got != tt.want {
				t.Fatalf("guessUnitFromMetricName(%q) = %q, want %q", tt.metricName, got, tt.want)
			}
		})
	}
}

func TestGuessUnitFromHelpText(t *testing.T) {
	tests := []struct {
		name     string
		helpText string
		want     string
	}{
		{name: "milliseconds mention", helpText: "Time taken in milliseconds to process events", want: "milliseconds"},
		{name: "seconds mention", helpText: "Duration in seconds", want: "seconds"},
		{name: "seconds short sec mention", helpText: "Duration in sec", want: "seconds"},
		{name: "bytes mention", helpText: "Payload size in bytes", want: "bytes"},
		{name: "milliseconds alias msec", helpText: "Call took 32 msec", want: "milliseconds"},
		{name: "milliseconds alias millis", helpText: "Observed latency in millis", want: "milliseconds"},
		{name: "byte singular mention", helpText: "Equals to /memory/classes/total:byte.", want: "bytes"},
		{name: "ambiguous mentions", helpText: "Latency in milliseconds and seconds", want: ""},
		{name: "ambiguous time and bytes mentions", helpText: "Duration in seconds with payload bytes", want: ""},
		{name: "no known units", helpText: "Number of retries", want: ""},
		{name: "empty help text", helpText: "", want: ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := guessUnitFromHelpText(tt.helpText)
			if got != tt.want {
				t.Fatalf("guessUnitFromHelpText(%q) = %q, want %q", tt.helpText, got, tt.want)
			}
		})
	}
}

func TestGuessMetricUnit(t *testing.T) {
	tests := []struct {
		name       string
		metricName string
		helpText   string
		want       string
	}{
		{
			name:       "metric name wins over help text",
			metricName: "process_cpu_seconds_total",
			helpText:   "CPU usage in milliseconds",
			want:       "seconds",
		},
		{
			name:       "falls back to help text when name has no unit",
			metricName: "rox_central_event_processing",
			helpText:   "Time taken in milliseconds",
			want:       "milliseconds",
		},
		{
			name:       "real cluster duration without unit in name uses help",
			metricName: "rox_sensor_k8s_event_processing_duration",
			helpText:   "Time spent fully processing an event from Kubernetes in milliseconds",
			want:       "milliseconds",
		},
		{
			name:       "returns empty when both sources fail",
			metricName: "rox_sensor_events",
			helpText:   "Count of events",
			want:       "",
		},
		{
			name:       "real metric without unit in name or help",
			metricName: "rox_sensor_k8s_event_processing_duration",
			helpText:   "Time taken to fully process an event from Kubernetes",
			want:       "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := guessMetricUnit(tt.metricName, tt.helpText)
			if got != tt.want {
				t.Fatalf("guessMetricUnit(%q, %q) = %q, want %q", tt.metricName, tt.helpText, got, tt.want)
			}
		})
	}
}
