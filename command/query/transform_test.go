package query

import (
	"testing"

	"gitlab.morningconsult.com/mci/go-elasticsearch-alerts/command/alert"
)

func TestTransform(t *testing.T) {
	cases := []struct{
		name    string
		input   map[string]interface{}
		filters []string
		output  []*alert.Record
		err     bool
	}{
		{
			"one-level",
			map[string]interface{}{
				"aggregations": map[string]interface{}{
					"hostname": map[string]interface{}{
						"buckets": []interface{}{
							map[string]interface{}{
								"key":      "foo",
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
				&alert.Record{
					Title:  "aggregations.hostname.buckets",
					Fields: []*alert.Field{
						&alert.Field{
							Key:   "foo",
							Count: 2,
						},
						&alert.Field{
							Key:   "bar",
							Count: 3,
						},
					},
				},
			},
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
				&alert.Record{
					Title:  "aggregations.hostname.buckets",
					Fields: []*alert.Field{
						&alert.Field{
							Key:   "bar",
							Count: 3,
						},
					},
				},
			},
			false,
		},
		{
			"zero-count",
			map[string]interface{}{
				"aggregations": map[string]interface{}{
					"hostname": map[string]interface{}{
						"buckets": []interface{}{
							map[string]interface{}{
								"key":      "foo",
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
				&alert.Record{
					Title:  "aggregations.hostname.buckets",
					Fields: []*alert.Field{
						&alert.Field{
							Key:   "bar",
							Count: 3,
						},
					},
				},
			},
			false,
		},
		{
			"two-levels",
			map[string]interface{}{
				"aggregations": map[string]interface{}{
					"hostname": map[string]interface{}{
						"buckets": []interface{}{
							map[string]interface{}{
								"key":      "foo",
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
								"program":  map[string]interface{}{
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
				&alert.Record{
					Title:  "aggregations.hostname.buckets.program.buckets",
					Fields: []*alert.Field{
						&alert.Field{
							Key:   "foo - bim",
							Count: 2,
						},
						&alert.Field{
							Key:   "foo - baz",
							Count: 3,
						},
						&alert.Field{
							Key:   "bar - ayy",
							Count: 1,
						},
						&alert.Field{
							Key:   "bar - lmao",
							Count: 2,
						},
					},
				},
			},
			false,
		},
		{
			"hits-not-array",
			map[string]interface{}{
				"aggregations": map[string]interface{}{
					"hostname": map[string]interface{}{
						"buckets": []interface{}{
							map[string]interface{}{
								"key":      "foo",
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
				&alert.Record{
					Title:  "aggregations.hostname.buckets",
					Fields: []*alert.Field{
						&alert.Field{
							Key:   "foo",
							Count: 2,
						},
						&alert.Field{
							Key:   "bar",
							Count: 3,
						},
					},
				},
			},
			false,
		},
		{
			"hit-elems-not-maps",
			map[string]interface{}{
				"aggregations": map[string]interface{}{
					"hostname": map[string]interface{}{
						"buckets": []interface{}{
							map[string]interface{}{
								"key":      "foo",
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
				&alert.Record{
					Title:  "aggregations.hostname.buckets",
					Fields: []*alert.Field{
						&alert.Field{
							Key:   "foo",
							Count: 2,
						},
						&alert.Field{
							Key:   "bar",
							Count: 3,
						},
					},
				},
			},
			false,
		},
		{
			"hit-elems-have-no-source",
			map[string]interface{}{
				"aggregations": map[string]interface{}{
					"hostname": map[string]interface{}{
						"buckets": []interface{}{
							map[string]interface{}{
								"key":      "foo",
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
				&alert.Record{
					Title:  "aggregations.hostname.buckets",
					Fields: []*alert.Field{
						&alert.Field{
							Key:   "foo",
							Count: 2,
						},
						&alert.Field{
							Key:   "bar",
							Count: 3,
						},
					},
				},
				&alert.Record{
					Title: "hits.hits._source",
					Text: `{"ayy": "lmao"}`,
				},
			},
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
				&alert.Record{
					Title: "hits.hits._source",
					Text:  `{
    "ayy": "lmao"
}
----------------------------------------
{
    "yeah": "buddy"
}`,
				},
			},
			false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			qh := &QueryHandler{
				filters: tc.filters,
			}
			records, err := qh.transform(tc.input)
			if !tc.err && err != nil {
				t.Fatal(err)
			}
			if tc.err && err == nil {
				t.Fatal("expected an error but did not receive one")
			}
			for i, record := range tc.output {
				if len(records) < i + 1 {
					t.Fatal("received records do not match expected records")
				}
				if records[i].Title != record.Title {
					t.Fatalf("record %d has unexpected title (got %q, expected %q)", i,
						records[i].Title, record.Title)
				}
				for j, field := range record.Fields {
					if len(records[i].Fields) < j + 1 {
						t.Fatal("received records.Fields does not match expected fields")
					}
					if records[i].Fields[j].Key != field.Key {
						t.Fatalf("field %d of record %d has unexpected key (got %q, expected %q)", i, j, 
							records[i].Fields[j].Key, field.Key)
					}
					if records[i].Fields[j].Count != field.Count {
						t.Fatalf("field %d of record %d has unexpected key (got %q, expected %q)", i, j, 
							records[i].Fields[j].Count, field.Count)
					}
				}
			}
		})
	}
}