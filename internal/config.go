// Copyright Splunk Inc. 2025
// SPDX-License-Identifier: Apache-2.0

package internal

import (
	"bufio"
	"encoding/xml"
	"os"
	"strconv"
)

const (
	DefaultFormat        = "{{ .LogRecord.Timestamp | iso8601 }} {{.LogRecord.Body.AsString }} {{ .LogRecord.Attributes | mapToString }} {{ .Resource.Attributes | mapToString }}"
	DefaultGrpcPort      = 4317
	DefaultHTTPPort      = 4318
	DefaultListenAddress = "0.0.0.0"
)

type XMLInput struct {
	Configuration XMLConfig `xml:"configuration"`
}

type XMLConfig struct {
	Stanza XMLStanza `xml:"stanza"`
}

type XMLStanza struct {
	Name   string     `xml:"name,attr"`
	App    string     `xml:"app,attr"`
	Params []XMLParam `xml:"param"`
}

type XMLParam struct {
	Name  string `xml:"name,attr"`
	Value string `xml:",innerxml"`
}

func (x XMLInput) Extract() (string, int, int, string) {
	format := DefaultFormat
	grpcPort := DefaultGrpcPort
	httpPort := DefaultHTTPPort
	listeningAddress := DefaultListenAddress

	for _, p := range x.Configuration.Stanza.Params {
		switch p.Name {
		case "format":
			format = p.Value
		case "grpc_port":
			grpcPort, _ = strconv.Atoi(p.Value)
		case "http_port":
			httpPort, _ = strconv.Atoi(p.Value)
		case "listen_address":
			listeningAddress = p.Value
		}
	}

	return format, grpcPort, httpPort, listeningAddress
}

func ReadFromStdin() (XMLInput, error) {
	scanner := bufio.NewScanner(os.Stdin)
	text := ""
	for scanner.Scan() {
		text += scanner.Text()
	}

	var config XMLInput
	err := xml.Unmarshal([]byte(text), &config)
	return config, err
}
