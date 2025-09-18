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

package query

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"reflect"
	"sync"
	"testing"
	"time"

	cleanhttp "github.com/hashicorp/go-cleanhttp"
	hclog "github.com/hashicorp/go-hclog"
	uuid "github.com/hashicorp/go-uuid"

	"github.com/morningconsult/go-elasticsearch-alerts/command/alert"
	"github.com/morningconsult/go-elasticsearch-alerts/command/alert/file"
	"github.com/morningconsult/go-elasticsearch-alerts/internal/lock"
)

const (
	ElasticsearchURL string = "http://127.0.0.1:9200"
	badURL           string = "bad-url"
)

func TestNewQueryHandler(t *testing.T) {
	cases := []struct {
		name   string
		config *QueryHandlerConfig
		err    bool
		errMsg string
	}{
		{
			"success",
			&QueryHandlerConfig{
				Name:         "Test Errors",
				ESUrl:        ElasticsearchURL,
				QueryIndex:   "test-*",
				AlertMethods: []alert.Method{&file.AlertMethod{}},
				QueryData: map[string]any{
					"ayy": "lmao",
				},
				Schedule: "@every 10m",
			},
			false,
			"",
		},
		{
			"no-name",
			&QueryHandlerConfig{
				ESUrl:        ElasticsearchURL,
				QueryIndex:   "test-*",
				AlertMethods: []alert.Method{&file.AlertMethod{}},
				QueryData: map[string]any{
					"ayy": "lmao",
				},
				Schedule: "@every 10m",
			},
			true,
			"1 error occurred:\n\t* no rule name provided\n\n",
		},
		{
			"no-es-url",
			&QueryHandlerConfig{
				Name:         "Test Errors",
				QueryIndex:   "test-*",
				AlertMethods: []alert.Method{&file.AlertMethod{}},
				QueryData: map[string]any{
					"ayy": "lmao",
				},
				Schedule: "@every 10m",
			},
			true,
			"1 error occurred:\n\t* no Elasticsearch URL provided\n\n",
		},
		{
			"no-query-index",
			&QueryHandlerConfig{
				Name:         "Test Errors",
				ESUrl:        ElasticsearchURL,
				AlertMethods: []alert.Method{&file.AlertMethod{}},
				QueryData: map[string]any{
					"ayy": "lmao",
				},
				Schedule: "@every 10m",
			},
			true,
			"1 error occurred:\n\t* no Elasticsearch index provided\n\n",
		},
		{
			"no-alert-methods",
			&QueryHandlerConfig{
				Name:         "Test Errors",
				ESUrl:        ElasticsearchURL,
				QueryIndex:   "test-*",
				AlertMethods: []alert.Method{},
				QueryData: map[string]any{
					"ayy": "lmao",
				},
				Schedule: "@every 10m",
			},
			true,
			"1 error occurred:\n\t* at least one alert method must be specified\n\n",
		},
		{
			"cron-parse-error",
			&QueryHandlerConfig{
				Name:         "Test Errors",
				ESUrl:        ElasticsearchURL,
				QueryIndex:   "test-*",
				AlertMethods: []alert.Method{&file.AlertMethod{}},
				QueryData: map[string]any{
					"ayy": "lmao",
				},
				Schedule: "blah",
			},
			true,
			"error parsing cron schedule: Expected 5 to 6 fields, found 1: blah",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := NewQueryHandler(tc.config)
			if tc.err {
				if err == nil {
					t.Fatal("expected an error but didn't receive one")
				}
				if err.Error() != tc.errMsg {
					t.Fatalf("Expected error:\n\t%q\nGot:\n\t%q\n", tc.errMsg, err.Error())
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestBuildHTTPRequestFunc(t *testing.T) {
	t.Run("basic-auth", func(t *testing.T) {
		username := "foo@bar.com"
		password := "baz"

		t.Setenv(envESBasicAuthUsername, username)

		t.Setenv(envESBasicAuthPassword, password)

		reqFunc, err := buildHTTPRequestFunc()
		if err != nil {
			t.Fatal(err)
		}
		req, err := reqFunc(t.Context(), http.MethodGet, "http://example.com", nil)
		if err != nil {
			t.Fatal(err)
		}
		gotUsername, gotPassword, ok := req.BasicAuth()
		if !ok {
			t.Fatal("Basic auth should be enabled")
		}
		if gotUsername != username {
			t.Errorf("Expected basic auth username %q, got username %q", username, gotUsername)
		}
		if gotPassword != password {
			t.Errorf("Expected basic auth password %q, got password %q", password, gotPassword)
		}
	})

	t.Run("username-and-password-not-both-set", func(t *testing.T) {
		username := "foo@bar.com"

		t.Setenv(envESBasicAuthUsername, username)

		t.Setenv(envESBasicAuthPassword, "")
		os.Unsetenv(envESBasicAuthPassword)

		_, err := buildHTTPRequestFunc()
		if err == nil {
			t.Fatal("Expected an error")
		}
		gotError := err.Error()
		expectError := fmt.Sprintf(
			"both %s and %s should be set when using basic auth",
			envESBasicAuthUsername,
			envESBasicAuthPassword,
		)
		if gotError != expectError {
			t.Errorf("Expected error:\n%s\nGot error:\n%s", expectError, gotError)
		}
	})
}

func TestPutTemplate(t *testing.T) {
	reqFunc, err := buildHTTPRequestFunc()
	if err != nil {
		t.Fatal(err)
	}
	cases := []struct {
		name   string
		status int
		data   any
		err    bool
	}{
		{badURL, 200, "lol", true},
		{"non-200-response", 500, "", true},
		{"non-json-response", 200, "not a json!!", true},
		{"no-acknowledged-field", 200, map[string]any{"ayy": "lmao"}, true},
		{"non-bool-acknowledged-field", 200, map[string]any{"acknowledged": "true"}, true},
		{"success", 200, map[string]any{"acknowledged": true}, false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				switch r.Method {
				case "PUT", "POST":
					if tc.status-200 > 3 {
						http.Error(w, http.StatusText(tc.status), tc.status)
						return
					}
					w.WriteHeader(tc.status)
					var data []byte
					var err error
					switch v := tc.data.(type) {
					case map[string]any:
						data, err = json.Marshal(v)
						if err != nil {
							t.Fatal(err)
						}
					case string:
						data = []byte(v)
					case []byte:
						data = v
					default:
						t.Fatalf("unsupported data type: %T", v)
					}
					w.Write(data)
				default:
					http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
				}
			}))
			defer ts.Close()

			u := ts.URL
			if tc.name == badURL {
				u = fmt.Sprintf("http://example.%s.co.nz", randomUUID(t))
			}
			qh := &QueryHandler{
				client:     cleanhttp.DefaultClient(),
				esURL:      u,
				newRequest: reqFunc,
			}

			err := qh.PutTemplate(t.Context())
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

func TestGetNextQuery(t *testing.T) {
	expected := time.Now().Add(2 * time.Hour).Format(time.RFC3339)
	cases := []struct {
		name   string
		status int
		data   any
		err    bool
	}{
		{
			badURL,
			200,
			"lol",
			true,
		},
		{
			"non-200-response",
			500,
			"",
			true,
		},
		{
			"non-json-response",
			200,
			"not a json!!",
			true,
		},
		{
			"no-hits-field",
			200,
			map[string]any{
				"ayy": "lmao",
			},
			true,
		},
		{
			"non-string-next-query-field",
			200,
			map[string]any{
				"hits": map[string]any{
					"hits": []any{
						map[string]any{
							"_source": map[string]any{
								"next_query": map[string]any{
									"ayy": "lmao",
								},
							},
						},
					},
				},
			},
			true,
		},
		{
			"non-timestamp-next-query-field",
			200,
			map[string]any{
				"hits": map[string]any{
					"hits": []any{
						map[string]any{
							"_source": map[string]any{
								"next_query": "not a timestamp!!!",
							},
						},
					},
				},
			},
			true,
		},
		{
			"success",
			200,
			map[string]any{
				"hits": map[string]any{
					"hits": []any{
						map[string]any{
							"_source": map[string]any{
								"next_query": expected,
							},
						},
					},
				},
			},
			false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				switch r.Method {
				case "GET":
					if tc.status-200 > 3 {
						http.Error(w, http.StatusText(tc.status), tc.status)
						return
					}
					w.WriteHeader(tc.status)
					var data []byte
					var err error
					switch v := tc.data.(type) {
					case map[string]any:
						data, err = json.Marshal(v)
						if err != nil {
							t.Fatal(err)
						}
					case string:
						data = []byte(v)
					case []byte:
						data = v
					default:
						t.Fatalf("unsupported data type: %T", v)
					}
					w.Write(data)
				default:
					http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
				}
			}))
			defer ts.Close()

			u := ts.URL
			if tc.name == badURL {
				u = fmt.Sprintf("http://example.%s.co.nz", randomUUID(t))
			}
			qh, err := NewQueryHandler(&QueryHandlerConfig{
				Name:         "Test Errors",
				ESUrl:        u,
				QueryIndex:   "test-*",
				AlertMethods: []alert.Method{&file.AlertMethod{}},
				QueryData: map[string]any{
					"hello": "world",
				},
				Schedule: "@every 10m",
			})
			if err != nil {
				t.Fatal(err)
			}

			timestamp, err := qh.getNextQuery(t.Context())
			if tc.err {
				if err == nil {
					t.Fatal("expected an error but didn't receive one")
				}
				// t.Log(err.Error())
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if timestamp.Format(time.RFC3339) != expected {
				t.Fatalf("Got timestamp %q, expected %q", timestamp.Format(time.RFC3339), expected)
			}
		})
	}
}

func TestRun(t *testing.T) {
	queryIndex := randomUUID(t)
	expected := map[string]any{
		"hits": map[string]any{
			"hits": []any{
				map[string]any{
					"_source": map[string]any{
						"hello": "world",
					},
				},
			},
		},
	}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Mocks successful response to QueryHandler.putTemplate()
		if r.URL.Path == fmt.Sprintf("/_template/%s-%s", defaultStateIndexAlias, templateVersion) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(200)
			w.Write([]byte(`{"acknowledged": true}`))
			return
		}

		// Mocks successful response to QueryHandler.getNextQuery()
		if r.URL.Path == fmt.Sprintf("/%s-%s/_search", defaultStateIndexAlias, templateVersion) {
			payload := fmt.Sprintf(
				`{"hits":{"hits":[{"_source":{"next_query":%q}}]}}`,
				time.Now().Add(-1*time.Hour).Format(time.RFC3339),
			)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(200)
			w.Write([]byte(payload))
			return
		}

		// Mocks successful response to QueryHandler.setNextQuery()
		if r.URL.Path == fmt.Sprintf("/<%s-status-%s-{now/d}>/_doc", defaultStateIndexAlias, templateVersion) {
			w.WriteHeader(201)
			w.Write([]byte("ok"))
			return
		}

		// Mocks successful response to QueryHandler.query()
		if r.URL.Path == fmt.Sprintf("/%s/_search", queryIndex) {
			w.Header().Set("Content-Type", "application/json")
			if err := json.NewEncoder(w).Encode(&expected); err != nil {
				http.Error(w, err.Error(), 500)
			}
		}
	}))
	defer ts.Close()

	filename := filepath.Join("testdata", "testfile.log")
	fileAM, err := file.NewAlertMethod(&file.AlertMethodConfig{
		OutputFilepath: filename,
	})
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(filename)

	qh, err := NewQueryHandler(&QueryHandlerConfig{
		Name:         "Test Errors",
		Logger:       hclog.NewNullLogger(),
		ESUrl:        ts.URL,
		QueryIndex:   queryIndex,
		AlertMethods: []alert.Method{fileAM},
		QueryData: map[string]any{
			"query": map[string]any{
				"term": map[string]any{
					"hello": "world",
				},
			},
		},
		Schedule: "@every 10s",
	})
	if err != nil {
		t.Fatal(err)
	}

	var wg sync.WaitGroup
	ctx, cancel := context.WithTimeout(t.Context(), 20*time.Second)
	defer cancel()

	outputCh := make(chan *alert.Alert, 1)
	lock := lock.NewLock()
	lock.Set(true)
	wg.Add(1)

	go qh.Run(ctx, outputCh, &wg, lock)

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
		if _, ok := a.Methods[0].(*file.AlertMethod); !ok {
			t.Fatalf("alert method should be of the AlertMethod type (got type: %T)", a.Methods[0])
		}
		if len(a.Records) != 1 {
			t.Fatalf("expected only one alert.Record (got %d records)", len(a.Records))
		}

		expected := "{\n    \"hello\": \"world\"\n}"
		if a.Records[0].Text != expected {
			t.Fatalf("unexpected alert.Record[0].Text value (got %q, expected %q)", a.Records[0].Text, expected)
		}
	}
}

func TestSetNextQuery(t *testing.T) {
	cases := []struct {
		name   string
		status int
		err    bool
	}{
		{
			"success",
			201,
			false,
		},
		{
			"non-200-response",
			500,
			true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				switch r.Method {
				case "POST", "PUT":
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(tc.status)
					w.Write([]byte(`{"acknowledged": true}`))
				default:
					http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
				}
			}))
			defer ts.Close()

			qh, err := NewQueryHandler(&QueryHandlerConfig{
				Name:         "Test Errors",
				Logger:       hclog.NewNullLogger(),
				ESUrl:        ts.URL,
				QueryIndex:   "test-*",
				AlertMethods: []alert.Method{&file.AlertMethod{}},
				QueryData: map[string]any{
					"query": "test",
				},
				Schedule: "@every 10s",
			})
			if err != nil {
				t.Fatal(err)
			}

			err = qh.setNextQuery(t.Context(), time.Now().Add(1*time.Hour), nil)
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

func TestQuery(t *testing.T) {
	expected := map[string]any{"some": "data"}
	cases := []struct {
		name   string
		status int
		data   any
		err    bool
	}{
		{badURL, 200, "lol", true},
		{"non-200-response", 500, "", true},
		{"non-json-response", 200, "not a json!!", true},
		{"success", 200, expected, false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ts := newTestServer(tc.status, tc.data)
			defer ts.Close()

			u := ts.URL
			if tc.name == badURL {
				u = fmt.Sprintf("http://example.%s.co.nz", randomUUID(t))
			}
			qh, err := NewQueryHandler(&QueryHandlerConfig{
				Name:         "Test Errors",
				ESUrl:        u,
				QueryIndex:   "test-*",
				AlertMethods: []alert.Method{&file.AlertMethod{}},
				QueryData: map[string]any{
					"hello": "world",
				},
				Schedule: "@every 10m",
			})
			if err != nil {
				t.Fatal(err)
			}

			data, err := qh.query(t.Context())
			if tc.err {
				if err == nil {
					t.Fatal("expected an error but didn't receive one")
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(data, tc.data) {
				t.Fatalf("unexpected data received. Expected:\n\t%+v\nGot:\n\t%+v\n", tc.data, data)
			}
		})
	}
}

func TestNewRequestErrors(t *testing.T) {
	reqFunc, err := buildHTTPRequestFunc()
	if err != nil {
		t.Fatal(err)
	}
	cases := []struct {
		name    string
		method  string
		payload []byte
		err     bool
	}{
		{"bad-method-with-data", "ASDFASDF ASD", []byte("some data"), true},
		{"bad-method-without-data", "ASDFADS FASD ", nil, true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			qh := &QueryHandler{newRequest: reqFunc}
			_, err := qh.newRequest(t.Context(), tc.method, "http://example.com", bytes.NewBuffer(tc.payload))
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

func TestCanceledContext(t *testing.T) {
	qh, err := NewQueryHandler(&QueryHandlerConfig{
		Name:         "Test Errors",
		Logger:       hclog.NewNullLogger(),
		ESUrl:        "http://example.com",
		QueryIndex:   "test-*",
		AlertMethods: []alert.Method{&file.AlertMethod{}},
		QueryData: map[string]any{
			"hello": "world",
		},
		Schedule: "@every 10m",
	})
	if err != nil {
		t.Fatal(err)
	}

	ctx, cancel := context.WithCancel(t.Context())
	outputCh := make(chan *alert.Alert, 1)
	doneCh := make(chan struct{})
	cancel()

	var wg sync.WaitGroup
	wg.Add(1)
	lock := lock.NewLock()
	lock.Set(true)
	go qh.Run(ctx, outputCh, &wg, lock)

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

func randomUUID(t *testing.T) string {
	id, err := uuid.GenerateUUID()
	if err != nil {
		t.Fatal(err)
	}
	return id
}

func newTestServer(status int, payload any) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "GET":
			if status-200 > 3 {
				http.Error(w, http.StatusText(status), status)
				return
			}
			w.WriteHeader(status)
			var data []byte
			var err error
			switch v := payload.(type) {
			case map[string]any:
				data, err = json.Marshal(v)
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
			case string:
				data = []byte(v)
			case []byte:
				data = v
			default:
				http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
				return
			}
			w.Write(data)
		default:
			http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		}
	}))
}
