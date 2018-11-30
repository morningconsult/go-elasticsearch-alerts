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
	"github.com/robfig/cron"
)

const (
	templateVersion string = "0.0.1"
	defaultStateIndexAlias string = "go-es-alerts"
	defaultTimestampFormat string = time.RFC3339
)

type QueryHandlerConfig struct {
	Name         string
	Logger       hclog.Logger
	Distributed  bool
	AlertMethods []alert.AlertMethod
	Client       *http.Client
	ESUrl        string
	QueryData    map[string]interface{}
	QueryIndex   string
	Schedule     string
	Filters      []string
}

type QueryHandler struct {
	HaveLockCh chan bool

	name         string
	distributed  bool
	hostname     string
	logger       hclog.Logger
	alertMethods []alert.AlertMethod
	client       *http.Client
	esURL        string
	queryIndex   string
	// queryURL     string
	queryData    map[string]interface{}
	// stateURL     string
	schedule     cron.Schedule
	filters      []string
}

func NewQueryHandler(config *QueryHandlerConfig) (*QueryHandler, error) {
	if config.Name == "" {
		return nil, errors.New("no rule name provided")
	}

	config.ESUrl = strings.TrimRight(config.ESUrl, "/")

	if config.ESUrl == "" {
		return nil, errors.New("no ElasticSearch URL provided")
	}

	if config.QueryIndex == "" {
		return nil, errors.New("no ElasticSearch index provided")
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

	return &QueryHandler{
		HaveLockCh:   make(chan bool, 1),
		name:         config.Name,
		distributed:  config.Distributed,
		hostname:     hostname,
		logger:       config.Logger,
		alertMethods: config.AlertMethods,
		client:       config.Client,
		esURL:        config.ESUrl,
		queryIndex:   config.QueryIndex,
		queryData:    config.QueryData,
		schedule:     schedule,
		filters:      config.Filters,
	}, nil
}

func (q *QueryHandler) Run(ctx context.Context, outputCh chan *alert.Alert, wg *sync.WaitGroup) {
	var (
		now           = time.Now()
		next          = now
		maintainState = true
		doneCh        = make(chan struct{})
		lockAcquired  = new(bool)
	)

	if q.distributed {
		go func(ctx context.Context) {
			for {
				select {
				case <-ctx.Done():
					close(doneCh)
					return
				case b := <-q.HaveLockCh:
					*lockAcquired = b
				}
			}
		}(ctx)
	} else {
		*lockAcquired = true
		close(doneCh)
	}

	defer func() {
		<-doneCh
		wg.Done()
	}()

	err := q.putTemplate(ctx)
	if err != nil {
		q.logger.Error(fmt.Sprintf("[Rule: %q] error creating template %q", q.name, q.templateName()), "error", err)
		select {
		case <-ctx.Done():
			return
		default:
			maintainState = false
		}
	}
	q.logger.Info(fmt.Sprintf("[Rule: %q] successfully created template %q", q.name, q.templateName()))

	if maintainState {
		t, err := q.getNextQuery(ctx)
		if err != nil {
			q.logger.Error(fmt.Sprintf("[Rule: %q] error looking up next scheduled query in ElasticSearch, running query now instead", q.name),
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
	} else {
		q.logger.Info(fmt.Sprintf("[Rule: %q] continuing without maintaining job state in ElasticSearch", q.name))
	}

	if *lockAcquired {
		q.logger.Info(fmt.Sprintf("[Rule: %q] scheduling query now (next execution at: %s)", q.name, next.Format(time.RFC822)))
	}

	for {
		hits := -1
		select {
		case <-ctx.Done():
			return
		case <-time.After(next.Sub(now)):
			if *lockAcquired {
				data, err := q.query(ctx)
				if err != nil {
					q.logger.Error(fmt.Sprintf("[Rule: %q] error querying ElasticSearch", q.name), "error", err)
					break
				}

				records, n, err := q.Transform(data)
				if err != nil {
					q.logger.Error(fmt.Sprintf("[Rule: %q] error processing response", q.name), "error", err)
					break
				}
				hits = n

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
				q.logger.Error(fmt.Sprintf("[Rule: %q] error creating next query document in ElasticSearch", q.name), "error", err)
			}
		}
	}
}

func (q *QueryHandler) putTemplate(ctx context.Context) error {
	payload := fmt.Sprintf(`{"index_patterns":["%s-status-%s-*"],"order":0,"aliases":{%q:{}},"settings":{"index":{"number_of_shards":3,"number_of_replicas":1,"auto_expand_replicas":"0-2","translog":{"flush_threshold_size":"752mb"},"sort":{"field":["next_query","rule_name","hostname"],"order":["desc","desc","desc"]}}},"mappings":{"_doc":{"dynamic_templates":[{"strings_as_keywords":{"match_mapping_type":"string","mapping":{"type":"keyword"}}}],"properties":{"@timestamp":{"type":"date"},"rule_name":{"type":"keyword"},"next_query":{"type":"date"},"hostname":{"type":"keyword"},"hits_count":{"type":"long","null_value":0},"hits":{"enabled":false}}}}}`, defaultStateIndexAlias, templateVersion, q.templateName())

	resp, err := q.makeRequest(ctx, "PUT", fmt.Sprintf("%s/_template/%s", q.esURL, q.templateName()), []byte(payload))
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
		return errors.New("ElasticSearch did not acknowledge creation of new template")
	}
	return nil
}

func (q *QueryHandler) getNextQuery(ctx context.Context) (*time.Time, error) {
	payload := fmt.Sprintf(`{"query":{"bool":{"should":[{"term":{"rule_name":{"value":%q}}},{"term":{"hostname":{"value":%q}}}]}},"sort":[{"next_query":{"order":"desc"}}],"size":1}`, q.cleanedName(), q.hostname)

	u, err := url.Parse(q.stateAliasURL() + "/_search")
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
		return nil, fmt.Errorf("received non-200 response status (status: %q). Response body:\n%s",
			resp.Status, q.readErrRespBody(resp))
	}

	var data = make(map[string]interface{})
	if err := jsonutil.DecodeJSONFromReader(resp.Body, &data); err != nil {
		return nil, fmt.Errorf("error JSON-decoding HTTP response: %v", err)
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

func (q *QueryHandler) setNextQuery(ctx context.Context, ts time.Time, hits int) error {
	payload := fmt.Sprintf(`{"@timestamp":%q,"rule_name":%q,"next_query":%q,"hostname":%q,"hits_count":%d}`, time.Now().Format(time.RFC3339), q.cleanedName(), ts.Format(defaultTimestampFormat), q.hostname, hits)

	resp, err := q.makeRequest(ctx, "POST", q.stateIndexURL()+"/_doc", []byte(payload))
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
		return nil, fmt.Errorf("error JSON-encoding ElasticSearch query body: %v", err)
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

func (q *QueryHandler) stateAliasURL() string {
	return fmt.Sprintf("%s/%s", q.esURL, q.templateName())
}

func (q *QueryHandler) stateIndexURL() string {
	return fmt.Sprintf("%s/%s", q.esURL, url.PathEscape(fmt.Sprintf("<%s-status-%s-{now/d}>", defaultStateIndexAlias, templateVersion)))
}

func (q *QueryHandler) templateName() string {
	return fmt.Sprintf("%s-%s", defaultStateIndexAlias, templateVersion)
}
