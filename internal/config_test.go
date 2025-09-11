// Copyright Splunk Inc. 2025
// SPDX-License-Identifier: Apache-2.0

package internal

import (
	"encoding/xml"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestParseInput(t *testing.T) {
	input := `
<?xml version="1.0" encoding="UTF-8"?>
<input>
  <server_host>773c28971b2a</server_host>
  <server_uri>https://127.0.0.1:8089</server_uri>
  <session_key>OwLHq7jpfgz0WLe5t8KwZuxT4QZRggryMB2io6Phimb2zi5ErifFvx0Eu8WTmfviO^KUKEA8CsGbVltVlCDlYOBM0RE8QoOjOHZhKnHsphk20XoqaK1KXTZj1N</session_key>
  <checkpoint_dir>/opt/splunk/var/lib/splunk/modinputs/otlpinput</checkpoint_dir>
  <configuration>
    <stanza name="otlpinput://specialmind" app="search">
      <param name="grpc_port">4317</param>
      <param name="host">$decideOnStartup</param>
      <param name="http_port">4318</param>
      <param name="index">main</param>
      <param name="listen_address">0.0.0.0</param>
      <param name="sourcetype">_otlpinput</param>
      <param name="start_by_shell">false</param>
    </stanza>
  </configuration>
</input>`

	var config XMLInput
	err := xml.Unmarshal([]byte(input), &config)
	require.NoError(t, err)

	require.Equal(t, "otlpinput://specialmind", config.Configuration.Stanza.Name)
	require.Equal(t, "main", config.Configuration.Stanza.Params[3].Value)

	grpcPort, httpPort, listeningAddress := config.Extract()

	require.Equal(t, 4317, grpcPort)
	require.Equal(t, "0.0.0.0", listeningAddress)
	require.Equal(t, 4318, httpPort)
}
