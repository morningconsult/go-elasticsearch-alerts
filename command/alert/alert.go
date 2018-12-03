// Copyright 2018 The Morning Consult, LLC or its affiliates. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License"). You may
// not use this file except in compliance with the License. A copy of the
// License is located at
//
//         https://www.apache.org/licenses/LICENSE-2.0
//
// or in the "license" file accompanying this file. This file is distributed
// on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either
// express or implied. See the License for the specific language governing
// permissions and limitations under the License.

package alert

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"github.com/hashicorp/go-hclog"
)

type Field struct {
	Key   string `json:"key" mapstructure:"key`
	Count int    `json:"doc_count" mapstructure:"doc_count"`
}

type Record struct {
	Title  string   `json:"filter,omitempty"`
	Text   string   `json:"text,omitempty"`
	Fields []*Field `json:"fields,omitempty"`
}

type Alert struct {
	ID       string
	RuleName string
	Methods  []AlertMethod
	Records  []*Record
}

type AlertMethod interface {
	Write(context.Context, string, []*Record) error
}

type AlertHandlerConfig struct {
	Logger hclog.Logger
}

type AlertHandler struct {
	logger hclog.Logger
	rand   *rand.Rand
	StopCh chan struct{}
	DoneCh chan struct{}
}

func NewAlertHandler(config *AlertHandlerConfig) *AlertHandler {
	return &AlertHandler{
		logger: config.Logger,
		rand:   rand.New(rand.NewSource(int64(time.Now().Nanosecond()))),
		StopCh: make(chan struct{}),
		DoneCh: make(chan struct{}),
	}
}

func (a *AlertHandler) Run(ctx context.Context, outputCh <-chan *Alert) {
	defer func() {
		close(a.DoneCh)
	}()

	a.logger.Info("starting alert handler")

	alertCh := make(chan func() (int, error), 8)
	active := newInventory()

	alertFunc := func(ctx context.Context, alertID, rule string, method AlertMethod, records []*Record) func() (int, error) {
		return func() (int, error) {
			if active.remaining(alertID) < 1 {
				active.deregister(alertID)
				return 0, nil
			}
			active.decrement(alertID)
			err := method.Write(ctx, rule, records)
			return active.remaining(alertID), err
		}
	}

	for {
		select {
		case <-ctx.Done():
			return
		case <-a.StopCh:
			return
		case alert := <-outputCh:
			a.logger.Info(fmt.Sprintf("[Alert Handler] new query results received from rule %q", alert.RuleName))
			for i, method := range alert.Methods {
				alertMethodID := fmt.Sprintf("%d|%s", i, alert.ID)
				active.register(alertMethodID)
				alertCh <- alertFunc(ctx, alertMethodID, alert.RuleName, method, alert.Records)
			}
		case writeAlert := <-alertCh:
			select {
			case <-ctx.Done():
				return
			default:
			}

			n, err := writeAlert()
			if err != nil {
				backoff := a.newBackoff()
				a.logger.Error("[Alert Handler] error returned by alert function", "error", err,
					"remaining_retries", n, "backoff", backoff.String())
				select {
				case <-ctx.Done():
					return
				case <-time.After(backoff):
					alertCh <- writeAlert
				}
			}
		}
	}
}

func (a *AlertHandler) newBackoff() time.Duration {
	return 2*time.Second + time.Duration(a.rand.Int63()%int64(time.Second*2)-int64(time.Second))
}
