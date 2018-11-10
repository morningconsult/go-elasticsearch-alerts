package poll

import (
	"bytes"
	"context"
	"fmt"
	"encoding/json"
	"net/http"
	"path"
	// "strconv"
	"sync"
	"time"

	"github.com/robfig/cron"
	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/go-uuid"
	"github.com/mitchellh/mapstructure"
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

	t, err := q.nextQuery(ctx)
	if err != nil {
		q.logger.Error("error looking up next scheduled query in ElasticSearch", err.Error())
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
	}
}

func (q *QueryHandler) query(ctx context.Context) (map[string]interface{}, error) {
	req, err := q.newSearchRequest(ctx, q.queryURL, q.queryData)
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

func (q *QueryHandler) transform(respData map[string]interface{}) ([]*alert.Record, error) {
	var records []*alert.Record

	for _, filter := range q.filters {
		elems := utils.GetAll(respData, filter)
		if elems == nil || len(elems) < 1 {
			continue
		}

		record := &alert.Record{
			Title: filter,
		}

		var fields []*alert.Field
		for _, elem := range elems {
			obj, ok := elem.(map[string]interface{})
			if !ok {
				continue
			}

			field := new(alert.Field)
			if err := mapstructure.Decode(obj, field); err != nil {
				return nil, err
			}

			if field.Key == "" || field.Count < 1 {
				continue
			}

			fields = append(fields, field)
		}
		record.Fields = fields
		records = append(records, record)
	}

	// Make one record per hits.hits
	hitsRaw := utils.Get(respData, "hits.hits")
	if hitsRaw == nil {
		return records, nil
	}

	hits, ok := hitsRaw.([]interface{})
	if !ok {
		return records, nil
	}

	for _, hit := range hits {
		obj, ok := hit.(map[string]interface{})
		if !ok {
			continue
		}

		source, ok := obj["_source"].(map[string]interface{})
		if !ok {
			continue
		}

		data, err := json.MarshalIndent(source, "", "    ")
		if err != nil {
			return nil, err
		}
		
		record := &alert.Record{
			Title: "hits.hits._source",
			Text:  string(data),
		}
		records = append(records, record)
	}
	return records, nil
}

func (q *QueryHandler) nextQuery(ctx context.Context) (*time.Time, error) {
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

	req, err := q.newSearchRequest(ctx, q.stateURL, payload)
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
		return nil, err
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

func (q *QueryHandler) newSearchRequest(ctx context.Context, url string, payload map[string]interface{}) (*http.Request, error) {
	data, err := jsonutil.EncodeJSON(payload)
	if err != nil {
		return nil, fmt.Errorf("error JSON-encoding payload: %v", err)
	}

	req, err := http.NewRequest("GET", path.Join(url, "_search"), bytes.NewBuffer(data))
	if err != nil {
		return nil, fmt.Errorf("error creating new HTTP request instance: %v", err)
	}
	req = req.WithContext(ctx)
	req.Header.Add("Content-Type", "application/json")
	return req, nil
}
