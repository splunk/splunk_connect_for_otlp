// Copyright Splunk Inc. 2025
// SPDX-License-Identifier: Apache-2.0

package stdoutexporter

import (
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/exporter"
)

// This file implements factory for stdout exporter.

const (
	typeStr        = "stdout"
	stabilityLevel = component.StabilityLevelDevelopment
)

// NewFactory creates a factory for stdin receiver.
func NewFactory() exporter.Factory {
	return exporter.NewFactory(
		component.MustNewType(typeStr),
		createDefaultConfig,
		exporter.WithLogs(newLogsExporter, stabilityLevel))
}

// CreateDefaultConfig creates the default configuration for stdin receiver.
func createDefaultConfig() component.Config {
	return &Config{}
}
