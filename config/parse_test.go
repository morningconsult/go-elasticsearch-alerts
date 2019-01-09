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

package config

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/hashicorp/vault/helper/jsonutil"
)

func TestParseConfig_MainConfig(t *testing.T) {
	cases := []struct {
		name string
		path string
		data interface{}
		err  bool
	}{
		{
			"success",
			"testdata/config.json",
			map[string]interface{}{
				"elasticsearch": map[string]interface{}{
					"server": map[string]interface{}{
						"url": "http://127.0.0.1:9200",
					},
				},
				"distributed": true,
				"consul": map[string]string{
					"consul_http_addr": "http://127.0.0.1:8500",
					"consul_lock_key":  "go-elasticsearch-alerts/leader",
				},
			},
			false,
		},
		{
			"homedir-error",
			"~testdata",
			map[string]interface{}{},
			true,
		},
		{
			"file-doesnt-exist",
			"testdata/config.json",
			map[string]interface{}{},
			true,
		},
		{
			"empty-file",
			"testdata/config.json",
			"not a json!",
			true,
		},
		{
			"no-elasticsearch-stanza",
			"testdata/config.json",
			map[string]interface{}{},
			true,
		},
		{
			"no-elasticsearch-server-field",
			"testdata/config.json",
			map[string]interface{}{
				"elasticsearch": map[string]interface{}{
					"client": map[string]interface{}{
						"tls_enabled": false,
					},
				},
			},
			true,
		},
		{
			"no-elasticsearch-server-url-field",
			"testdata/config.json",
			map[string]interface{}{
				"elasticsearch": map[string]interface{}{
					"server": map[string]interface{}{},
				},
			},
			true,
		},
		{
			"no-consul-field-when-distributed",
			"testdata/config.json",
			map[string]interface{}{
				"elasticsearch": map[string]interface{}{
					"server": map[string]interface{}{
						"url": "http://127.0.0.1:9200",
					},
				},
				"distributed": true,
			},
			true,
		},
		{
			"no-consul-addr-field-when-distributed",
			"testdata/config.json",
			map[string]interface{}{
				"elasticsearch": map[string]interface{}{
					"server": map[string]interface{}{
						"url": "http://127.0.0.1:9200",
					},
				},
				"distributed": true,
				"consul": map[string]string{
					"irrelevant": "key",
				},
			},
			true,
		},
		{
			"no-consul-lock-field-when-distributed",
			"testdata/config.json",
			map[string]interface{}{
				"elasticsearch": map[string]interface{}{
					"server": map[string]interface{}{
						"url": "http://127.0.0.1:9200",
					},
				},
				"distributed": true,
				"consul": map[string]string{
					"consul_http_addr": "http://127.0.0.1:8500",
				},
			},
			true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			os.Setenv(envConfigFile, tc.path)
			defer os.Unsetenv(envConfigFile)

			if tc.name == "success" {
				os.Setenv(envRulesDir, "testdata/rules-main")
				defer os.Unsetenv(envRulesDir)
			}

			if tc.name != "homedir-error" && tc.name != "file-doesnt-exist" {
				writeJSONToFile(t, tc.path, tc.data)
				defer os.Remove(tc.path)
			}

			cfg, err := ParseConfig()
			if tc.err {
				if err == nil {
					t.Fatal("expected an error but didn't receive one")
				}
				return
			}

			if err != nil {
				t.Fatal(err)
			}

			if cfg.Elasticsearch.Server.ElasticsearchURL != "http://127.0.0.1:9200" {
				t.Fatalf("got %q, expected \"http://127.0.0.1:9200\"", cfg.Elasticsearch.Server.ElasticsearchURL)
			}

			if !cfg.Distributed {
				t.Fatalf("got %t, expected true", cfg.Distributed)
			}

			v, ok := cfg.Consul["consul_http_addr"]
			if !ok {
				t.Fatal("config.Consul does not have key \"consul_http_addr\"")
			}
			if v != "http://127.0.0.1:8500" {
				t.Fatalf("config.Consul[\"consul_http_addr\"] unexpected value (got %q, expected \"http://127.0.0.1:8500\")", v)
			}

			l, ok := cfg.Consul["consul_lock_key"]
			if !ok {
				t.Fatal("config.Consul does not have key \"consul_lock_key\"")
			}
			if l != "go-elasticsearch-alerts/leader" {
				t.Fatalf("config.Consul[\"consul_lock_key\"] unexpected value (got %q, expected \"go-elasticsearch-alerts/leader\")", l)
			}
		})
	}
}

func TestParseConfig_Rules(t *testing.T) {
	type ruleFile struct {
		filename string
		data     interface{}
	}

	cases := []struct {
		name  string
		path  string
		files []*ruleFile
		err   bool
	}{
		{
			"homedir-error",
			"~testdata/rules",
			[]*ruleFile{},
			true,
		},
		{
			"not-a-json",
			"testdata/rules",
			[]*ruleFile{
				&ruleFile{
					"testrule-1.json",
					"not a json!",
				},
			},
			true,
		},
		{
			"homedir-error",
			"~testdata/rules",
			[]*ruleFile{},
			true,
		},
		{
			"malformed-json-string",
			"testdata/rules",
			[]*ruleFile{
				&ruleFile{
					"testrule-1.json",
					map[string]interface{}{
						"body": `{{"bad": "json"}}`,
					},
				},
			},
			true,
		},
		{
			"valid-json-string",
			"testdata/rules",
			[]*ruleFile{
				&ruleFile{
					"testrule-1.json",
					map[string]interface{}{
						"name":     "test",
						"body":     `{"ayy": "lmao"}`,
						"index":    "test-*",
						"schedule": "* * * * * *",
						"outputs": []interface{}{
							map[string]interface{}{
								"type": "file",
								"config": map[string]string{
									"file": "test.log",
								},
							},
						},
					},
				},
			},
			false,
		},
		{
			"unsupported-body-type",
			"testdata/rules",
			[]*ruleFile{
				&ruleFile{
					"testrule-1.json",
					map[string]interface{}{
						"body": 123,
					},
				},
			},
			true,
		},
		{
			"no-rule-name",
			"testdata/rules",
			[]*ruleFile{
				&ruleFile{
					"testrule-1.json",
					map[string]interface{}{
						"body": map[string]interface{}{
							"query": map[string]interface{}{
								"term": map[string]interface{}{
									"hostname": "test",
								},
							},
						},
					},
				},
			},
			true,
		},
		{
			"no-index",
			"testdata/rules",
			[]*ruleFile{
				&ruleFile{
					"testrule-1.json",
					map[string]interface{}{
						"name": "test-rule",
						"body": map[string]interface{}{
							"query": map[string]interface{}{
								"term": map[string]interface{}{
									"hostname": "test",
								},
							},
						},
					},
				},
			},
			true,
		},
		{
			"no-schedule",
			"testdata/rules",
			[]*ruleFile{
				&ruleFile{
					"testrule-1.json",
					map[string]interface{}{
						"name":  "test-rule",
						"index": "testindex",
						"body": map[string]interface{}{
							"query": map[string]interface{}{
								"term": map[string]interface{}{
									"hostname": "test",
								},
							},
						},
					},
				},
			},
			true,
		},
		{
			"no-output-field",
			"testdata/rules",
			[]*ruleFile{
				&ruleFile{
					"testrule-1.json",
					map[string]interface{}{
						"name":     "test-rule",
						"index":    "testindex",
						"schedule": "@every 1m",
						"body": map[string]interface{}{
							"query": map[string]interface{}{
								"term": map[string]interface{}{
									"hostname": "test",
								},
							},
						},
					},
				},
			},
			true,
		},
		{
			"no-outputs",
			"testdata/rules",
			[]*ruleFile{
				&ruleFile{
					"testrule-1.json",
					map[string]interface{}{
						"name":     "test-rule",
						"index":    "testindex",
						"schedule": "@every 1m",
						"body": map[string]interface{}{
							"query": map[string]interface{}{
								"term": map[string]interface{}{
									"hostname": "test",
								},
							},
						},
						"outputs": []interface{}{},
					},
				},
			},
			true,
		},
		{
			"output-missing-type",
			"testdata/rules",
			[]*ruleFile{
				&ruleFile{
					"testrule-1.json",
					map[string]interface{}{
						"name":     "test-rule",
						"index":    "testindex",
						"schedule": "@every 1m",
						"body": map[string]interface{}{
							"query": map[string]interface{}{
								"term": map[string]interface{}{
									"hostname": "test",
								},
							},
						},
						"outputs": []interface{}{
							map[string]interface{}{
								"type": "",
							},
						},
					},
				},
			},
			true,
		},
		{
			"output-missing-config",
			"testdata/rules",
			[]*ruleFile{
				&ruleFile{
					"testrule-1.json",
					map[string]interface{}{
						"name":     "test-rule",
						"index":    "testindex",
						"schedule": "@every 1m",
						"body": map[string]interface{}{
							"query": map[string]interface{}{
								"term": map[string]interface{}{
									"hostname": "test",
								},
							},
						},
						"outputs": []interface{}{
							map[string]interface{}{
								"type": "file",
							},
						},
					},
				},
			},
			true,
		},
		{
			"success",
			"testdata/rules",
			[]*ruleFile{
				&ruleFile{
					"testrule-1.json",
					map[string]interface{}{
						"name":     "test-rule",
						"index":    "testindex",
						"schedule": "@every 1m",
						"body": map[string]interface{}{
							"query": map[string]interface{}{
								"term": map[string]interface{}{
									"hostname": "test",
								},
							},
						},
						"outputs": []interface{}{
							map[string]interface{}{
								"type": "file",
								"config": map[string]string{
									"file": "test.log",
								},
							},
						},
					},
				},
			},
			false,
		},
		{
			"multiple-rules",
			"testdata/rules",
			[]*ruleFile{
				&ruleFile{
					"testrule-1.json",
					map[string]interface{}{
						"name":     "test-rule",
						"index":    "testindex",
						"schedule": "@every 1m",
						"body": map[string]interface{}{
							"query": map[string]interface{}{
								"term": map[string]interface{}{
									"hostname": "test",
								},
							},
						},
						"outputs": []interface{}{
							map[string]interface{}{
								"type": "file",
								"config": map[string]string{
									"file": "test.log",
								},
							},
						},
					},
				},
				&ruleFile{
					"testrule-2.json",
					map[string]interface{}{
						"name":     "test-rule-2",
						"index":    "testindex",
						"schedule": "@every 2m",
						"filters": []string{
							"aggregations.hostname.buckets",
							"aggregations.hostname.buckets.program.buckets",
						},
						"body": map[string]interface{}{
							"query": map[string]interface{}{
								"term": map[string]interface{}{
									"hostname": "test-2",
								},
							},
						},
						"outputs": []interface{}{
							map[string]interface{}{
								"type": "file",
								"config": map[string]string{
									"file": "test-2.log",
								},
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
			os.Setenv(envRulesDir, tc.path)
			defer os.Unsetenv(envRulesDir)

			for _, file := range tc.files {
				fname := filepath.Join(tc.path, file.filename)
				writeJSONToFile(t, fname, file.data)
				defer os.Remove(fname)
			}

			rules, err := ParseRules()
			if tc.err {
				if err == nil {
					t.Fatal("expected an error but didn't receive one")
				}
				return
			}

			if err != nil {
				t.Fatal(err)
			}

			if len(rules) != len(tc.files) {
				t.Fatalf("ParseRules() should have created one *RuleConfig per rule file (got %d, expected %d)",
					len(rules), len(tc.files))
			}

			for i, file := range tc.files {
				contents, ok := file.data.(map[string]interface{})
				if !ok {
					continue
				}

				name, ok := contents["name"].(string)
				if !ok {
					continue
				}

				if rules[i].Name != name {
					t.Fatalf("unexpected rule name (got %q, expected %q)", rules[i].Name, name)
				}

				index, ok := contents["index"].(string)
				if !ok {
					continue
				}

				if rules[i].ElasticsearchIndex != index {
					t.Fatalf("unexpected index value (got %q, expected %q)",
						rules[i].ElasticsearchIndex, index)
				}

				schedule, ok := contents["schedule"].(string)
				if !ok {
					continue
				}

				if rules[i].CronSchedule != schedule {
					t.Fatalf("unexpected schedule value (got %q, expected %q)",
						rules[i].CronSchedule, schedule)
				}

				filters, ok := contents["filters"].([]string)
				if !ok {
					continue
				}

				if len(rules[i].Filters) != len(filters) {
					t.Fatalf("returned rule has unexpected number of filters (got %d, expected %d)",
						len(rules[i].Filters), len(filters))
				}

				for j, filter := range filters {
					if filter != rules[i].Filters[j] {
						t.Fatalf("got unexpected filter (got %q, expected %q)",
							rules[i].Filters[j], filter)
					}
				}

				body, ok := contents["body"].(map[string]interface{})
				if !ok {
					continue
				}

				if !reflect.DeepEqual(body, rules[i].ElasticsearchBody) {
					t.Fatalf("rule 'body' is unexpected:\nGot:\n\t%+v\nExpected:\n\t%+v",
						rules[i].ElasticsearchBody, body)
				}

				outputs, ok := contents["outputs"].([]interface{})
				if !ok {
					continue
				}
				if len(rules[i].Outputs) != len(outputs) {
					t.Fatalf("rule 'outputs' is an unexpected length (got %d, expected %d)",
						len(rules[i].Outputs), len(outputs))
				}
			}
		})
	}
}

func writeJSONToFile(t *testing.T, path string, contents interface{}) {
	data, err := jsonutil.EncodeJSON(contents)
	if err != nil {
		t.Fatal(err)
	}

	if err = ioutil.WriteFile(path, data, 0666); err != nil {
		t.Fatal(err)
	}
}
