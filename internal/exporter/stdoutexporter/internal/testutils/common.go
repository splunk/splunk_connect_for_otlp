package testutils

// Constants for Splunk components.
const (
	DefaultSourceTypeLabel     = "com.splunk.sourcetype"
	DefaultSourceLabel         = "com.splunk.source"
	DefaultIndexLabel          = "com.splunk.index"
	DefaultNameLabel           = "otel.log.name"
	DefaultSeverityTextLabel   = "otel.log.severity.text"
	DefaultSeverityNumberLabel = "otel.log.severity.number"
	HECTokenHeader             = "Splunk"
	HTTPSplunkChannelHeader    = "X-Splunk-Request-Channel"

	HecTokenLabel = "com.splunk.hec.access_token" // #nosec
	// HecEventMetricType is the type of HEC event. Set to metric, as per https://docs.splunk.com/Documentation/Splunk/8.0.3/Metrics/GetMetricsInOther.
	HecEventMetricType = "metric"
	DefaultRawPath     = "/services/collector/raw"
	DefaultHealthPath  = "/services/collector/health"
	DefaultAckPath     = "/services/collector/ack"

	// https://docs.splunk.com/Documentation/Splunk/9.2.1/Metrics/Overview#What_is_a_metric_data_point.3F
	// metric name can contain letters, numbers, underscore, dot or colon. cannot start with number or underscore, or contain metric_name
	metricNamePattern = `^metric_name:([A-Za-z.:][A-Za-z0-9_.:\\-]*)$`
)

func BuildHTTPHeaders() map[string]string {
	return map[string]string{
		"Connection":           "keep-alive",
		"Content-Type":         "application/json",
		"User-Agent":           "SplunkAppName/vx.x.x",
		"Authorization":        HECTokenHeader + " " + "fake-token",
		"__splunk_app_name":    "SplunkAppName",
		"__splunk_app_version": "SplunkAppVersion",
	}
}
