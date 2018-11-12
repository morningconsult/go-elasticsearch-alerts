package query

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"path"
	// "strconv"
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
	defaultStateIndex      string = "elasticsearch-alerts-status"
	defaultTimestampFormat string = time.RFC3339
)

type QueryHandlerConfig struct {
	Name         string
	Logger       hclog.Logger
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
	name         string
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

	return &QueryHandler{
		name:         config.Name,
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
	var now  = time.Now()
	var next = now

	defer func() {
		wg.Done()
	}()

	t, err := q.getNextQuery(ctx)
	if err != nil {
		q.logger.Error("error looking up next scheduled query in ElasticSearch", err.Error(), "running query now instead")
	}
	if t != nil {
		next = *t
	}

	for {
		select {
		case <-ctx.Done():
			return
		case <-time.After(next.Sub(now)):
			data, err := q.query(ctx)
			if err != nil {
				q.logger.Error("error making HTTP request to ElasticSearch", err.Error())
				break
			}

			records, err := q.transform(data)
			if err != nil {
				q.logger.Error("error processing response", err.Error())
				break
			}

			id, err := uuid.GenerateUUID()
			if err != nil {
				q.logger.Error("error creating new UUID", err.Error())
				break
			}

			if records != nil && len(records) > 0 {
				a := &alert.Alert{
					ID:      id,
					Records: records,
					Methods: q.alertMethods,
				}
				outputCh <- a
			}
		}
		now = time.Now()
		next = q.schedule.Next(now)
		if err = q.setNextQuery(ctx, next); err != nil {
			q.logger.Error("error creating next query document in ElasticSearch", err.Error())
		}
	}
}

func (q *QueryHandler) query(ctx context.Context) (map[string]interface{}, error) {
	req, err := q.newRequest(ctx, "GET", path.Join(q.queryURL, "_search"), q.queryData)
	if err != nil {
		return nil, fmt.Errorf("error making HTTP request to ElasticSearch: %v", err)
	}

	resp, err := q.client.Do(req)
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
	payload := map[string]interface{}{
		"rule_name":  q.name,
		"next_query": ts.Format(defaultTimestampFormat),
	}

	req, err := q.newRequest(ctx, "POST", path.Join(q.stateURL, "_doc"), payload)
	if err != nil {
		return fmt.Errorf("error creating new request: %v", err)
	}

	resp, err := q.client.Do(req)
	if err != nil {
		return fmt.Errorf("error making HTTP request: %v", err)
	}
	resp.Body.Close()

	if resp.StatusCode != 201 {
		return fmt.Errorf("failed to create new document (received status: %q)", resp.Status)
	}
	return nil
}

func (q *QueryHandler) getNextQuery(ctx context.Context) (*time.Time, error) {
	payload := map[string]interface{}{
		"query": map[string]interface{}{
			"term": map[string]interface{}{
				"rule_name": q.name,
			},
		},
		"sort": []map[string]interface{}{
			map[string]interface{}{
				"next_query": map[string]interface{}{
					"order": "desc",
				},
			},
		},
		"size": 1,
	}

	req, err := q.newRequest(ctx, "GET", path.Join(q.stateURL, "_search"), payload)
	if err != nil {
		return nil, fmt.Errorf("error creating new request: %v", err)
	}
	u := req.URL.Query()

	// NOTE: If this URL query contains an asterisk or a comma or some other
	// character that would normally be URL-encoded, the following call to
	// q.Encode() will encode them
	u.Add("filter_path", "hits.hits._source.next_query")
	req.URL.RawQuery = u.Encode()

	resp, err := q.client.Do(req)
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

func (q *QueryHandler) newRequest(ctx context.Context, method, url string, payload map[string]interface{}) (*http.Request, error) {
	data, err := jsonutil.EncodeJSON(payload)
	if err != nil {
		return nil, fmt.Errorf("error JSON-encoding payload: %v", err)
	}

	req, err := http.NewRequest(method, url, bytes.NewBuffer(data))
	if err != nil {
		return nil, fmt.Errorf("error creating new HTTP request instance: %v", err)
	}
	req = req.WithContext(ctx)
	req.Header.Add("Content-Type", "application/json")
	return req, nil
}
