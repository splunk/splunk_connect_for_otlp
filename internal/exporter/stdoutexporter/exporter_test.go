// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package stdoutexporter

import (
	"context"
	"encoding/hex"
	"fmt"
	"go.opentelemetry.io/collector/component/componenttest"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"go.opentelemetry.io/collector/exporter/exportertest"
	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/plog"
	"go.opentelemetry.io/collector/pdata/pmetric"
	"go.opentelemetry.io/collector/pdata/ptrace"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest"

	"github.com/splunk/otlpinput/internal/exporter/stdoutexporter/internal/testutils"
)

type splunkContainerConfig struct {
	conCtx    context.Context
	container testcontainers.Container
}

func setup() splunkContainerConfig {
	// Perform setup operations here
	cfg := startSplunk()
	return cfg
}

func teardown(cfg splunkContainerConfig) {
	// Perform teardown operations here
	fmt.Println("Tearing down...")
	// Stop and remove the container
	fmt.Println("Stopping container")
	err := cfg.container.Terminate(cfg.conCtx)
	if err != nil {
		fmt.Printf("Error while terminating container")
		panic(err)
	}
	// Remove docker image after tests
	splunkImage := testutils.GetConfigVariable("SPLUNK_IMAGE")
	cmd := exec.Command("docker", "rmi", splunkImage)

	// Execute command
	output, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Printf("Error removing Docker image: %v\n", err)
	}
	fmt.Printf("Removed Docker image: %s\n", splunkImage)
	fmt.Printf("Command output:\n%s\n", output)
}

func startSplunk() splunkContainerConfig {
	logger, err := zap.NewProduction()
	if err != nil {
		panic(err)
	}

	conContext := context.Background()

	// Create a new container
	splunkImage := testutils.GetConfigVariable("SPLUNK_IMAGE")
	req := testcontainers.ContainerRequest{
		Image:        splunkImage,
		ExposedPorts: []string{"8000/tcp", "8088/tcp", "8089/tcp"},
		Env: map[string]string{
			"SPLUNK_START_ARGS": "--accept-license",
			"SPLUNK_HEC_TOKEN":  testutils.GetConfigVariable("HEC_TOKEN"),
			"SPLUNK_PASSWORD":   testutils.GetConfigVariable("PASSWORD"),
		},
		Files: []testcontainers.ContainerFile{
			{
				HostFilePath:      filepath.Join("testdata", "splunk.yaml"),
				ContainerFilePath: "/tmp/defaults/default.yml",
				FileMode:          0o644,
			},
		},
		WaitingFor: wait.ForHealthCheck().WithStartupTimeout(5 * time.Minute),
	}

	container, err := testcontainers.GenericContainer(conContext, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		logger.Info("Error while creating container")
		panic(err)
	}

	// Get the container host and port
	uiPort, err := container.MappedPort(conContext, "8000")
	if err != nil {
		logger.Info("Error while getting port")
		panic(err)
	}

	hecPort, err := container.MappedPort(conContext, "8088")
	if err != nil {
		logger.Info("Error while getting port")
		panic(err)
	}
	managementPort, err := container.MappedPort(conContext, "8089")
	if err != nil {
		logger.Info("Error while getting port")
		panic(err)
	}
	host, err := container.Host(conContext)
	if err != nil {
		logger.Info("Error while getting host")
		panic(err)
	}

	// Use the container's host and port for your tests
	logger.Info("Splunk running at:", zap.String("host", host), zap.Int("uiPort", uiPort.Int()), zap.Int("hecPort", hecPort.Int()), zap.Int("managementPort", managementPort.Int()))
	testutils.SetConfigVariable("HOST", host)
	testutils.SetConfigVariable("UI_PORT", strconv.Itoa(uiPort.Int()))
	testutils.SetConfigVariable("HEC_PORT", strconv.Itoa(hecPort.Int()))
	testutils.SetConfigVariable("MANAGEMENT_PORT", strconv.Itoa(managementPort.Int()))
	cfg := splunkContainerConfig{
		conCtx:    conContext,
		container: container,
	}
	return cfg
}

func prepareLogs() plog.Logs {
	logs := plog.NewLogs()
	rl := logs.ResourceLogs().AppendEmpty()
	sl := rl.ScopeLogs().AppendEmpty()
	sl.Scope().SetName("test")
	ts := pcommon.Timestamp(0)
	logRecord := sl.LogRecords().AppendEmpty()
	logRecord.Body().SetStr("test log")
	logRecord.Attributes().PutStr(testutils.DefaultNameLabel, "test- label")
	logRecord.Attributes().PutStr("host.name", "myhost")
	logRecord.Attributes().PutStr("custom", "custom")
	logRecord.SetTimestamp(ts)
	return logs
}

func prepareLogsNonDefaultParams(index, source, sourcetype, event string) plog.Logs {
	logs := plog.NewLogs()
	rl := logs.ResourceLogs().AppendEmpty()
	sl := rl.ScopeLogs().AppendEmpty()
	sl.Scope().SetName("test")
	ts := pcommon.Timestamp(0)

	logRecord := sl.LogRecords().AppendEmpty()
	logRecord.Body().SetStr(event)
	logRecord.Attributes().PutStr(testutils.DefaultNameLabel, "label")
	logRecord.Attributes().PutStr(testutils.DefaultSourceLabel, source)
	logRecord.Attributes().PutStr(testutils.DefaultSourceTypeLabel, sourcetype)
	logRecord.Attributes().PutStr(testutils.DefaultIndexLabel, index)
	logRecord.Attributes().PutStr("host.name", "myhost")
	logRecord.Attributes().PutStr("custom", "custom")
	logRecord.SetTimestamp(ts)
	return logs
}

func prepareMetricsData(metricName string) pmetric.Metrics {
	metricData := pmetric.NewMetrics()
	metric := metricData.ResourceMetrics().AppendEmpty().ScopeMetrics().AppendEmpty().Metrics().AppendEmpty()
	g := metric.SetEmptyGauge()
	g.DataPoints().AppendEmpty().SetDoubleValue(132.929)
	metric.SetName(metricName)
	return metricData
}

func prepareTracesData(index, source, sourcetype string) ptrace.Traces {
	ts := pcommon.Timestamp(0)

	traces := ptrace.NewTraces()
	rs := traces.ResourceSpans().AppendEmpty()
	rs.Resource().Attributes().PutStr(testutils.DefaultSourceLabel, source)
	rs.Resource().Attributes().PutStr("host.name", "myhost")
	rs.Resource().Attributes().PutStr(testutils.DefaultSourceTypeLabel, sourcetype)
	rs.Resource().Attributes().PutStr(testutils.DefaultIndexLabel, index)
	ils := rs.ScopeSpans().AppendEmpty()
	initSpan("myspan", ts, ils.Spans().AppendEmpty())
	return traces
}

type cfg struct {
	event      string
	index      string
	source     string
	sourcetype string
}

type telemetryType string

var (
	metricsType = telemetryType("metrics")
	logsType    = telemetryType("logs")
	tracesType  = telemetryType("traces")
)

type testCfg struct {
	name      string
	config    *cfg
	startTime string
	telType   telemetryType
}

func captureOutput(f func(context.Context, plog.Logs) error, ctx context.Context, logs plog.Logs) (string, error) {
	orig := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	err := f(ctx, logs)
	os.Stdout = orig
	w.Close()
	out, _ := io.ReadAll(r)
	return string(out), err
}

func logsTest(t *testing.T, test testCfg) {
	settings := exportertest.NewNopSettings(exportertest.NopType)
	var logs plog.Logs
	if test.config.index != "main" {
		logs = prepareLogsNonDefaultParams(test.config.index, test.config.source, test.config.sourcetype, test.config.event)
	} else {
		logs = prepareLogs()
	}

	exporter, err := newLogsExporter(t.Context(), settings, createDefaultConfig())
	require.NoError(t, err)
	err = exporter.Start(t.Context(), componenttest.NewNopHost())
	require.NoError(t, err)
	out, err := captureOutput(exporter.ConsumeLogs, t.Context(), logs)
	require.NotEmpty(t, out)
	t.Log("Captured output:", out)
	require.NoError(t, err, "Must not error while sending Logs data")
	waitForEventToBeIndexed()

	events := testutils.CheckEventsFromSplunk("index="+test.config.index+" *", test.startTime)
	assert.Len(t, events, 1)
	// check events fields
	data, ok := events[0].(map[string]any)
	assert.True(t, ok, "Invalid event format")
	assert.Equal(t, test.config.event, data["_raw"].(string))
	assert.Equal(t, test.config.index, data["index"].(string))
	assert.Equal(t, test.config.source, data["source"].(string))
	assert.Equal(t, test.config.sourcetype, data["sourcetype"].(string))
}

func metricsTest(t *testing.T, test testCfg) {
	settings := exportertest.NewNopSettings(exportertest.NopType)
	metricData := prepareMetricsData(test.config.event)

	exporter, err := newMetricsExporter(t.Context(), settings, createDefaultConfig())
	require.NoError(t, err)
	err = exporter.Start(t.Context(), componenttest.NewNopHost())
	require.NoError(t, err)
	err = exporter.ConsumeMetrics(t.Context(), metricData)
	require.NoError(t, err, "Must not error while sending Metrics data")
	waitForEventToBeIndexed()

	events := testutils.CheckMetricsFromSplunk(test.config.index, test.config.event)
	assert.Len(t, events, 1, "Events length is less than 1. No metrics found")
}

func tracesTest(t *testing.T, test testCfg) {
	settings := exportertest.NewNopSettings(exportertest.NopType)
	tracesData := prepareTracesData(test.config.index, test.config.source, test.config.sourcetype)

	exporter, err := newTracesExporter(t.Context(), settings, createDefaultConfig())
	require.NoError(t, err)
	err = exporter.Start(t.Context(), componenttest.NewNopHost())
	require.NoError(t, err)
	err = exporter.ConsumeTraces(t.Context(), tracesData)
	require.NoError(t, err, "Must not error while sending Traces data")
	require.NoError(t, err, "Must not error while sending Trace data")
	waitForEventToBeIndexed()

	events := testutils.CheckEventsFromSplunk("index="+test.config.index+" *", test.startTime)
	assert.Len(t, events, 1)
	// check fields
	data, ok := events[0].(map[string]any)
	assert.True(t, ok, "Invalid event format")
	assert.Equal(t, test.config.index, data["index"].(string))
	assert.Equal(t, test.config.source, data["source"].(string))
	assert.Equal(t, test.config.sourcetype, data["sourcetype"].(string))
}

func TestSplunkHecExporter(t *testing.T) {
	splunkContCfg := setup()
	defer teardown(splunkContCfg)

	tests := []testCfg{
		{
			name: "Events to Splunk",
			config: &cfg{
				event:      "test log",
				index:      "main",
				source:     "otel",
				sourcetype: "st-otel",
			},
			startTime: "-3h@h",
			telType:   logsType,
		},
		{
			name: "Events to Splunk - Non default index",
			config: &cfg{
				event:      "This is my new event! And some number 101",
				index:      testutils.GetConfigVariable("EVENT_INDEX"),
				source:     "otel-source",
				sourcetype: "sck-otel-st",
			},
			startTime: "-1m@m",
			telType:   logsType,
		},
		{
			name: "Events to Splunk - metrics",
			config: &cfg{
				event:      "test.metric",
				index:      testutils.GetConfigVariable("METRIC_INDEX"),
				source:     "otel",
				sourcetype: "st-otel",
			},
			startTime: "",
			telType:   metricsType,
		},
		{
			name: "Events to Splunk - traces",
			config: &cfg{
				event:      "",
				index:      testutils.GetConfigVariable("TRACE_INDEX"),
				source:     "trace-source",
				sourcetype: "trace-sourcetype",
			},
			startTime: "-1m@m",
			telType:   tracesType,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			logger := zaptest.NewLogger(t)
			logger.Info("Test -> Splunk running at:", zap.String("host", testutils.GetConfigVariable("HOST")),
				zap.String("uiPort", testutils.GetConfigVariable("UI_PORT")),
				zap.String("hecPort", testutils.GetConfigVariable("HEC_PORT")),
				zap.String("managementPort", testutils.GetConfigVariable("MANAGEMENT_PORT")),
			)

			switch test.telType {
			case logsType:
				logsTest(t, test)
			case metricsType:
				metricsTest(t, test)
			case tracesType:
				tracesTest(t, test)
			default:
				assert.Fail(t, "Telemetry type must be set to one of the following values: metrics, traces, or logs.")
			}
		})
	}
}

func waitForEventToBeIndexed() {
	time.Sleep(3 * time.Second)
}

func initSpan(name string, ts pcommon.Timestamp, span ptrace.Span) {
	span.Attributes().PutStr("foo", "bar")
	span.SetName(name)
	span.SetStartTimestamp(ts)
	spanLink := span.Links().AppendEmpty()
	spanLink.TraceState().FromRaw("OK")
	bytes, _ := hex.DecodeString("12345678")
	var traceID [16]byte
	copy(traceID[:], bytes)
	spanLink.SetTraceID(traceID)
	bytes, _ = hex.DecodeString("1234")
	var spanID [8]byte
	copy(spanID[:], bytes)
	spanLink.SetSpanID(spanID)
	spanLink.Attributes().PutInt("foo", 1)
	spanLink.Attributes().PutBool("bar", false)
	foobarContents := spanLink.Attributes().PutEmptySlice("foobar")
	foobarContents.AppendEmpty().SetStr("a")
	foobarContents.AppendEmpty().SetStr("b")

	spanEvent := span.Events().AppendEmpty()
	spanEvent.Attributes().PutStr("foo", "bar")
	spanEvent.SetName("myEvent")
	spanEvent.SetTimestamp(ts + 3)
}
