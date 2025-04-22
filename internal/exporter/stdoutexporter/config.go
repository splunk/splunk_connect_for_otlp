// Copyright Splunk Inc. 2025
// SPDX-License-Identifier: Apache-2.0

package stdoutexporter

import "go.opentelemetry.io/collector/exporter/exporterhelper"

type Config struct {
	Template         string                          `mapstructure:"template"`
	QueueBatchConfig exporterhelper.QueueBatchConfig `mapstructure:"batch_config"`
}
