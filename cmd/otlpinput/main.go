// Copyright Splunk Inc. 2025
// SPDX-License-Identifier: Apache-2.0

// Program otlpinput is a binary listening for OTLP data and exporting it to stdout.
package main

import (
	"context"
	"fmt"
	"github.com/splunk/otlpinput/internal"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/exporter"
	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/receiver"
	"go.opentelemetry.io/collector/receiver/otlpreceiver"
	noopmetric "go.opentelemetry.io/otel/metric/noop"
	"go.opentelemetry.io/otel/trace/noop"
	"log"
	"os"

	"github.com/splunk/otlpinput/internal/exporter/stdoutexporter"
)

func main() {
	defer func() {
		if r := recover(); r != nil {
			log.Fatal(r)
		}
	}()
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "--scheme":
			fmt.Println(internal.Scheme)
		case "--validate-arguments":

		}
	} else if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	config, err := internal.ReadFromStdin()
	if err != nil {
		return err
	}

	logger, err := internal.CreateLogger()
	if err != nil {
		return err
	}

	settings := component.TelemetrySettings{
		Logger:         logger,
		TracerProvider: noop.NewTracerProvider(),
		MeterProvider:  noopmetric.NewMeterProvider(),
		Resource:       pcommon.NewResource(),
	}

	format, grpcPort, httpPort, listeningAddress := config.Extract()

	stdoutCfg := stdoutexporter.NewFactory().CreateDefaultConfig().(*stdoutexporter.Config)
	stdoutCfg.Template = format
	ctx := context.Background()
	e, err := stdoutexporter.NewFactory().CreateLogs(ctx, exporter.Settings{
		TelemetrySettings: settings,
		ID:                component.MustNewID("stdout"),
	}, stdoutCfg)
	if err != nil {
		return err
	}

	cfg := otlpreceiver.NewFactory().CreateDefaultConfig().(*otlpreceiver.Config)
	cfg.GRPC.NetAddr.Endpoint = fmt.Sprintf("%s:%d", listeningAddress, grpcPort)
	cfg.HTTP.ServerConfig.Endpoint = fmt.Sprintf("%s:%d", listeningAddress, httpPort)

	r, err := otlpreceiver.NewFactory().CreateLogs(ctx, receiver.Settings{
		TelemetrySettings: settings,
		ID:                component.MustNewID("otlp"),
	}, cfg, e)

	h := &internal.TTYHost{
		ErrStatus: make(chan error, 1),
	}
	h.Start()

	if err = e.Start(ctx, h); err != nil {
		return err
	}
	if err = r.Start(ctx, h); err != nil {
		return err
	}

	err = h.Wait()

	_ = r.Shutdown(ctx)
	_ = e.Shutdown(ctx)

	return err
}
