// Copyright Splunk Inc. 2025
// SPDX-License-Identifier: Apache-2.0

package internal

// Scheme of the OTLP Input, presented to users of Splunk Platform in the Data Inputs view.
const Scheme = `
<scheme>
    <title>OTLP Input</title>
    <description>Receive data from OTLP</description>
    <streaming_mode>simple</streaming_mode>
    <use_single_instance>false</use_single_instance>
    <endpoint>
        <args>
            <arg name="grpc_port">
                <title>gRPC port</title>
                <description>Port on which the receiver will listen for gRPC OTLP traffic</description>
                <validation>is_avail_tcp_port('grpc_port')</validation>
                <required_on_create>false</required_on_create>
            </arg>

            <arg name="http_port">
                <title>HTTP Port</title>
                <description>Port on which the receiver will listen for HTTP OTLP traffic</description>
                <validation>is_avail_tcp_port('http_port')</validation>
                <required_on_create>false</required_on_create>
            </arg>

            <arg name="listen_address">
                <title>Listening address</title>
                <description>The listening address to bind the receiver to</description>
                <validation>
                  validate(!match("listen_address", "^\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3}$"), "Listening address is not valid")
                </validation>
                <required_on_create>false</required_on_create>
            </arg>

        </args>
    </endpoint>
</scheme>`
