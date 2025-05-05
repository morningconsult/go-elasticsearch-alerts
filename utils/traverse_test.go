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

package utils

import (
	"fmt"
	"reflect"
	"testing"
)

func TestGetAll(t *testing.T) {
	cases := []struct {
		name   string
		json   map[string]any
		path   string
		output any
	}{
		{
			"not-nested",
			map[string]any{
				"hello": map[string]any{
					"darkness": map[string]any{
						"my": []any{
							map[string]any{
								"old": "friend",
							},
							map[string]any{
								"ive": "come",
							},
							map[string]any{
								"to": "talk",
							},
						},
					},
				},
			},
			"hello.darkness.my",
			[]any{
				map[string]any{
					"old": "friend",
				},
				map[string]any{
					"ive": "come",
				},
				map[string]any{
					"to": "talk",
				},
			},
		},
		{
			"nested",
			map[string]any{
				"hello": map[string]any{
					"darkness": map[string]any{
						"buckets": []any{
							map[string]any{
								"key": "old",
								"ayy": map[string]any{
									"buckets": []any{
										map[string]any{
											"key":   "greg",
											"hello": "world",
										},
										map[string]any{
											"key":   "friend",
											"hello": "darkness",
										},
									},
								},
							},
							map[string]any{
								"key": "yesterday",
								"ayy": map[string]any{
									"buckets": []any{
										map[string]any{
											"key": "troubles",
											"far": "away",
										},
										map[string]any{
											"key": "here",
											"to":  "stay",
										},
									},
								},
							},
						},
					},
				},
			},
			"hello.darkness.buckets.ayy.buckets",
			[]any{
				map[string]any{
					"key":   "old - greg",
					"hello": "world",
				},
				map[string]any{
					"key":   "old - friend",
					"hello": "darkness",
				},
				map[string]any{
					"key": "yesterday - troubles",
					"far": "away",
				},
				map[string]any{
					"key": "yesterday - here",
					"to":  "stay",
				},
			},
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			out := GetAll(tc.json, tc.path)
			if !reflect.DeepEqual(out, tc.output) {
				t.Fatalf("Got:\n%+v\n\nExpected:\n%+v", out, tc.output)
			}
		})
	}
}

func ExampleGetAll() {
	jsonData := map[string]any{
		"hits": map[string]any{
			"hits": []any{
				map[string]any{
					"_source": map[string]any{
						"foo": "bar",
					},
				},
				map[string]any{
					"_source": map[string]any{
						"bim": "baz",
					},
				},
			},
		},
		"aggregations": map[string]any{
			"hostname": map[string]any{
				"buckets": []any{
					map[string]any{
						"key":   "foo",
						"count": 10,
						"program": map[string]any{
							"buckets": []any{
								map[string]any{
									"key":   "bar",
									"count": 3,
								},
								map[string]any{
									"key":   "bim",
									"count": 7,
								},
							},
						},
					},
					map[string]any{
						"key":   "hello",
						"count": 6,
						"program": map[string]any{
							"buckets": []any{
								map[string]any{
									"key":   "world",
									"count": 2,
								},
								map[string]any{
									"key":   "darkness",
									"count": 4,
								},
							},
						},
					},
				},
			},
		},
	}

	expected1 := []any{
		map[string]any{
			"key":   "foo",
			"count": 10,
			"program": map[string]any{
				"buckets": []any{
					map[string]any{
						"key":   "bar",
						"count": 3,
					},
					map[string]any{
						"key":   "bim",
						"count": 7,
					},
				},
			},
		},
		map[string]any{
			"key":   "hello",
			"count": 6,
			"program": map[string]any{
				"buckets": []any{
					map[string]any{
						"key":   "world",
						"count": 2,
					},
					map[string]any{
						"key":   "darkness",
						"count": 4,
					},
				},
			},
		},
	}

	expected2 := []any{
		map[string]any{
			"key":   "foo - bar",
			"count": 3,
		},
		map[string]any{
			"key":   "foo - bim",
			"count": 7,
		},
		map[string]any{
			"key":   "hello - world",
			"count": 2,
		},
		map[string]any{
			"key":   "hello - darkness",
			"count": 4,
		},
	}

	v := GetAll(jsonData, "aggregations.hostname.buckets")
	fmt.Println(reflect.DeepEqual(v, expected1))

	v = GetAll(jsonData, "aggregations.hostname.buckets.program.buckets")
	fmt.Println(reflect.DeepEqual(v, expected2))

	v = GetAll(jsonData, "aggregations.buckets")
	fmt.Printf("%v\n", v)

	v = GetAll(jsonData, "hits.hits._source")
	fmt.Printf("%#v\n", v)

	// Output:
	// true
	// true
	// [<nil>]
	// []interface {}{map[string]interface {}{"foo":"bar"}, map[string]interface {}{"bim":"baz"}}
}
