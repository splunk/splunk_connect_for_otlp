# OTLP Input Technical Addon

This repository contains a technical addon that exposes a OTLP endpoint for consumption of logs,
rendered as log lines to be ingested by Splunk Platform.

## Configuration

The input is configured as a data input in the Splunk Data Input settings.

You can set:
* The gRPC port and HTTP ports the OTLP receiver will listen on
* The listening address on which the OTLP input

## Build

Prerequisites:
* Go 1.23
* Make

Run:
```shell
$> make tgz
```

This will build the binaries, and assemble the TA tar.gz archive.

The archive is created as otlpinput.tgz.

## Testing

You can generate a payload using telemetrygen:
```shell
$> telemetrygen metrics --otlp-insecure --otlp-endpoint 0.0.0.0:4317 --metrics 100 --workers 10 --otlp-header Authorization=\"Splunk\ 000000-0000-00000-0000000000\"
$> telemetrygen logs --otlp-insecure --otlp-endpoint 0.0.0.0:4317 --logs 100 --workers 10 --otlp-header Authorization=\"Splunk\ 000000-0000-00000-0000000000\"
$> telemetrygen traces --otlp-insecure --otlp-endpoint 0.0.0.0:4317 --traces 100 --workers 10 --otlp-header Authorization=\"Splunk\ 000000-0000-00000-0000000000\"
```