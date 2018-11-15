package query

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"io"
	"encoding/json"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/robfig/cron"
	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/go-uuid"
	"github.com/hashicorp/vault/helper/jsonutil"
	"gitlab.morningconsult.com/mci/go-elasticsearch-alerts/utils"
	"gitlab.morningconsult.com/mci/go-elasticsearch-alerts/command/alert"
)

const (
	defaultStateIndex      string = "go_elasticsearch_alerts_state"
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
	StateIndex   string
	Filters      []string
}

type QueryHandler struct {
	HaveLockCh   chan bool

	name         string
	distributed  bool
	hostname     string
	logger       hclog.Logger
	alertMethods []alert.AlertMethod
	client       *http.Client
	queryURL     string
	queryData    map[string]interface{}
	stateURL     string
	schedule     cron.Schedule
	filters      []string
}

func NewQueryHandler(config *QueryHandlerConfig) (*QueryHandler, error) {
	schedule, err := cron.Parse(config.Schedule)
	if err != nil {
		return nil, fmt.Errorf("error parsing cron schedule: %v", err)
	}

	if config.StateIndex == "" {
		config.StateIndex = defaultStateIndex
	}

	config.ESUrl = strings.TrimRight(config.ESUrl, "/")

	hostname, err := os.Hostname()
	if err != nil {
		return nil, fmt.Errorf("error getting hostname: %v", err)
	}

	return &QueryHandler{
		HaveLockCh:   make(chan bool, 1),
		name:         config.Name,
		distributed:  config.Distributed,
		hostname:     hostname,
		logger:       config.Logger,
		alertMethods: config.AlertMethods,
		client:       config.Client,
		queryURL:     fmt.Sprintf("%s/%s", config.ESUrl, config.QueryIndex),
		queryData:    config.QueryData,
		stateURL:     fmt.Sprintf("%s/%s", config.ESUrl, config.StateIndex),
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

	exists, err := q.stateIndexExists(ctx)
	if err != nil {
		q.logger.Error(fmt.Sprintf("[Rule: %q] error checking if index %q exists", q.name, q.stateURL), "error", err)
		select {
		case <-ctx.Done():
			return
		default:
		}
		q.logger.Info("continuing without maintaining job state in ElasticSearch")
		maintainState = false
	} else if !exists {
		q.logger.Info(fmt.Sprintf("[Rule: %q] ElasticSearch index %q does not exist. Attempting to create it.", q.name, q.stateURL))
		if err := q.createStateIndex(ctx); err != nil {
			q.logger.Error(fmt.Sprintf("[Rule: %q] error creating ElasticSearch state index %q", q.name, q.stateURL), "error", err)
			select {
			case <-ctx.Done():
				return
			default:
			}
			q.logger.Info("continuing without maintaining job state in ElasticSearch")
			maintainState = false
		} else {
			q.logger.Info(fmt.Sprintf("[Rule: %q] created new ElasticSearch index %q", q.name, q.stateURL))
		}
	}

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
	}

	if *lockAcquired {
		q.logger.Info(fmt.Sprintf("[Rule: %q] scheduling query now (next execution at: %s)", q.name, next.Format(time.RFC822)))
	}

	for {
		select {
		case <-ctx.Done():
			return
		case <-time.After(next.Sub(now)):
			if *lockAcquired {
				data, err := q.query(ctx)
				if err != nil {
					q.logger.Error(fmt.Sprintf("[Rule: %q] error making HTTP request to ElasticSearch", q.name), "error", err)
					break
				}

				records, err := q.transform(data)
				if err != nil {
					q.logger.Error(fmt.Sprintf("[Rule: %q] error processing response", q.name), "error", err)
					break
				}

				id, err := uuid.GenerateUUID()
				if err != nil {
					q.logger.Error(fmt.Sprintf("[Rule: %q] error creating new UUID", q.name), "error", err)
					break
				}

				if records != nil && len(records) > 0 {
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
			if err := q.setNextQuery(ctx, next); err != nil {
				q.logger.Error(fmt.Sprintf("[Rule: %q] error creating next query document in ElasticSearch", q.name), "error", err)
			}
		}
	}
}

func (q *QueryHandler) stateIndexExists(ctx context.Context) (bool, error) {
	resp, err := q.makeRequest(ctx, "GET", q.stateURL, nil)
	if err != nil {
		return false, fmt.Errorf("error making HTTP request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		if resp.StatusCode == 404 {
			return false, nil
		}
		return false, fmt.Errorf("error looking up ElasticSearch index %q (received status: %q). Response body:\n%s",
			q.stateURL, resp.Status, q.readErrRespBody(resp.Body))
	}
	return true, nil
}

func (q *QueryHandler) createStateIndex(ctx context.Context) error {
	payload := `{"settings":{"index":{"number_of_shards":3,"number_of_replicas":1,"auto_expand_replicas":"0-2","translog":{"flush_threshold_size":"752mb"},"sort":{"field":["next_query","rule_name","hostname"],"order":["desc","desc","desc"]}}},"mappings":{"_doc":{"dynamic_templates":[{"strings_as_keywords":{"match_mapping_type":"string","mapping":{"type":"keyword"}}}],"properties":{"@timestamp":{"type":"date"},"rule_name":{"type":"keyword"},"next_query":{"type":"date"},"hostname":{"type":"keyword"}}}}}`

	resp, err := q.makeRequest(ctx, "PUT", q.stateURL, []byte(payload))
	if err != nil {
		return fmt.Errorf("error making HTTP request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("error creating ElasticSearch index %q (received status: %q). Response body:\n%s",
			q.stateURL, resp.Status, q.readErrRespBody(resp.Body))
	}
	return nil
}

func (q *QueryHandler) query(ctx context.Context) (map[string]interface{}, error) {
	queryData, err := jsonutil.EncodeJSON(q.queryData)
	if err != nil {
		return nil, fmt.Errorf("error JSON-encoding ElasticSearch query body: %v", err)
	}

	resp, err := q.makeRequest(ctx, "GET", q.queryURL+"/_search", queryData)
	if err != nil {
		return nil, fmt.Errorf("error making HTTP request: %v", err)
	}
	defer resp.Body.Close()

	var data = make(map[string]interface{})
	if err := jsonutil.DecodeJSONFromReader(resp.Body, &data); err != nil {
		return nil, err
	}
	return data, nil
}

func (q *QueryHandler) setNextQuery(ctx context.Context, ts time.Time) error {
	payload := fmt.Sprintf(`{"rule_name":%q,"next_query":%q,"hostname":%q}`, q.name, ts.Format(defaultTimestampFormat), q.hostname)

	resp, err := q.makeRequest(ctx, "POST", q.stateURL+"/_doc", []byte(payload))
	if err != nil {
		return fmt.Errorf("error making HTTP request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 201 {
		return fmt.Errorf("failed to create new document (received status: %q). Response body:\n%s",
			resp.Status, q.readErrRespBody(resp.Body))
	}
	return nil
}

func (q *QueryHandler) getNextQuery(ctx context.Context) (*time.Time, error) {
	payload := fmt.Sprintf(`{"query":{"bool":{"should":[{"term":{"rule_name":{"value":%q}}},{"term":{"hostname":{"value":%q}}}]}},"sort":[{"next_query":{"order":"desc"}}],"size":1}`, q.name, q.hostname)

	u, err := url.Parse(q.stateURL+"/_search")
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

	var data = make(map[string]interface{})
	if err := jsonutil.DecodeJSONFromReader(resp.Body, &data); err != nil {
		return nil, fmt.Errorf("error JSON-decoding HTTP response: %v", err)
	}

	nextRaw := utils.Get(data, "hits.hits[0]._source.next_query")
	if nextRaw == nil {
		return nil, fmt.Errorf("no 'next_query' timestamp found")
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

func (q *QueryHandler) readErrRespBody(body io.Reader) string {
	var data map[string]interface{}
	if err := jsonutil.DecodeJSONFromReader(body, &data); err != nil {
		return ""
	}

	buf, err := json.MarshalIndent(data, "", "    ")
	if err != nil {
		return ""
	}
	return string(buf)
}