// Copyright Splunk Inc. 2025
// SPDX-License-Identifier: Apache-2.0

package internal

import (
	"os"
	"os/signal"
	"sync"
	"syscall"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/component/componentstatus"
)

var (
	_ component.Host           = &TTYHost{}
	_ componentstatus.Reporter = &TTYHost{}
)

type TTYHost struct {
	ErrStatus    chan error
	Extensions   map[component.ID]component.Component
	shutdownOnce sync.Once
}

func (t *TTYHost) Start() {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigs
		t.Report(componentstatus.NewEvent(componentstatus.StatusStopping))
	}()
}

func (t *TTYHost) Wait() error {
	return <-t.ErrStatus
}

func (t *TTYHost) Report(event *componentstatus.Event) {
	if event.Status() == componentstatus.StatusStopping {
		t.shutdownOnce.Do(func() {
			close(t.ErrStatus)
		})
	}
	if event.Err() != nil {
		t.ErrStatus <- event.Err()
	}
}

func (t *TTYHost) GetExtensions() map[component.ID]component.Component {
	return t.Extensions
}
