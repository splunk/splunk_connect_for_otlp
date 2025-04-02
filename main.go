// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

// Program otlpstdout is a binary listening for OTLP data and exporting it to stdout.
package main

import (
	"context"
	"log"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/component/componentstatus"
	"go.opentelemetry.io/collector/exporter"
	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/receiver"
	"go.opentelemetry.io/collector/receiver/otlpreceiver"
	noopmetric "go.opentelemetry.io/otel/metric/noop"
	"go.opentelemetry.io/otel/sdk/trace"
	"go.uber.org/zap"

	"github.com/otel-warez/otlpstdout/internal/exporter/stdoutexporter"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	h := ttyHost{
		errStatus: make(chan error, 1),
	}

	logger := zap.NewNop()
	settings := component.TelemetrySettings{
		Logger:         logger,
		TracerProvider: trace.NewTracerProvider(),
		MeterProvider:  noopmetric.NewMeterProvider(),
		Resource:       pcommon.NewResource(),
	}

	ctx := context.Background()

	e, err := stdoutexporter.NewFactory().CreateLogs(ctx, exporter.Settings{
		TelemetrySettings: settings,
		ID:                component.MustNewID("stdout"),
	}, stdoutexporter.NewFactory().CreateDefaultConfig())
	if err != nil {
		return err
	}

	cfg := otlpreceiver.NewFactory().CreateDefaultConfig().(*otlpreceiver.Config)
	cfg.GRPC.NetAddr.Endpoint = "0.0.0.0:4317"
	cfg.HTTP.ServerConfig.Endpoint = "0.0.0.0:4318"

	r, err := otlpreceiver.NewFactory().CreateLogs(ctx, receiver.Settings{
		TelemetrySettings: settings,
		ID:                component.MustNewID("otlp"),
	}, cfg, e)

	err = e.Start(ctx, h)
	if err != nil {
		return err
	}
	err = r.Start(ctx, h)
	if err != nil {
		return err
	}

	err = <-h.errStatus

	_ = r.Shutdown(ctx)
	_ = e.Shutdown(ctx)

	return err
}

var _ component.Host = ttyHost{}
var _ componentstatus.Reporter = ttyHost{}

type ttyHost struct {
	errStatus chan error
}

func (t ttyHost) Report(event *componentstatus.Event) {
	if event.Status() == componentstatus.StatusStopping {
		close(t.errStatus)
	}
	if event.Err() != nil {
		t.errStatus <- event.Err()
	}
}

func (t ttyHost) GetExtensions() map[component.ID]component.Component {
	return nil
}
