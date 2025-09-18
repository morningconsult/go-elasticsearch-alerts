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
	"encoding/json"
	"testing"

	"github.com/google/go-cmp/cmp"
	hclog "github.com/hashicorp/go-hclog"

	"github.com/morningconsult/go-elasticsearch-alerts/command/alert"
	"github.com/morningconsult/go-elasticsearch-alerts/config"
)

func TestProcess(t *testing.T) {
	cases := []struct {
		name       string
		input      map[string]any
		filters    []string
		conditions []config.Condition
		output     []*alert.Record
		hits       int
		err        bool
	}{
		{
			name: "one-level",
			input: map[string]any{
				"aggregations": map[string]any{
					"hostname": map[string]any{
						"buckets": []any{
							map[string]any{
								"key":       "foo",
								"doc_count": json.Number("2"),
							},
							map[string]any{
								"key":       "bar",
								"doc_count": json.Number("3"),
							},
						},
					},
				},
			},
			filters: []string{"aggregations.hostname.buckets"},
			output: []*alert.Record{
				{
					Filter: "aggregations.hostname.buckets",
					Fields: []*alert.Field{
						{
							Key:   "foo",
							Count: 2,
						},
						{
							Key:   "bar",
							Count: 3,
						},
					},
				},
			},
			hits: 0,
			err:  false,
		},
		{
			name: "field-not-map",
			input: map[string]any{
				"aggregations": map[string]any{
					"hostname": map[string]any{
						"buckets": []any{
							"string",
							map[string]any{
								"key":       "bar",
								"doc_count": json.Number("3"),
							},
						},
					},
				},
			},
			filters: []string{"aggregations.hostname.buckets"},
			output: []*alert.Record{
				{
					Filter: "aggregations.hostname.buckets",
					Fields: []*alert.Field{
						{
							Key:   "bar",
							Count: 3,
						},
					},
				},
			},
			hits: 0,
			err:  false,
		},
		{
			name: "zero-count",
			input: map[string]any{
				"aggregations": map[string]any{
					"hostname": map[string]any{
						"buckets": []any{
							map[string]any{
								"key":       "foo",
								"doc_count": json.Number("0"),
							},
							map[string]any{
								"key":       "bar",
								"doc_count": json.Number("3"),
							},
						},
					},
				},
			},
			filters: []string{"aggregations.hostname.buckets"},
			output: []*alert.Record{
				{
					Filter: "aggregations.hostname.buckets",
					Fields: []*alert.Field{
						{
							Key:   "bar",
							Count: 3,
						},
					},
				},
			},
			hits: 0,
			err:  false,
		},
		{
			name: "two-levels",
			input: map[string]any{
				"aggregations": map[string]any{
					"hostname": map[string]any{
						"buckets": []any{
							map[string]any{
								"key":       "foo",
								"doc_count": 5,
								"program": map[string]any{
									"buckets": []any{
										map[string]any{
											"key":       "bim",
											"doc_count": json.Number("2"),
										},
										map[string]any{
											"key":       "baz",
											"doc_count": json.Number("3"),
										},
									},
								},
							},
							map[string]any{
								"key":       "bar",
								"doc_count": 3,
								"program": map[string]any{
									"buckets": []any{
										map[string]any{
											"key":       "ayy",
											"doc_count": json.Number("1"),
										},
										map[string]any{
											"key":       "lmao",
											"doc_count": json.Number("2"),
										},
									},
								},
							},
						},
					},
				},
			},
			filters: []string{"aggregations.hostname.buckets.program.buckets"},
			output: []*alert.Record{
				{
					Filter: "aggregations.hostname.buckets.program.buckets",
					Fields: []*alert.Field{
						{
							Key:   "foo - bim",
							Count: 2,
						},
						{
							Key:   "foo - baz",
							Count: 3,
						},
						{
							Key:   "bar - ayy",
							Count: 1,
						},
						{
							Key:   "bar - lmao",
							Count: 2,
						},
					},
				},
			},
			hits: 0,
			err:  false,
		},
		{
			name: "hits-not-array",
			input: map[string]any{
				"aggregations": map[string]any{
					"hostname": map[string]any{
						"buckets": []any{
							map[string]any{
								"key":       "foo",
								"doc_count": json.Number("2"),
							},
							map[string]any{
								"key":       "bar",
								"doc_count": json.Number("3"),
							},
						},
					},
				},
				"hits": map[string]any{
					"hits": map[string]any{
						"ayy": "lmao",
					},
				},
			},
			filters: []string{"aggregations.hostname.buckets"},
			output: []*alert.Record{
				{
					Filter: "aggregations.hostname.buckets",
					Fields: []*alert.Field{
						{
							Key:   "foo",
							Count: 2,
						},
						{
							Key:   "bar",
							Count: 3,
						},
					},
				},
			},
			hits: 0,
			err:  false,
		},
		{
			name: "hit-elems-not-maps",
			input: map[string]any{
				"aggregations": map[string]any{
					"hostname": map[string]any{
						"buckets": []any{
							map[string]any{
								"key":       "foo",
								"doc_count": json.Number("2"),
							},
							map[string]any{
								"key":       "bar",
								"doc_count": json.Number("3"),
							},
						},
					},
				},
				"hits": map[string]any{
					"hits": []any{
						"sadly",
						"i",
						"am",
						"only",
						"a",
						"string",
					},
				},
			},
			filters: []string{"aggregations.hostname.buckets"},
			output: []*alert.Record{
				{
					Filter: "aggregations.hostname.buckets",
					Fields: []*alert.Field{
						{
							Key:   "foo",
							Count: 2,
						},
						{
							Key:   "bar",
							Count: 3,
						},
					},
				},
			},
			hits: 0,
			err:  false,
		},
		{
			name: "hit-elems-have-no-source",
			input: map[string]any{
				"aggregations": map[string]any{
					"hostname": map[string]any{
						"buckets": []any{
							map[string]any{
								"key":       "foo",
								"doc_count": json.Number("2"),
							},
							map[string]any{
								"key":       "bar",
								"doc_count": json.Number("3"),
							},
						},
					},
				},
				"hits": map[string]any{
					"hits": []any{
						map[string]any{
							"any": "field",
							"but": "_source!",
						},
						map[string]any{
							"_source": map[string]any{
								"ayy": "lmao",
							},
						},
					},
				},
			},
			filters: []string{"aggregations.hostname.buckets"},
			output: []*alert.Record{
				{
					Filter: "aggregations.hostname.buckets",
					Fields: []*alert.Field{
						{
							Key:   "foo",
							Count: 2,
						},
						{
							Key:   "bar",
							Count: 3,
						},
					},
				},
				{
					Filter:    "hits.hits._source",
					Text:      "{\n    \"ayy\": \"lmao\"\n}",
					BodyField: true,
				},
			},
			hits: 1,
			err:  false,
		},
		{
			name: "hits-only",
			input: map[string]any{
				"hits": map[string]any{
					"hits": []any{
						map[string]any{
							"_source": map[string]any{
								"ayy": "lmao",
							},
						},
						map[string]any{
							"_source": map[string]any{
								"yeah": "buddy",
							},
						},
					},
				},
			},
			filters: []string{},
			output: []*alert.Record{
				{
					Filter: "hits.hits._source",
					Text: `{
    "ayy": "lmao"
}
----------------------------------------
{
    "yeah": "buddy"
}`,
					BodyField: true,
				},
			},
			hits: 2,
			err:  false,
		},
		{
			name: "no-hits-no-filters",
			input: map[string]any{
				"hits": map[string]any{
					"hits": []any{},
				},
			},
			filters: []string{},
			output:  []*alert.Record{},
			hits:    0,
			err:     false,
		},
		{
			name: "conditions-not-met",
			input: map[string]any{
				"aggregations": map[string]any{
					"hostname": map[string]any{
						"buckets": []any{
							map[string]any{
								"key":       "foo",
								"doc_count": 2,
								"queue_size": map[string]any{
									"value": json.Number("10"),
								},
							},
							map[string]any{
								"key":       "bar",
								"doc_count": 3,
								"queue_size": map[string]any{
									"value": json.Number("20"),
								},
							},
						},
					},
				},
			},
			conditions: []config.Condition{
				{
					"field":      "aggregations.hostname.buckets.queue_size.value",
					"quantifier": "any",
					"lt":         json.Number("9"),
				},
			},
			output: nil,
			hits:   0,
			err:    false,
		},
		{
			name: "conditions-met",
			input: map[string]any{
				"aggregations": map[string]any{
					"hostname": map[string]any{
						"buckets": []any{
							map[string]any{
								"key":       "foo",
								"doc_count": 2,
								"queue_size": map[string]any{
									"value": json.Number("10"),
								},
							},
							map[string]any{
								"key":       "bar",
								"doc_count": 3,
								"queue_size": map[string]any{
									"value": json.Number("20"),
								},
							},
						},
					},
				},
			},
			filters: []string{"aggregations.hostname.buckets"},
			conditions: []config.Condition{
				{
					"field":      "aggregations.hostname.buckets.queue_size.value",
					"quantifier": "any",
					"lt":         json.Number("11"),
				},
			},
			output: []*alert.Record{
				{
					Filter: "aggregations.hostname.buckets",
					Fields: []*alert.Field{
						{
							Key:   "foo",
							Count: 2,
						},
						{
							Key:   "bar",
							Count: 3,
						},
					},
				},
			},
			hits: 0,
			err:  false,
		},
	}

	logger := hclog.NewNullLogger()

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			qh := &QueryHandler{
				logger:     logger,
				filters:    tc.filters,
				bodyField:  defaultBodyField,
				conditions: tc.conditions,
			}
			records, hits, err := qh.process(tc.input)
			if tc.hits != len(hits) {
				t.Fatalf("Got %d hits, expected %d", len(hits), tc.hits)
			}
			if !tc.err && err != nil {
				t.Fatal(err)
			}
			if tc.err && err == nil {
				t.Fatal("expected an error but did not receive one")
			}
			if !cmp.Equal(tc.output, records) {
				t.Errorf("Results differ:\n%v", cmp.Diff(tc.output, records))
			}
		})
	}
}
