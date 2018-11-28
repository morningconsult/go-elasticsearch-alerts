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
	"fmt"
	"io"
	"net/http"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/hashicorp/go-cleanhttp"
	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/go-uuid"
	"github.com/morningconsult/go-elasticsearch-alerts/command/alert"
	"github.com/morningconsult/go-elasticsearch-alerts/command/alert/file"
)

var SkipElasticSearchTests bool = false

const (
	ElasticSearchURL string = "http://127.0.0.1:9200"
	ConsulURL        string = "http://127.0.0.1:8500"
)

func TestNewQueryHandler_DefaultStateIndex(t *testing.T) {
	qh, err := NewQueryHandler(&QueryHandlerConfig{
		ESUrl:    ElasticSearchURL,
		Schedule: "* * * * * *",
	})
	if err != nil {
		t.Fatal(err)
	}
	expected := fmt.Sprintf("%s/%s", ElasticSearchURL, defaultStateIndex)
	if qh.stateURL != expected {
		t.Errorf("QueryHandler has wrong stateURL value (got %q, expected %q)", qh.stateURL, expected)
	}
}

func TestNewQueryHandler_CronParseError(t *testing.T) {
	_, err := NewQueryHandler(&QueryHandlerConfig{
		Schedule: "i am not a valid cron",
	})
	if err == nil {
		t.Error("expected an error but didn't receive one")
	}
}

func TestStateIndexExists(t *testing.T) {
	if SkipElasticSearchTests {
		t.Skipf("unable to connect to ElasticSearch at %s. Skipping test.", ElasticSearchURL)
	}

	id, err := uuid.GenerateUUID()
	if err != nil {
		t.Fatal(err)
	}

	stateIndex := ElasticSearchURL + "/" + id
	client := cleanhttp.DefaultClient()

	delIndexFunc := func() {
		req, err := http.NewRequest("DELETE", stateIndex, nil)
		if err != nil {
			t.Fatal(err)
		}
		_, err = client.Do(req)
		if err != nil {
			t.Fatal(err)
		}
	}
	createIndexFunc := func() {
		req, err := http.NewRequest("PUT", stateIndex, nil)
		if err != nil {
			t.Fatal(err)
		}
		_, err = client.Do(req)
		if err != nil {
			t.Fatal(err)
		}
	}

	qh := &QueryHandler{
		stateURL: stateIndex,
		client:   client,
		logger:   hclog.NewNullLogger(),
	}

	cases := []struct {
		name   string
		exists bool
		err    bool
	}{
		{
			"does-not-exist",
			false,
			false,
		},
		{
			"bad-url",
			false,
			true,
		},
		{
			"exists",
			true,
			false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			delIndexFunc()
			if tc.name == "exists" {
				createIndexFunc()
				defer delIndexFunc()
			}
			if tc.name == "bad-url" {
				id, err := uuid.GenerateUUID()
				if err != nil {
					t.Fatal(err)
				}
				currentURL := qh.stateURL
				qh.stateURL = "http://" + id + ".com"
				defer func() {
					qh.stateURL = currentURL
				}()
			}

			ctx := context.Background()
			exists, err := qh.stateIndexExists(ctx)
			if tc.err {
				if err == nil {
					t.Fatal("expected an error but didn't receive one")
				}
				return
			}

			if err != nil {
				t.Fatal(err)
			}

			if exists != tc.exists {
				t.Fatalf("returned unexpected value (got %t, expected %t)", exists, tc.exists)
			}
		})
	}
}

func TestCreateStateIndex(t *testing.T) {
	if SkipElasticSearchTests {
		t.Skipf("unable to connect to ElasticSearch at %q. Skipping test.", ElasticSearchURL)
	}

	id, err := uuid.GenerateUUID()
	if err != nil {
		t.Fatal(err)
	}

	stateIndex := ElasticSearchURL + "/" + id
	client := cleanhttp.DefaultClient()

	delIndexFunc := func() {
		req, err := http.NewRequest("DELETE", stateIndex, nil)
		if err != nil {
			t.Fatal(err)
		}
		_, err = client.Do(req)
		if err != nil {
			t.Fatal(err)
		}
	}
	createIndexFunc := func() {
		req, err := http.NewRequest("PUT", stateIndex, nil)
		if err != nil {
			t.Fatal(err)
		}
		_, err = client.Do(req)
		if err != nil {
			t.Fatal(err)
		}
	}

	qh := &QueryHandler{
		stateURL: stateIndex,
		client:   client,
		logger:   hclog.NewNullLogger(),
	}

	cases := []struct {
		name   string
		exists bool
		err    bool
	}{
		{
			"creates",
			false,
			false,
		},
		{
			"bad-url",
			false,
			true,
		},
		{
			"already-exists",
			true,
			true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			delIndexFunc()
			if tc.name == "already-exists" {
				createIndexFunc()
				defer delIndexFunc()
			}
			if tc.name == "bad-url" {
				id, err := uuid.GenerateUUID()
				if err != nil {
					t.Fatal(err)
				}
				currentURL := qh.stateURL
				qh.stateURL = "http://" + id + ".com"
				defer func() {
					qh.stateURL = currentURL
				}()
			}

			ctx := context.Background()
			err = qh.createStateIndex(ctx)
			if tc.err {
				if err == nil {
					t.Fatal("expected an error but didn't receive one")
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}

			resp, err := client.Get(qh.stateURL)
			if err != nil {
				t.Fatal(err)
			}
			if resp.StatusCode != 200 {
				t.Fatalf("ElasticSearch index %q should have been created, but wasn't", qh.stateURL)
			}
		})
	}
}

func TestGetNextQuery(t *testing.T) {
	if SkipElasticSearchTests {
		t.Skipf("unable to connect to ElasticSearch at %q. Skipping test.", ElasticSearchURL)
	}

	hostname := "testing"
	ruleName := "test_rule"

	client := cleanhttp.DefaultClient()

	randomUUID := func() string {
		id, err := uuid.GenerateUUID()
		if err != nil {
			t.Fatal(err)
		}
		return id
	}

	changeIndexFunc := func(method, index string, body io.Reader) {
		req, err := http.NewRequest(method, ElasticSearchURL+"/"+index, body)
		if err != nil {
			t.Fatal(err)
		}
		req.Header.Add("Content-Type", "application/json")

		_, err = client.Do(req)
		if err != nil {
			t.Fatal(err)
		}
	}

	stateIndex := randomUUID()

	cases := []struct {
		name     string
		stateURL string
		doc      string
		err      bool
	}{
		{
			"success",
			ElasticSearchURL + "/" + stateIndex,
			fmt.Sprintf(`{"next_query": %q, "hostname": %q, "rule_name": %q}`, time.Now().Format(time.RFC3339), hostname, ruleName),
			false,
		},
		{
			"url-parse-error",
			"@#$#!@#$%$#@#$%",
			"",
			true,
		},
		{
			"wrong-url",
			fmt.Sprintf("http://%s.com", randomUUID()),
			"",
			true,
		},
		{
			"non-string-timestamp",
			ElasticSearchURL + "/" + stateIndex,
			fmt.Sprintf(`{"next_query": 20, "hostname": %q, "rule_name": %q}`, hostname, ruleName),
			true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			changeIndexFunc("PUT", stateIndex, nil)
			defer changeIndexFunc("DELETE", stateIndex, nil)
			if tc.doc != "" {
				resp, err := client.Post(tc.stateURL+"/_doc", "application/json", bytes.NewBufferString(tc.doc))
				if err != nil {
					t.Fatal(err)
				}
				resp.Body.Close()

				time.Sleep(3 * time.Second)
			}

			qh := &QueryHandler{
				name:     ruleName,
				client:   cleanhttp.DefaultClient(),
				logger:   hclog.NewNullLogger(),
				stateURL: tc.stateURL,
				hostname: hostname,
			}

			ctx := context.Background()
			_, err := qh.getNextQuery(ctx)
			if tc.err {
				if err == nil {
					t.Fatal("expected an error but didn't receive one")
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestQueryError(t *testing.T) {
	qh, err := NewQueryHandler(&QueryHandlerConfig{
		Name:     "test_errors",
		Client:   cleanhttp.DefaultClient(),
		Logger:   hclog.NewNullLogger(),
		Schedule: "@every 10s",
		ESUrl:    "not a url!",
	})
	if err != nil {
		t.Fatal(err)
	}

	ctx := context.Background()
	_, err = qh.query(ctx)
	if err == nil {
		t.Fatal("expected an error but didn't receive one")
	}
}

func TestRun(t *testing.T) {
	if SkipElasticSearchTests {
		t.Skipf("unable to connect to ElasticSearch at %q. Skipping test.", ElasticSearchURL)
	}

	client := cleanhttp.DefaultClient()

	randomUUID := func() string {
		id, err := uuid.GenerateUUID()
		if err != nil {
			t.Fatal(err)
		}
		return id
	}

	changeIndexFunc := func(method, index string, body io.Reader) {
		req, err := http.NewRequest(method, ElasticSearchURL+"/"+index, body)
		if err != nil {
			t.Fatal(err)
		}
		req.Header.Add("Content-Type", "application/json")

		_, err = client.Do(req)
		if err != nil {
			t.Fatal(err)
		}
	}

	queryIndex := randomUUID()
	stateIndex := randomUUID()

	changeIndexFunc("PUT", queryIndex, nil)
	payload := bytes.NewBufferString(`{"hello": "world", "this": {"is": "a-test"}}`)
	changeIndexFunc("POST", queryIndex+"/_doc", payload)
	defer changeIndexFunc("DELETE", queryIndex, nil)
	defer changeIndexFunc("DELETE", stateIndex, nil)

	fileAM, err := file.NewFileAlertMethod(&file.FileAlertMethodConfig{
		OutputFilepath: "/tmp/testfile.log",
	})
	if err != nil {
		t.Fatal(err)
	}

	cases := []struct {
		name        string
		distributed bool
	}{
		{
			"distributed",
			true,
		},
		{
			"non-distributed",
			false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			qh, err := NewQueryHandler(&QueryHandlerConfig{
				Name:   "test_errors",
				Logger: hclog.NewNullLogger(),
				Client: client,
				ESUrl:  ElasticSearchURL,
				QueryData: map[string]interface{}{
					"query": map[string]interface{}{
						"term": map[string]interface{}{
							"hello": "world",
						},
					},
				},
				Distributed:  tc.distributed,
				QueryIndex:   queryIndex,
				Schedule:     "@every 10s",
				StateIndex:   stateIndex,
				AlertMethods: []alert.AlertMethod{fileAM},
			})
			if err != nil {
				t.Fatal(err)
			}
			if tc.distributed {
				qh.HaveLockCh <- true
			}

			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
			outputCh := make(chan *alert.Alert, 1)
			var wg sync.WaitGroup
			wg.Add(1)
			go qh.Run(ctx, outputCh, &wg)

			select {
			case <-ctx.Done():
				t.Fatal("context timeout")
			case a := <-outputCh:
				cancel()
				defer func() {
					wg.Wait()
				}()
				if a.RuleName != qh.name {
					t.Fatalf("bad alert rule name (expected: %q, got: %q)", qh.name, a.RuleName)
				}
				if len(a.Methods) != 1 {
					t.Fatal("alert should have just one alert method (file)")
				}
				if _, ok := a.Methods[0].(*file.FileAlertMethod); !ok {
					t.Fatalf("alert method should be of the FileAlertMethod type (got type: %T)", a.Methods[0])
				}
				if len(a.Records) != 1 {
					t.Fatalf("expected only one alert.Record (got %d records)", len(a.Records))
				}

				expected := "{\n    \"hello\": \"world\",\n    \"this\": {\n        \"is\": \"a-test\"\n    }\n}"
				if a.Records[0].Text != expected {
					t.Fatalf("unexpected alert.Record[0].Text value (got %q, expected %q)", a.Records[0].Text, expected)
				}
			}
		})
	}
}

func TestCanceledContext(t *testing.T) {
	qh, err := NewQueryHandler(&QueryHandlerConfig{
		Name:     "test_errors",
		Client:   cleanhttp.DefaultClient(),
		Logger:   hclog.NewNullLogger(),
		Schedule: "@every 10s",
	})
	if err != nil {
		t.Fatal(err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	outputCh := make(chan *alert.Alert, 1)
	doneCh := make(chan struct{})
	cancel()

	var wg sync.WaitGroup
	wg.Add(1)
	go qh.Run(ctx, outputCh, &wg)

	go func() {
		wg.Wait()
		doneCh <- struct{}{}
	}()

	select {
	case <-doneCh:
		return
	case <-time.After(10 * time.Second):
		t.Fatal("QueryHandler.Run() should immediately return when context canceled is passed to it")
	}
}

func TestMain(m *testing.M) {
	client := cleanhttp.DefaultClient()
	resp, err := client.Get(ElasticSearchURL)
	if err != nil {
		SkipElasticSearchTests = true
	} else {
		if resp.StatusCode != 200 {
			SkipElasticSearchTests = true
		}
	}
	os.Exit(m.Run())
}
