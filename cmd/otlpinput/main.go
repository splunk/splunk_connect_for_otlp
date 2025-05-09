// Copyright Splunk Inc. 2025
// SPDX-License-Identifier: Apache-2.0

// Program otlpinput is a binary listening for OTLP data and exporting it to stdout.
package main

import (
	"context"
	"fmt"
	"github.com/open-telemetry/opentelemetry-collector-contrib/exporter/splunkhecexporter"
	"github.com/open-telemetry/opentelemetry-collector-contrib/extension/headerssetterextension"
	"github.com/splunk/otlpinput/internal"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/config/configauth"
	"go.opentelemetry.io/collector/exporter"
	"go.opentelemetry.io/collector/extension"
	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/receiver"
	"go.opentelemetry.io/collector/receiver/otlpreceiver"
	noopmetric "go.opentelemetry.io/otel/metric/noop"
	"go.opentelemetry.io/otel/trace/noop"
	"log"
	"os"
	"time"
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
	logger.Info("Starting OTLP input")

	settings := component.TelemetrySettings{
		Logger:         logger,
		TracerProvider: noop.NewTracerProvider(),
		MeterProvider:  noopmetric.NewMeterProvider(),
		Resource:       pcommon.NewResource(),
	}

	extFactory := headerssetterextension.NewFactory()
	extCfg := extFactory.CreateDefaultConfig().(*headerssetterextension.Config)
	authorization := "Authorization"
	extCfg.HeadersConfig = []headerssetterextension.HeaderConfig{
		{
			Action:      "upsert",
			Key:         &authorization,
			FromContext: &authorization,
		},
	}
	headerExtension, err := extFactory.Create(context.Background(), extension.Settings{
		ID:                component.MustNewID("headers_setter"),
		TelemetrySettings: settings,
	}, extCfg)
	if err != nil {
		return err
	}

	grpcPort, httpPort, listeningAddress := config.Extract()
	f := splunkhecexporter.NewFactory()
	hecCfg := f.CreateDefaultConfig().(*splunkhecexporter.Config)
	hecCfg.Endpoint = "http://localhost:8088/services/collector/event"
	hecCfg.Timeout = 10 * time.Second
	hecCfg.BackOffConfig.Enabled = false
	hecCfg.QueueSettings.Enabled = false
	hecCfg.Auth = &configauth.Authentication{AuthenticatorID: component.MustNewID("headers_setter")}
	ctx := context.Background()
	telemetrySettings := exporter.Settings{
		TelemetrySettings: settings,
		ID:                component.MustNewID("splunk_hec"),
	}
	le, err := f.CreateLogs(ctx, telemetrySettings, hecCfg)
	if err != nil {
		return err
	}
	me, err := f.CreateMetrics(ctx, telemetrySettings, hecCfg)
	if err != nil {
		return err
	}
	te, err := f.CreateTraces(ctx, telemetrySettings, hecCfg)
	if err != nil {
		return err
	}

	rf := otlpreceiver.NewFactory()
	cfg := rf.CreateDefaultConfig().(*otlpreceiver.Config)
	cfg.GRPC.NetAddr.Endpoint = fmt.Sprintf("%s:%d", listeningAddress, grpcPort)
	cfg.HTTP.ServerConfig.Endpoint = fmt.Sprintf("%s:%d", listeningAddress, httpPort)
	cfg.GRPC.IncludeMetadata = true
	cfg.HTTP.ServerConfig.IncludeMetadata = true

	if _, err = rf.CreateLogs(ctx, receiver.Settings{
		TelemetrySettings: settings,
		ID:                component.MustNewID("otlp"),
	}, cfg, le); err != nil {
		return err
	}
	if _, err = rf.CreateMetrics(ctx, receiver.Settings{
		TelemetrySettings: settings,
		ID:                component.MustNewID("otlp"),
	}, cfg, me); err != nil {
		return err
	}
	r, err := rf.CreateTraces(ctx, receiver.Settings{
		TelemetrySettings: settings,
		ID:                component.MustNewID("otlp"),
	}, cfg, te)
	if err != nil {
		return err
	}

	h := &internal.TTYHost{
		ErrStatus: make(chan error, 1),
		Extensions: map[component.ID]component.Component{
			component.MustNewID("headers_setter"): headerExtension,
		},
	}
	h.Start()

	if err = le.Start(ctx, h); err != nil {
		return err
	}
	if err = me.Start(ctx, h); err != nil {
		return err
	}
	if err = te.Start(ctx, h); err != nil {
		return err
	}
	if err = r.Start(ctx, h); err != nil {
		return err
	}

	logger.Info("OTLP Input started")

	err = h.Wait()

	_ = r.Shutdown(ctx)
	_ = le.Shutdown(ctx)
	_ = te.Shutdown(ctx)
	_ = me.Shutdown(ctx)

	return err
}
