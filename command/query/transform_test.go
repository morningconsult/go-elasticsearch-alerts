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
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/morningconsult/go-elasticsearch-alerts/command/alert"
)

func TestTransform(t *testing.T) {
	cases := []struct {
		name    string
		input   map[string]interface{}
		filters []string
		output  []*alert.Record
		hits    int
		err     bool
	}{
		{
			"one-level",
			map[string]interface{}{
				"aggregations": map[string]interface{}{
					"hostname": map[string]interface{}{
						"buckets": []interface{}{
							map[string]interface{}{
								"key":       "foo",
								"doc_count": 2,
							},
							map[string]interface{}{
								"key":       "bar",
								"doc_count": 3,
							},
						},
					},
				},
			},
			[]string{"aggregations.hostname.buckets"},
			[]*alert.Record{
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
			0,
			false,
		},
		{
			"field-not-map",
			map[string]interface{}{
				"aggregations": map[string]interface{}{
					"hostname": map[string]interface{}{
						"buckets": []interface{}{
							"string",
							map[string]interface{}{
								"key":       "bar",
								"doc_count": 3,
							},
						},
					},
				},
			},
			[]string{"aggregations.hostname.buckets"},
			[]*alert.Record{
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
			0,
			false,
		},
		{
			"zero-count",
			map[string]interface{}{
				"aggregations": map[string]interface{}{
					"hostname": map[string]interface{}{
						"buckets": []interface{}{
							map[string]interface{}{
								"key":       "foo",
								"doc_count": 0,
							},
							map[string]interface{}{
								"key":       "bar",
								"doc_count": 3,
							},
						},
					},
				},
			},
			[]string{"aggregations.hostname.buckets"},
			[]*alert.Record{
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
			0,
			false,
		},
		{
			"two-levels",
			map[string]interface{}{
				"aggregations": map[string]interface{}{
					"hostname": map[string]interface{}{
						"buckets": []interface{}{
							map[string]interface{}{
								"key":       "foo",
								"doc_count": 5,
								"program": map[string]interface{}{
									"buckets": []interface{}{
										map[string]interface{}{
											"key":       "bim",
											"doc_count": 2,
										},
										map[string]interface{}{
											"key":       "baz",
											"doc_count": 3,
										},
									},
								},
							},
							map[string]interface{}{
								"key":       "bar",
								"doc_count": 3,
								"program": map[string]interface{}{
									"buckets": []interface{}{
										map[string]interface{}{
											"key":       "ayy",
											"doc_count": 1,
										},
										map[string]interface{}{
											"key":       "lmao",
											"doc_count": 2,
										},
									},
								},
							},
						},
					},
				},
			},
			[]string{"aggregations.hostname.buckets.program.buckets"},
			[]*alert.Record{
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
			0,
			false,
		},
		{
			"hits-not-array",
			map[string]interface{}{
				"aggregations": map[string]interface{}{
					"hostname": map[string]interface{}{
						"buckets": []interface{}{
							map[string]interface{}{
								"key":       "foo",
								"doc_count": 2,
							},
							map[string]interface{}{
								"key":       "bar",
								"doc_count": 3,
							},
						},
					},
				},
				"hits": map[string]interface{}{
					"hits": map[string]interface{}{
						"ayy": "lmao",
					},
				},
			},
			[]string{"aggregations.hostname.buckets"},
			[]*alert.Record{
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
			0,
			false,
		},
		{
			"hit-elems-not-maps",
			map[string]interface{}{
				"aggregations": map[string]interface{}{
					"hostname": map[string]interface{}{
						"buckets": []interface{}{
							map[string]interface{}{
								"key":       "foo",
								"doc_count": 2,
							},
							map[string]interface{}{
								"key":       "bar",
								"doc_count": 3,
							},
						},
					},
				},
				"hits": map[string]interface{}{
					"hits": []interface{}{
						"sadly",
						"i",
						"am",
						"only",
						"a",
						"string",
					},
				},
			},
			[]string{"aggregations.hostname.buckets"},
			[]*alert.Record{
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
			0,
			false,
		},
		{
			"hit-elems-have-no-source",
			map[string]interface{}{
				"aggregations": map[string]interface{}{
					"hostname": map[string]interface{}{
						"buckets": []interface{}{
							map[string]interface{}{
								"key":       "foo",
								"doc_count": 2,
							},
							map[string]interface{}{
								"key":       "bar",
								"doc_count": 3,
							},
						},
					},
				},
				"hits": map[string]interface{}{
					"hits": []interface{}{
						map[string]interface{}{
							"any": "field",
							"but": "_source!",
						},
						map[string]interface{}{
							"_source": map[string]interface{}{
								"ayy": "lmao",
							},
						},
					},
				},
			},
			[]string{"aggregations.hostname.buckets"},
			[]*alert.Record{
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
			1,
			false,
		},
		{
			"hits-only",
			map[string]interface{}{
				"hits": map[string]interface{}{
					"hits": []interface{}{
						map[string]interface{}{
							"_source": map[string]interface{}{
								"ayy": "lmao",
							},
						},
						map[string]interface{}{
							"_source": map[string]interface{}{
								"yeah": "buddy",
							},
						},
					},
				},
			},
			[]string{},
			[]*alert.Record{
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
			2,
			false,
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			qh := &QueryHandler{
				filters:   tc.filters,
				bodyField: defaultBodyField,
			}
			records, hits, err := qh.Transform(tc.input)
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
