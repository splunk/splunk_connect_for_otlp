// Copyright Splunk Inc. 2025
// SPDX-License-Identifier: Apache-2.0

package stdoutexporter

import (
	"context"
	"errors"
	"fmt"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/exporter"
	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/plog"
	"os"
	"text/template"
	"time"
)

func newLogsExporter(_ context.Context, _ exporter.Settings, cfg component.Config) (exporter.Logs, error) {
	oCfg := cfg.(*Config)
	tmpl := template.New("stdout")
	tmpl.Funcs(map[string]any{
		"epoch": func(t pcommon.Timestamp) int64 {
			return t.AsTime().Unix()
		},
		"iso8601": func(t pcommon.Timestamp) string {
			return t.AsTime().Format(time.RFC3339)
		},
		"mapToString": func(m pcommon.Map) string {
			str := ""
			for k, v := range m.All() {
				str += fmt.Sprintf("%s:%s", k, v.AsString())
			}
			return str
		},
	})

	tmpl, err := tmpl.Parse(oCfg.Template)
	if err != nil {
		return nil, err
	}

	return &stdoutExporter{
		format: tmpl,
	}, nil
}

type stdoutExporter struct {
	format *template.Template
}

func (s *stdoutExporter) Start(_ context.Context, _ component.Host) error {
	return nil
}

func (s *stdoutExporter) Shutdown(_ context.Context) error {
	return nil
}

func (s *stdoutExporter) Capabilities() consumer.Capabilities {
	return consumer.Capabilities{
		MutatesData: false,
	}
}

type logData struct {
	Resource  pcommon.Resource
	LogRecord plog.LogRecord
}

func (s *stdoutExporter) ConsumeLogs(_ context.Context, ld plog.Logs) error {
	var errs []error
	for i := 0; i < ld.ResourceLogs().Len(); i++ {
		rl := ld.ResourceLogs().At(i)
		r := rl.Resource()
		for j := 0; j < ld.ResourceLogs().Len(); j++ {
			sl := rl.ScopeLogs().At(j)
			for k := 0; k < sl.LogRecords().Len(); k++ {
				logRecord := sl.LogRecords().At(k)
				err := s.format.Execute(os.Stdout, logData{LogRecord: logRecord, Resource: r})
				if err != nil {
					errs = append(errs, err)
				}
				_, _ = os.Stdout.Write([]byte{'\n'})
			}
		}
	}
	return errors.Join(errs...)
}
