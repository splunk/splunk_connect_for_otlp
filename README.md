# OTLP Input Technical Addon

This repository contains a technical addon that exposes a OTLP endpoint for consumption of logs,
rendered as log lines to be ingested by Splunk Platform.

## Configuration

The input is configured as a data input in the Splunk Data Input settings.

You can set:
* The gRPC port and HTTP ports the OTLP receiver will listen on
* The listening address on which the OTLP input
* The format of the logs being emitted. The default emitted matches the `_otlpinput` sourcetype (see below).

### Sourcetype

The technical addon defines a sourcetype `_otlpinput` to be used with the default format.

It creates new records on detection of timestamps in RFC-3339 format.

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

