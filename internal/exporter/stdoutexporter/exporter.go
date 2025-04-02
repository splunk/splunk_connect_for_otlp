// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package stdoutexporter

import (
	"context"
	"fmt"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/exporter"
	"go.opentelemetry.io/collector/pdata/plog"
)

func newLogsExporter(_ context.Context, _ exporter.Settings, _ component.Config) (exporter.Logs, error) {
	return stdoutExporter{}, nil
}

type stdoutExporter struct {
}

func (s stdoutExporter) Start(ctx context.Context, host component.Host) error {
	return nil
}

func (s stdoutExporter) Shutdown(ctx context.Context) error {
	return nil
}

func (s stdoutExporter) Capabilities() consumer.Capabilities {
	return consumer.Capabilities{
		MutatesData: false,
	}
}

func (s stdoutExporter) ConsumeLogs(_ context.Context, ld plog.Logs) error {
	for i := 0; i < ld.ResourceLogs().Len(); i++ {
		rl := ld.ResourceLogs().At(i)
		for j := 0; j < ld.ResourceLogs().Len(); j++ {
			sl := rl.ScopeLogs().At(j)
			for k := 0; k < sl.LogRecords().Len(); k++ {
				logRecord := sl.LogRecords().At(k)
				fmt.Println(logRecord.Body().AsString())
			}
		}
	}
	return nil
}
