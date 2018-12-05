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

package query

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/go-uuid"
	"github.com/hashicorp/go-cleanhttp"
	"github.com/hashicorp/vault/helper/jsonutil"
	"github.com/morningconsult/go-elasticsearch-alerts/command/alert"
	"github.com/morningconsult/go-elasticsearch-alerts/utils"
	"github.com/morningconsult/go-elasticsearch-alerts/utils/lock"
	"github.com/robfig/cron"
)

const (
	templateVersion string = "0.0.1"
	defaultStateIndexAlias string = "go-es-alerts"
	defaultTimestampFormat string = time.RFC3339
	defaultBodyField  string = "hits.hits._source"
)

// QueryHandlerConfig is passed as an argument to NewQueryHandler()
type QueryHandlerConfig struct {
	// Name is the name of the rule. This should come from
	// the 'name' field of the rule configuration file
	Name string

	// AlertMethods will be passed along with any results returned
	// by a query to the alert handler via the outputCh
	AlertMethods []alert.AlertMethod

	// Client is an *http.Client instance that will be used to
	// query Elasticsearch
	Client *http.Client

	// ESUrl is the URL of the Elasticsearch instance. This should
	// come from the 'elasticsearch.server.url' field of the main
	// configuration file
	ESUrl string

	// QueryData is the payload to be included in the query. This
	// should come from the 'body' field of the rule configuration
	// file
	QueryData map[string]interface{}

	// QueryIndex is the Elasticsearch index to be queried. This
	// should come from the 'index' field of the rule configuration
	// file
	QueryIndex string

	// Schedule is the interval at which the defined Elasticsearch
	// query should executed (in cron syntax)
	Schedule string

	// BodyField is the field of the JSON response returned by
	// Elasticsearch to be grouped on and subsequently sent to
	// the specified outputs. This should come from the 'body_field'
	// field of the rule configuration file
	BodyField string

	// Filters are the additional fields to be grouped on. These
	// should come from the 'filters' field of the rule configuration
	// file
	Filters []string

	Logger hclog.Logger
}

// QueryHandler performs the defined Elasticsearch query at the
// specified interval and sends results to the AlertHandler if
// there are any.
type QueryHandler struct {
	// StopCh terminates the Run() method when closed
	StopCh       chan struct{}

	name         string
	hostname     string
	logger       hclog.Logger
	alertMethods []alert.AlertMethod
	client       *http.Client
	esURL        string
	queryIndex   string
	queryData    map[string]interface{}
	schedule     cron.Schedule
	bodyField    string
	filters      []string
}

// NewQueryHandler creates a new *QueryHandler instance
func NewQueryHandler(config *QueryHandlerConfig) (*QueryHandler, error) {
	if config == nil {
		config = &QueryHandlerConfig{}
	}

	if config.Name == "" {
		return nil, errors.New("no rule name provided")
	}

	config.ESUrl = strings.TrimRight(config.ESUrl, "/")

	if config.ESUrl == "" {
		return nil, errors.New("no Elasticsearch URL provided")
	}

	if config.QueryIndex == "" {
		return nil, errors.New("no Elasticsearch index provided")
	}

	if len(config.AlertMethods) < 1 {
		return nil, errors.New("at least one alert method must be specified")
	}

	if config.QueryData == nil || len(config.QueryData) < 1 {
		return nil, errors.New("no query body provided")
	}

	hostname, err := os.Hostname()
	if err != nil {
		return nil, fmt.Errorf("error getting hostname: %v", err)
	}

	schedule, err := cron.Parse(config.Schedule)
	if err != nil {
		return nil, fmt.Errorf("error parsing cron schedule: %v", err)
	}

	if config.Logger == nil {
		config.Logger = hclog.Default()
	}

	if config.Client == nil {
		config.Client = cleanhttp.DefaultClient()
	}

	if config.BodyField == "" {
		config.BodyField = defaultBodyField
	}

	return &QueryHandler{
		StopCh:       make(chan struct{}),

		name:         config.Name,
		hostname:     hostname,
		logger:       config.Logger,
		alertMethods: config.AlertMethods,
		client:       config.Client,
		esURL:        config.ESUrl,
		queryIndex:   config.QueryIndex,
		queryData:    config.QueryData,
		schedule:     schedule,
		bodyField:    config.BodyField,
		filters:      config.Filters,
	}, nil
}

// Run starts the QueryHandler. It first attempts to get the "state"
// document for this rule from Elasticsearch in order to schedule 
// the next execution at the last scheduled time. If it does not find
// such a document, or if the next scheduled query is in the past, it
// will execute the query immediately. Afterwards, it will attempt to
// write a new state document to Elasticsearch in which the 'next_query'
// equals the next time the query shall be executed per the provided
// cron schedule. It will only execute the query if distLock.Acquired()
// is true. 
func (q *QueryHandler) Run(ctx context.Context, outputCh chan *alert.Alert, wg *sync.WaitGroup, distLock *lock.Lock) {
	var (
		now           = time.Now()
		next          = now
		maintainState = true
	)

	defer func() {
		wg.Done()
	}()

	t, err := q.getNextQuery(ctx)
	if err != nil {
		q.logger.Error(fmt.Sprintf("[Rule: %q] error looking up next scheduled query in Elasticsearch, running query now instead", q.name),
			"error", err)
		select {
		case <-ctx.Done():
			return
		default:
		}
	}
	if t != nil {
		next = *t
	}

	if distLock.Acquired() {
		q.logger.Info(fmt.Sprintf("[Rule: %q] scheduling query now (next execution at: %s)", q.name, next.Format(time.RFC822)))
	}

	for {
		hits := []map[string]interface{}{}
		select {
		case <-ctx.Done():
			return
		case <-q.StopCh:
			return
		case <-time.After(next.Sub(now)):
			if distLock.Acquired() {
				data, err := q.query(ctx)
				if err != nil {
					q.logger.Error(fmt.Sprintf("[Rule: %q] error querying Elasticsearch", q.name), "error", err)
					break
				}

				records, tmp, err := q.Transform(data)
				if err != nil {
					q.logger.Error(fmt.Sprintf("[Rule: %q] error processing response", q.name), "error", err)
					break
				}
				hits = tmp

				if records != nil && len(records) > 0 {
					id, err := uuid.GenerateUUID()
					if err != nil {
						q.logger.Error(fmt.Sprintf("[Rule: %q] error creating new random UUID", q.name), "error", err)
						break
					}

					a := &alert.Alert{
						ID:       id,
						RuleName: q.name,
						Records:  records,
						Methods:  q.alertMethods,
					}
					outputCh <- a
				}
			}
		}
		now = time.Now()
		next = q.schedule.Next(now)
		if maintainState {
			if err := q.setNextQuery(ctx, next, hits); err != nil {
				q.logger.Error(fmt.Sprintf("[Rule: %q] error creating next query document in Elasticsearch", q.name), "error", err)
				q.logger.Info(fmt.Sprintf("[Rule: %q] continuing without maintaining job state in Elasticsearch", q.name))
				maintainState = false
			}
		}
	}
}

// PutTemplate attempts to create a template in Elasticsearch which
// will serve as an alias for the state indices. The state indices
// will be named 'go-es-alerts-status-{date}'; therefore, this template
// enables searching all state indices via this alias
func (q *QueryHandler) PutTemplate(ctx context.Context) error {
	payload := fmt.Sprintf(`{"index_patterns":["%s-status-%s-*"],"order":0,"aliases":{%q:{}},"settings":{"index":{"number_of_shards":5,"number_of_replicas":1,"auto_expand_replicas":"0-2","translog":{"flush_threshold_size":"752mb"},"sort":{"field":["next_query","rule_name","hostname"],"order":["desc","desc","desc"]}}},"mappings":{"_doc":{"dynamic_templates":[{"strings_as_keywords":{"match_mapping_type":"string","mapping":{"type":"keyword"}}}],"properties":{"@timestamp":{"type":"date"},"rule_name":{"type":"keyword"},"next_query":{"type":"date"},"hostname":{"type":"keyword"},"hits_count":{"type":"long","null_value":0},"hits":{"enabled":false}}}}}`, defaultStateIndexAlias, templateVersion, q.TemplateName())

	resp, err := q.makeRequest(ctx, "PUT", fmt.Sprintf("%s/_template/%s", q.esURL, q.TemplateName()), []byte(payload))
	if err != nil {
		return fmt.Errorf("error making HTTP request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("received non-200 response status (status: %q). Response body:\n%s",
			resp.Status, q.readErrRespBody(resp))
	}

	var data map[string]interface{}
	if err = jsonutil.DecodeJSONFromReader(resp.Body, &data); err != nil {
		return fmt.Errorf("error JSON-decoding response body: %v", err)
	}

	ackRaw, ok := data["acknowledged"]
	if !ok {
		return errors.New("JSON response has no 'acknowledged' field")
	}
	ack, ok := ackRaw.(bool)
	if !ok {
		return errors.New("value of 'acknowledged' field of JSON response cannot be cast to boolean")
	}
	if !ack {
		return errors.New("Elasticsearch did not acknowledge creation of new template")
	}
	return nil
}

// getNextQuery queries the state indices for the most recently-
// created document belonging to this rule. It then attempts to
// parse the 'next_query' field in order to inform the Run() loop
// when to next execute the query.
func (q *QueryHandler) getNextQuery(ctx context.Context) (*time.Time, error) {
	payload := fmt.Sprintf(`{"query":{"bool":{"must":[{"term":{"rule_name":{"value":%q}}}]}},"sort":[{"next_query":{"order":"desc"}}],"size":1}`, q.cleanedName())

	u, err := url.Parse(q.StateAliasURL() + "/_search")
	if err != nil {
		return nil, fmt.Errorf("error parsing URL: %v", err)
	}
	query := u.Query()
	query.Add("filter_path", "hits.hits._source.next_query")
	u.RawQuery = query.Encode()

	resp, err := q.makeRequest(ctx, "GET", u.String(), []byte(payload))
	if err != nil {
		return nil, fmt.Errorf("error making HTTP request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("received non-200 response status (status: %q).", resp.Status)
	}

	var data = make(map[string]interface{})
	if err := jsonutil.DecodeJSONFromReader(resp.Body, &data); err != nil {
		return nil, fmt.Errorf("error JSON-decoding HTTP response: %v", err)
	}

	if len(data) < 1 {
		return nil, errors.New("no records found for this rule")
	}

	nextRaw := utils.Get(data, "hits.hits[0]._source.next_query")
	if nextRaw == nil {
		return nil, fmt.Errorf("field 'next_query' not found")
	}

	nextString, ok := nextRaw.(string)
	if !ok {
		return nil, fmt.Errorf("'next_query' value could not be cast to string")
	}

	t, err := time.Parse(defaultTimestampFormat, nextString)
	if err != nil {
		return nil, fmt.Errorf("error parsing time: %v", err)
	}
	return &t, nil
}

// setNextQuery creates a new document in a state index to
// inform the Run() loop when to next execute the query if
// the process gets restarted
func (q *QueryHandler) setNextQuery(ctx context.Context, ts time.Time, hits []map[string]interface{}) error {
	status := struct {
		Time  string `json:"@timestamp"`
		Name  string `json:"rule_name"`
		Next  string `json:"next_query"`
		Host  string `json:"hostname"`
		NHits int    `json:"hits_count"`
		Hits  []map[string]interface{} `json:"hits,omitempty"`
	}{
		Time:  time.Now().Format(defaultTimestampFormat),
		Name:  q.cleanedName(),
		Next:  ts.Format(defaultTimestampFormat),
		Host:  q.hostname,
		NHits: len(hits),
		Hits:  hits,
	}

	payload, err := jsonutil.EncodeJSON(status)
	if err != nil {
		return fmt.Errorf("error JSON-encoding data: %v", err)
	}

	resp, err := q.makeRequest(ctx, "POST", q.StateIndexURL()+"/_doc", payload)
	if err != nil {
		return fmt.Errorf("error making HTTP request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 201 {
		return fmt.Errorf("failed to create new document (received status: %q). Response body:\n%s",
			resp.Status, q.readErrRespBody(resp))
	}
	return nil
}

func (q *QueryHandler) query(ctx context.Context) (map[string]interface{}, error) {
	queryData, err := jsonutil.EncodeJSON(q.queryData)
	if err != nil {
		return nil, fmt.Errorf("error JSON-encoding Elasticsearch query body: %v", err)
	}

	resp, err := q.makeRequest(ctx, "GET", fmt.Sprintf("%s/%s/_search", q.esURL, q.queryIndex), queryData)
	if err != nil {
		return nil, fmt.Errorf("error making HTTP request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("received non-200 response status (status: %q). Response body:\n%s",
			resp.Status, q.readErrRespBody(resp))
	}

	var data = make(map[string]interface{})
	if err := jsonutil.DecodeJSONFromReader(resp.Body, &data); err != nil {
		return nil, err
	}
	return data, nil
}

func (q *QueryHandler) cleanedName() string {
	return strings.Replace(strings.ToLower(q.name), " ", "-", -1)
}

func (q *QueryHandler) makeRequest(ctx context.Context, method, url string, payload []byte) (*http.Response, error) {
	req, err := q.newRequest(ctx, method, url, payload)
	if err != nil {
		return nil, fmt.Errorf("error creating new request: %v", err)
	}
	return q.client.Do(req)
}

func (q *QueryHandler) newRequest(ctx context.Context, method, url string, payload []byte) (*http.Request, error) {
	var req *http.Request
	var err error
	if payload != nil {
		req, err = http.NewRequest(method, url, bytes.NewBuffer(payload))
		if err != nil {
			return nil, fmt.Errorf("error creating new HTTP request instance: %v", err)
		}
		req.Header.Add("Content-Type", "application/json")
	} else {
		req, err = http.NewRequest(method, url, nil)
		if err != nil {
			return nil, fmt.Errorf("error creating new HTTP request instance: %v", err)
		}
	}

	req = req.WithContext(ctx)
	return req, nil
}

func (q *QueryHandler) readErrRespBody(resp *http.Response) string {
	switch resp.Header.Get("Content-Type") {
	case "application/json":
		var data map[string]interface{}
		if err := jsonutil.DecodeJSONFromReader(resp.Body, &data); err != nil {
			return ""
		}

		buf, err := json.MarshalIndent(data, "", "    ")
		if err != nil {
			return ""
		}
		return string(buf)
	default:
		data, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return ""
		}
		return string(data)
	}
	return ""
}

func (q *QueryHandler) StateAliasURL() string {
	return fmt.Sprintf("%s/%s", q.esURL, q.TemplateName())
}

func (q *QueryHandler) StateIndexURL() string {
	return fmt.Sprintf("%s/%s", q.esURL, url.PathEscape(fmt.Sprintf("<%s-status-%s-{now/d}>", defaultStateIndexAlias, templateVersion)))
}

func (q *QueryHandler) TemplateName() string {
	return fmt.Sprintf("%s-%s", defaultStateIndexAlias, templateVersion)
}
