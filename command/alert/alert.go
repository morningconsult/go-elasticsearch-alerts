// Copyright 2019 The Morning Consult, LLC or its affiliates. All Rights Reserved.
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

	hclog "github.com/hashicorp/go-hclog"
)

// Field represents a summary of the query results that
// match one of the filters specified in the 'filters'
// field of a rule configuration file.
type Field struct {
	// Key is a concatenation of any 'key' fields match a filter.
	// See github.com/morningconsult/go-elasticsearch-alerts/utils.GetAll()
	// for more information on how this key is created
	Key string `json:"key" mapstructure:"key"`

	// Count is the number of fields which match a filter
	Count int `json:"doc_count" mapstructure:"doc_count"`
}

// Record is used to send the results of an Elasticsearch query
// to the *alert.AlertHandler.
type Record struct {
	// Filter is the filter (either one of the elements of
	// the 'filter' array field or the 'body_field' field
	// of a rule configuration file) on which the results
	// of an Elasticsearch query is grouped
	Filter string `json:"filter,omitempty"`

	// Text is any text to be included with this record.
	// This will generally only be non-empty if the Filter
	// is the body field (e.g. 'hits.hits._source'). It is
	// generally just the JSON objects stringified and
	// concatenated
	Text string `json:"text,omitempty"`

	// BodyField is whether this record used the 'body_field'
	// index (per the rule configuration file) to group the
	// Elasticsearch response JSON
	BodyField bool `json:"-"`

	// Fields is the collection of elements of the
	// Elasticsearch response JSON that match the filter.
	// This will be non-empty only when the Filter is not
	// the body field
	Fields []*Field `json:"fields,omitempty"`
}

// Alert represents a unique set of results from an
// Elasticsearch query that the AlertHandler sends
// to the specified outputs.
type Alert struct {
	// ID is a unique UUID string identifying this alert
	ID string

	// RuleName is the name of the rule that generated
	// this alert
	RuleName string

	// Method is a set of alert.AlertMethod instances
	// which that the AlertHAndler will use to send
	// alerts
	Methods []Method

	// Records are the processed response data from an
	// Elasticsearch query
	Records []*Record
}

// Method is used to send alerts to some output.
type Method interface {
	Write(context.Context, string, []*Record) error
}

// HandlerConfig is used to provide the logger
// with which the alert handlers will log messages.
type HandlerConfig struct {
	Logger hclog.Logger
}

// Handler is used to send alerts to various outputs.
type Handler struct {
	logger hclog.Logger
	rand   *rand.Rand

	// StopCh is used to terminate the Run() loop
	StopCh chan struct{}

	// DoneCh is closed when Run() returns. Once closed,
	// Run() should not be called again
	DoneCh chan struct{}
}

// NewHandler creates a new *Handler instance.
func NewHandler(config *HandlerConfig) *Handler {
	return &Handler{
		logger: config.Logger,
		rand:   rand.New(rand.NewSource(int64(time.Now().Nanosecond()))), // nolint: gosec
		StopCh: make(chan struct{}),
		DoneCh: make(chan struct{}),
	}
}

// Run starts the *AlertHandler running. Once started, it
// waits to receive a new *Alert from outputCh. When it
// receives the alert, it will attempt to send the alert
// with the AlertMethods included in the alert. If it fails,
// it will backoff for a few seconds before trying to send
// the alert twice more. If it fails all three attempts, it
// will quit trying to send the alert. Run will return if
// ctx.Done() or StopCh becomes unblocked. Before returning,
// it will close the DoneCh. Once DoneCh is closed, Run
// should not be called again.
func (a *Handler) Run(ctx context.Context, outputCh <-chan *Alert) { // nolint: gocyclo
	defer func() {
		close(a.DoneCh)
	}()

	a.logger.Info("Starting alert handler")

	alertCh := make(chan func() (int, error), 8)
	active := newInventory()

	alertFunc := func(ctx context.Context, alertID, rule string, method Method, records []*Record) func() (int, error) {
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
			a.logger.Info(fmt.Sprintf("new query results received from rule %q", alert.RuleName))
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
				a.logger.Error("error returned by alert function", "error", err,
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

func (a *Handler) newBackoff() time.Duration {
	return 2*time.Second + time.Duration(a.rand.Int63()%int64(time.Second*2)-int64(time.Second))
}
