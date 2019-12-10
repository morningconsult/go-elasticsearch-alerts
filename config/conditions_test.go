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
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	hclog "github.com/hashicorp/go-hclog"
)

func TestCondition_validate(t *testing.T) {
	cases := []struct {
		name      string
		condition Condition
		expectErr string
	}{
		{
			name: "success-numbers",
			condition: Condition{
				"field":      "foo.bar.bim.baz",
				"quantifier": "all",
				"gt":         json.Number("9"),
				"lt":         json.Number("21"),
				"ge":         json.Number("10"),
				"le":         json.Number("20"),
				"eq":         json.Number("15"),
				"ne":         json.Number("16"),
			},
			expectErr: "",
		},
		{
			name: "success-strings",
			condition: Condition{
				"field":      "foo.bar.bim.baz",
				"quantifier": "all",
				"eq":         "foo",
				"ne":         "bar",
			},
			expectErr: "",
		},
		{
			name:      "no-field",
			condition: Condition{},
			expectErr: "1 error occurred:\n\t* condition must have the field 'field'\n\n",
		},
		{
			name:      "field-is-empty",
			condition: Condition{"field": ""},
			expectErr: "1 error occurred:\n\t* field 'field' of condition must not be empty\n\n",
		},
		{
			name:      "field-not-string",
			condition: Condition{"field": 100},
			expectErr: "1 error occurred:\n\t* field 'field' of condition must not be empty\n\n",
		},
		{
			name: "bad-quantifier",
			condition: Condition{
				"field":      "foo.bar.bim.baz",
				"quantifier": "lmao",
			},
			expectErr: "1 error occurred:\n\t* field 'quantifier' of condition must either be 'any', 'all', or 'none'\n\n",
		},
		{
			name: "quantifier-not-string",
			condition: Condition{
				"field":      "foo.bar.bim.baz",
				"quantifier": 100,
			},
			expectErr: "1 error occurred:\n\t* field 'quantifier' of condition must be a string\n\n",
		},
		{
			name: "numeric-type-errors",
			condition: Condition{
				"field":      "foo.bar.bim.baz",
				"quantifier": "any",
				"ge":         "asdf",
				"gt":         10, // not a json.Number
				"le":         true,
				"lt":         map[string]interface{}{"ayy": "lmao"},
			},
			expectErr: `4 errors occurred:
	* value of operator 'le' should be a number
	* value of operator 'lt' should be a number
	* value of operator 'gt' should be a number
	* value of operator 'ge' should be a number

`,
		},
		{
			name: "all-type-errors",
			condition: Condition{
				"field":      "foo.bar.bim.baz",
				"quantifier": "any",
				"ge":         "asdf",
				"gt":         10, // not a json.Number
				"le":         true,
				"lt":         map[string]interface{}{"ayy": "lmao"},
				"eq":         true, // not a json.Number or string
				"ne":         100,  // not a json.Number or string
			},
			expectErr: `6 errors occurred:
	* value of operator 'le' should be a number
	* value of operator 'lt' should be a number
	* value of operator 'gt' should be a number
	* value of operator 'ge' should be a number
	* value of operator 'eq' should either be a number or a string
	* value of operator 'ne' should either be a number or a string

`,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.condition.validate()
			if tc.expectErr != "" {
				if err == nil {
					t.Fatal("Expected an error")
				}
				gotErr := err.Error()
				if tc.expectErr != gotErr {
					t.Errorf("Expected error:\n%q\nGot error:\n%q", tc.expectErr, gotErr)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestConditionsMet(t *testing.T) {
	defaultJSONResponse := []byte(`{
  "took" : 4863,
  "timed_out" : false,
  "_shards" : {
    "total" : 7,
    "successful" : 7,
    "skipped" : 0,
    "failed" : 0
  },
  "hits" : {
    "total" : {
      "value" : 376,
      "relation" : "eq"
    },
    "max_score" : null,
    "hits" : [ ]
  },
  "aggregations" : {
    "pipelines" : {
      "doc_count" : 3384,
      "queue" : {
        "doc_count_error_upper_bound" : 0,
        "sum_other_doc_count" : 0,
        "buckets" : [
          {
            "key" : "main",
            "doc_count" : 1128,
            "queue_size" : {
              "value" : 3.8679811209E10
            },
            "max_queue_size" : {
              "value" : 1.211180777472E13
            },
            "queue_usage" : {
              "value" : 0.3193562177376465
            }
          },
          {
            "key" : "message-queues",
            "doc_count" : 1128,
            "queue_size" : {
              "value" : 3.632980257E9
            },
            "max_queue_size" : {
              "value" : 1.211180777472E13
            },
            "queue_usage" : {
              "value" : 0.029995359277273426
	    },
	    "queue_empty" : {
	      "value" : true
	    }
          }
        ]
      }
    }
  }
}`)

	cases := []struct {
		name       string
		json       []byte
		conditions []Condition
		expectRes  bool
		expectLogs []string
	}{
		{
			name: "one-any-condition-satisfied",
			json: defaultJSONResponse,
			conditions: []Condition{
				{
					"field":      "aggregations.pipelines.queue.buckets.queue_usage.value",
					"type":       "number",
					"quantifier": "any",
					"ge":         json.Number("0.1"),
					"le":         json.Number("0.5"),
				},
			},
			expectRes: true,
		},
		{
			name: "one-all-condition-satisfied",
			json: defaultJSONResponse,
			conditions: []Condition{
				{
					"field":      "aggregations.pipelines.queue.buckets.queue_usage.value",
					"type":       "number",
					"quantifier": "all",
					"ge":         json.Number("0.01"),
					"le":         json.Number("0.5"),
				},
			},
			expectRes: true,
		},
		{
			name: "one-none-condition-satisfied",
			json: defaultJSONResponse,
			conditions: []Condition{
				{
					"field":      "aggregations.pipelines.queue.buckets.queue_usage.value",
					"type":       "number",
					"quantifier": "none",
					"ge":         json.Number("0.4"),
				},
			},
			expectRes: true,
		},
		{
			name: "no-any-condition-satisfied",
			json: defaultJSONResponse,
			conditions: []Condition{
				{
					"field":      "aggregations.pipelines.queue.buckets.queue_usage.value",
					"type":       "number",
					"quantifier": "any",
					"ge":         json.Number("0.5"),
					"le":         json.Number("0.1"),
				},
			},
			expectRes: false,
		},
		{
			name: "multiple-conditions-satisfied",
			json: defaultJSONResponse,
			conditions: []Condition{
				{
					"field":      "aggregations.pipelines.queue.buckets.queue_usage.value",
					"quantifier": "any",
					"ge":         json.Number("0.1"),
					"le":         json.Number("0.5"),
				},
				{
					"field":      "aggregations.pipelines.queue.buckets.key",
					"quantifier": "any",
					"eq":         "message-queues",
				},
				{
					"field":      "aggregations.pipelines.queue.buckets.queue_empty.value",
					"quantifier": "any",
					"eq":         true,
				},
			},
			expectRes: true,
		},
		{
			name: "one-condition-satisfied",
			json: defaultJSONResponse,
			conditions: []Condition{
				{
					"field":      "aggregations.pipelines.queue.buckets.queue_usage.value",
					"quantifier": "any",
					"ge":         json.Number("0.1"),
					"le":         json.Number("0.5"),
				},
				{
					"field":      "aggregations.pipelines.queue.buckets.key",
					"quantifier": "any",
					"eq":         "ayylmao", // not satisfied
				},
			},
			expectRes: false,
		},
		{
			name: "field-type-mismatch",
			json: defaultJSONResponse,
			conditions: []Condition{
				{
					"field":      "aggregations.pipelines.queue.buckets.queue_usage", // not a primitive
					"quantifier": "any",
					"eq":         "ayylmao",
				},
			},
			expectRes: true,
			expectLogs: []string{
				"[ERROR] Value of field in Elasticsearch response is not a string, number, or boolean. Ignoring condition for this value: field=aggregations.pipelines.queue.buckets.queue_usage",
			},
		},
		{
			name: "unsupported-string-operator",
			json: defaultJSONResponse,
			conditions: []Condition{
				{
					"field":      "aggregations.pipelines.queue.buckets.key",
					"quantifier": "any",
					"gt":         json.Number("0.1"), // cannot use greater than on a string
				},
			},
			expectRes: true,
		},
		{
			name: "unsupported-bool-operator",
			json: defaultJSONResponse,
			conditions: []Condition{
				{
					"field":      "aggregations.pipelines.queue.buckets.queue_empty.value",
					"quantifier": "any",
					"gt":         json.Number("0.1"), // cannot use greater than on a bool
				},
			},
			expectRes: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			buf := bytes.NewBuffer(tc.json)
			dec := json.NewDecoder(buf)
			dec.UseNumber()

			var res map[string]interface{}
			if err := dec.Decode(&res); err != nil {
				t.Fatal(err)
			}

			out := bytes.Buffer{}
			logger := hclog.New(&hclog.LoggerOptions{Output: &out})
			got := ConditionsMet(logger, res, tc.conditions)
			if got != tc.expectRes {
				t.Errorf("Expected conditions to be met? %t\nWere conditions met? %t", tc.expectRes, got)
			}
			if len(tc.expectLogs) > 0 {
				gotLogs := out.String()
				for _, log := range tc.expectLogs {
					if !strings.Contains(gotLogs, log) {
						t.Errorf("Error logs do not contain log:\n%s", log)
					}
				}
			}
		})
	}
}
