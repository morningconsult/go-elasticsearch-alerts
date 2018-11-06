package poll

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"path"
	"sync"
	"time"

	"github.com/robfig/cron"
	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/vault/helper/jsonutil"
	"gitlab.morningconsult.com/mci/go-elasticsearch-alerts/utils"
)

const (
	defaultStateIndex      string = "elasticsearch-alerts-status"
	defaultTimestampFormat string = time.RFC3339
)

type QueryHandlerConfig struct {
	Name       string
	Logger     hclog.Logger
	Client     *http.Client
	ESUrl      string
	Query      map[string]interface{}
	QueryIndex string
	Schedule   string
	StateIndex string
	Filters    []string
}

type QueryHandler struct {
	name      string
	logger    hclog.Logger
	client    *http.Client
	queryURL  string
	queryData map[string]interface{}
	stateURL  string
	schedule  cron.Schedule
	filters   []string
}

// type Record struct {
// 	NextQuery  time.Time
// 	Executed   bool
// 	ExecutedAt time.Time
// }

func NewQueryHandler(config *QueryHandlerConfig) (*QueryHandler, error) {
	schedule, err := cron.Parse(config.Schedule)
	if err != nil {
		return nil, fmt.Errorf("error parsing cron schedule: %v", err)
	}

	if config.StateIndex == "" {
		config.StateIndex = defaultStateIndex
	}

	return &QueryHandler{
		name:      config.Name,
		logger:    config.Logger,
		client:    config.Client,
		queryURL:  fmt.Sprintf("%s/%s", config.ESUrl, config.QueryIndex),
		queryData: config.QueryData,
		stateURL:  fmt.Sprintf("%s/%s", config.ESUrl, config.StateIndex),
		schedule:  schedule,
	}, nil
}

func (q *QueryHandler) Run(ctx context.Context, outputCh chan interface{}, wg *sync.WaitGroup) {
	var now = time.Now()

	defer func() {
		wg.Done()
	}()

	next, err := q.nextQuery(ctx)
	if err != nil {
		logger.Error("error looking up next scheduled query in ElasticSearch", err.Error())
		next = now
	}

	for {
		select {
		case <-ctx.Done():
			return
		case <-time.After(next.Sub(now)):
			records, err := q.query(ctx)
			if err != nil {
				logger.Error("error making HTTP request to ElasticSearch", err.Error())
			}
			if records != nil {
				outputCh <- records
			}
			now = time.Now()
			next = q.schedule.Next(now)
			// q.scheduleNext(next)
		}
	}
}

func (q *QueryHandler) query(ctx context.Context) (map[string]interface{}, error) {
	req, err := q.newSearchRequest(ctx, q.queryURL, q.queryData)
	if err != nil {
		return nil, fmt.Errorf("error making HTTP request to ElasticSearch: %v", err)
	}

	resp, err := h.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error making HTTP request: %v", err)
	}
	defer resp.Body.Close()

	var data = make(map[string]interface{})
	if err := jsonutil.DecodeJSONFromReader(resp.Body, &data); err != nil {
		return nil, err
	}

	// process the data
	attachments := q.transform(ctx, data)
}

func (q *QueryHandler) nextQuery(ctx context.Context) (time.Time, error) {
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
	q := req.URL.Query()

	// NOTE: If this URL query contains an asterisk or a comma or some other
	// character that would normally be URL-encoded, the following call to
	// q.Encode() will encode them
	q.Add("filter_path", "hits.hits._source.next_query")
	req.URL.RawQuery = q.Encode()

	resp, err := h.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error making HTTP request: %v", err)
	}
	defer resp.Body.Close()

	var data = make(map[string]interface{})
	if err := jsonutil.DecodeJSONFromReader(resp.Body, &data); err != nil {
		return nil, err
	}

	nextRaw := utils.Get("hits.hits[0]._source.next_query", data)
	if nextRaw == nil {
		return nil, fmt.Errorf("no 'next_query' timestamp found")
	}

	nextString, ok := ti.(string)
	if !ok {
		return nil, fmt.Errorf("'next_query' value could not be cast to string")
	}

	return time.Parse(defaultTimestampFormat, nextString)
}

func (q *QueryHandler) newSearchRequest(ctx context.Context, url string, payload map[string]interface{}) (*http.Request, error) {
	data, err := jsonutil.EncodeJSON(payload)
	if err != nil {
		return nil, fmt.Errorf("error JSON-encoding payload: %v", err)
	}

	req, err := http.NewRequest("GET", path.Join(url, "_search")), bytes.NewBuffer(data))
	if err != nil {
		return nil, fmt.Errorf("error creating new HTTP request instance: %v", err)
	}
	req = req.WithContext(ctx)
	req.Header.Add("Content-Type", "application/json")
	return req, nil
}

func (q *QueryHandler) transform(ctx context.Context, resp map[string]interface{}) []*Attachments
