package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/splunk/otlp2splunk/internal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMainPrintsScheme(t *testing.T) {
	output := captureStdout(t, func() {
		originalArgs := os.Args
		os.Args = []string{"splunk-connect-for-otlp", "--scheme"}
		defer func() {
			os.Args = originalArgs
		}()

		main()
	})

	require.Equal(t, internal.Scheme+"\n", output)
}

func TestRunReturnsErrorForInvalidInput(t *testing.T) {
	restoreStdin := writeToStdin(t, "not-xml")
	defer restoreStdin()

	err := run()
	require.Error(t, err)
}

func TestRunStartsAndStopsOnSignal(t *testing.T) {
	restoreStdin := writeToStdin(t, `<input><configuration><stanza name="splunk-connect-for-otlp://test" app="search"><param name="grpc_port">0</param><param name="http_port">0</param><param name="listen_address">127.0.0.1</param></stanza></configuration></input>`)
	defer restoreStdin()

	done := make(chan error, 1)
	go func() {
		done <- run()
	}()

	time.AfterFunc(500*time.Millisecond, func() {
		_ = syscall.Kill(os.Getpid(), syscall.SIGTERM)
	})

	select {
	case err := <-done:
		require.NoError(t, err)
	case <-time.After(5 * time.Second):
		t.Fatal("run did not complete in time")
	}
}

func TestExpectedHEC(t *testing.T) {
	tests := []struct {
		name         string
		otlpendpoint string
		inputPath    string
		expectedPath string
	}{

		{
			name:         "metrics",
			otlpendpoint: "/v1/metrics",
			inputPath:    filepath.Join("testdata", "otlp_metrics.json"),
			expectedPath: filepath.Join("testdata", "expected_hec_metrics.json"),
		},
		{
			name:         "traces",
			otlpendpoint: "/v1/traces",
			inputPath:    filepath.Join("testdata", "otlp_traces.json"),
			expectedPath: filepath.Join("testdata", "expected_hec_traces.json"),
		},
		{
			name:         "logs",
			otlpendpoint: "/v1/logs",
			inputPath:    filepath.Join("testdata", "otlp_logs.json"),
			expectedPath: filepath.Join("testdata", "expected_hec_logs.json"),
		},
		{
			name:         "large file metrics",
			otlpendpoint: "/v1/metrics",
			inputPath:    filepath.Join("testdata", "otlp_metrics_big.json"),
			expectedPath: filepath.Join("testdata", "expected_hec_metrics_big.json"),
		},
		{
			name:         "large file traces",
			otlpendpoint: "/v1/traces",
			inputPath:    filepath.Join("testdata", "otlp_traces_big.json"),
			expectedPath: filepath.Join("testdata", "expected_hec_traces_big.json"),
		},
		{
			name:         "large file logs",
			otlpendpoint: "/v1/logs",
			inputPath:    filepath.Join("testdata", "otlp_logs_big.json"),
			expectedPath: filepath.Join("testdata", "expected_hec_logs_big.json"),
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			grpcPort := getFreePort(t)
			httpPort := getFreePort(t)

			config := fmt.Sprintf(`<input><configuration><stanza name="splunk-connect-for-otlp://test" app="search"><param name="grpc_port">%d</param><param name="http_port">%d</param><param name="listen_address">127.0.0.1</param></stanza></configuration></input>`, grpcPort, httpPort)

			restoreStdin := writeToStdin(t, config)
			t.Cleanup(restoreStdin)

			stdoutLines, restoreStdout := captureStdoutLines(t)
			t.Cleanup(restoreStdout)

			runDone := make(chan error, 1)
			go func() {
				runDone <- run()
			}()

			payload, err := os.ReadFile(tt.inputPath)
			require.NoError(t, err)

			expected := loadExpectedHecData(t, tt.expectedPath)
			expectedLines := strings.Split(strings.TrimSpace(string(expected)), "\n")
			require.NotEmpty(t, expectedLines, "%s must contain fixture data", tt.expectedPath)

			postOTLP(t, httpPort, tt.otlpendpoint, payload)

			actual := collectLines(t, stdoutLines, len(expectedLines))
			require.Equal(t, expectedLines, actual)

			require.NoError(t, syscall.Kill(os.Getpid(), syscall.SIGTERM))
			require.NoError(t, <-runDone)
		})
	}
}

func writeToStdin(t *testing.T, content string) func() {
	t.Helper()

	r, w, err := os.Pipe()
	require.NoError(t, err)

	_, err = io.WriteString(w, content)
	require.NoError(t, err)
	require.NoError(t, w.Close())

	original := os.Stdin
	os.Stdin = r

	return func() {
		os.Stdin = original
		_ = r.Close()
	}
}

func captureStdout(t *testing.T, fn func()) string {
	t.Helper()

	original := os.Stdout
	r, w, err := os.Pipe()
	require.NoError(t, err)
	os.Stdout = w

	outputCh := make(chan string, 1)
	go func() {
		var buf bytes.Buffer
		_, _ = io.Copy(&buf, r)
		_ = r.Close()
		outputCh <- buf.String()
	}()

	fn()

	_ = w.Close()
	os.Stdout = original

	return <-outputCh
}

func captureStdoutLines(t *testing.T) (<-chan string, func()) {
	t.Helper()

	original := os.Stdout
	r, w, err := os.Pipe()
	require.NoError(t, err)
	os.Stdout = w

	lines := make(chan string, 50)
	go func() {
		scanner := bufio.NewScanner(r)
		buf := make([]byte, 0, 64*1024)
		scanner.Buffer(buf, 10*1024*1024)
		for scanner.Scan() {
			lines <- scanner.Text()
		}
		close(lines)
		_ = r.Close()
	}()

	return lines, func() {
		os.Stdout = original
		_ = w.Close()
	}
}

func getFreePort(t *testing.T) int {
	t.Helper()

	l, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	defer l.Close()
	return l.Addr().(*net.TCPAddr).Port
}

func postOTLP(t *testing.T, port int, path string, body []byte) {
	t.Helper()

	url := fmt.Sprintf("http://127.0.0.1:%d%s", port, path)

	lastRespCode := 0
	assert.Eventually(t, func() bool {
		resp, err := http.Post(url, "application/json", bytes.NewReader(body))
		if err != nil {
			return false
		}
		defer resp.Body.Close()
		lastRespCode = resp.StatusCode
		return resp.StatusCode == http.StatusOK
	}, 5*time.Second, 100*time.Millisecond, "failed to POST %s, response code: %v", path, lastRespCode)
}

func collectLines(t *testing.T, ch <-chan string, expectedCount int) []string {
	t.Helper()

	var lines []string
	timeout := time.After(10 * time.Second)
	for len(lines) < expectedCount {
		select {
		case line, ok := <-ch:
			if !ok {
				t.Fatalf("stdout closed early, got %d lines, expected %d", len(lines), expectedCount)
			}
			lines = append(lines, line)
		case <-timeout:
			t.Fatalf("timed out waiting for stdout lines; got %d expected %d", len(lines), expectedCount)
		}
	}
	return lines
}

func loadExpectedHecData(t *testing.T, path string) []byte {
	t.Helper()

	data, err := os.ReadFile(filepath.Clean(path))
	require.NoError(t, err)
	return data
}
