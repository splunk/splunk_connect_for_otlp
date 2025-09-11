// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package internal

import (
	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/ptrace"
	conventions "go.opentelemetry.io/otel/semconv/v1.27.0"
)

// hecEvent is a data structure holding a span event to export explicitly to Splunk HEC.
type hecEvent struct {
	Attributes map[string]any    `json:"attributes,omitempty"`
	Name       string            `json:"name"`
	Timestamp  pcommon.Timestamp `json:"timestamp"`
}

// hecLink is a data structure holding a span link to export explicitly to Splunk HEC.
type hecLink struct {
	Attributes map[string]any `json:"attributes,omitempty"`
	TraceID    string         `json:"trace_id"`
	SpanID     string         `json:"span_id"`
	TraceState string         `json:"trace_state"`
}

// hecSpanStatus is a data structure holding the status of a span to export explicitly to Splunk HEC.
type hecSpanStatus struct {
	Message string `json:"message"`
	Code    string `json:"code"`
}

// hecSpan is a data structure used to export explicitly a span to Splunk HEC.
type hecSpan struct {
	TraceID    string            `json:"trace_id"`
	SpanID     string            `json:"span_id"`
	ParentSpan string            `json:"parent_span_id"`
	Name       string            `json:"name"`
	Attributes map[string]any    `json:"attributes,omitempty"`
	EndTime    pcommon.Timestamp `json:"end_time"`
	Kind       string            `json:"kind"`
	Status     hecSpanStatus     `json:"status,omitempty"`
	StartTime  pcommon.Timestamp `json:"start_time"`
	Events     []hecEvent        `json:"events,omitempty"`
	Links      []hecLink         `json:"links,omitempty"`
}

func MapSpanToSplunkEvent(resource pcommon.Resource, span ptrace.Span) *Event {
	sourceKey := DefaultSourceLabel
	sourceTypeKey := DefaultSourceTypeLabel
	indexKey := DefaultIndexLabel
	hostKey := string(conventions.HostNameKey)

	host := unknownHostName
	source := ""
	sourceType := ""
	index := ""
	commonFields := map[string]any{}
	for k, v := range resource.Attributes().All() {
		switch k {
		case hostKey:
			host = v.Str()
		case sourceKey:
			source = v.Str()
		case sourceTypeKey:
			sourceType = v.Str()
		case indexKey:
			index = v.Str()
		case HecTokenLabel:
			// ignore
		default:
			commonFields[k] = v.AsString()
		}
	}

	se := &Event{
		Time:       timestampToSecondsWithMillisecondPrecision(span.StartTimestamp()),
		Host:       host,
		Source:     source,
		SourceType: sourceType,
		Index:      index,
		Event:      toHecSpan(span),
		Fields:     commonFields,
	}

	return se
}

func toHecSpan(span ptrace.Span) hecSpan {
	attributes := span.Attributes().AsRaw()

	links := make([]hecLink, span.Links().Len())
	for i := 0; i < span.Links().Len(); i++ {
		link := span.Links().At(i)
		linkAttributes := link.Attributes().AsRaw()
		links[i] = hecLink{
			Attributes: linkAttributes,
			TraceID:    TraceIDToHexOrEmptyString(link.TraceID()),
			SpanID:     SpanIDToHexOrEmptyString(link.SpanID()),
			TraceState: link.TraceState().AsRaw(),
		}
	}
	events := make([]hecEvent, span.Events().Len())
	for i := 0; i < span.Events().Len(); i++ {
		event := span.Events().At(i)
		eventAttributes := event.Attributes().AsRaw()
		events[i] = hecEvent{
			Attributes: eventAttributes,
			Name:       event.Name(),
			Timestamp:  event.Timestamp(),
		}
	}
	status := hecSpanStatus{
		Message: span.Status().Message(),
		Code:    StatusCodeStr(span.Status().Code()),
	}
	return hecSpan{
		TraceID:    TraceIDToHexOrEmptyString(span.TraceID()),
		SpanID:     SpanIDToHexOrEmptyString(span.SpanID()),
		ParentSpan: SpanIDToHexOrEmptyString(span.ParentSpanID()),
		Name:       span.Name(),
		Attributes: attributes,
		StartTime:  span.StartTimestamp(),
		EndTime:    span.EndTimestamp(),
		Kind:       SpanKindStr(span.Kind()),
		Status:     status,
		Links:      links,
		Events:     events,
	}
}
