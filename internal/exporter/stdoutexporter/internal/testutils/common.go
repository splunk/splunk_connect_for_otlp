// Copyright Splunk Inc. 2025
// SPDX-License-Identifier: Apache-2.0

package testutils

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// Constants for Splunk components.
const (
	DefaultSourceTypeLabel = "com.splunk.sourcetype"
	DefaultSourceLabel     = "com.splunk.source"
	DefaultIndexLabel      = "com.splunk.index"
	DefaultNameLabel       = "otel.log.name"
)

var configFilePath = "./testdata/test_config.yaml"

type IntegrationTestsConfig struct {
	Host           string `yaml:"HOST"`
	User           string `yaml:"USER"`
	Password       string `yaml:"PASSWORD"`
	UIPort         string `yaml:"UI_PORT"`
	HecPort        string `yaml:"HEC_PORT"`
	ManagementPort string `yaml:"MANAGEMENT_PORT"`
	EventIndex     string `yaml:"EVENT_INDEX"`
	MetricIndex    string `yaml:"METRIC_INDEX"`
	TraceIndex     string `yaml:"TRACE_INDEX"`
	HecToken       string `yaml:"HEC_TOKEN"`
	SplunkImage    string `yaml:"SPLUNK_IMAGE"`
}

func GetConfigVariable(key string) (string, error) {
	// Read YAML file
	fileData, err := os.ReadFile(configFilePath)
	if err != nil {
		return "", fmt.Errorf("Error reading file: %v", err)
	}

	var config IntegrationTestsConfig
	err = yaml.Unmarshal(fileData, &config)
	if err != nil {
		return "", fmt.Errorf("Error decoding YAML: %v", err)
	}

	switch key {
	case "HOST":
		return config.Host, nil
	case "USER":
		return config.User, nil
	case "PASSWORD":
		return config.Password, nil
	case "UI_PORT":
		return config.UIPort, nil
	case "HEC_PORT":
		return config.HecPort, nil
	case "MANAGEMENT_PORT":
		return config.ManagementPort, nil
	case "EVENT_INDEX":
		return config.EventIndex, nil
	case "METRIC_INDEX":
		return config.MetricIndex, nil
	case "TRACE_INDEX":
		return config.TraceIndex, nil
	case "HEC_TOKEN":
		return config.HecToken, nil
	case "SPLUNK_IMAGE":
		return config.SplunkImage, nil
	default:
		fmt.Println("Invalid field")
		return "None", nil
	}
}
