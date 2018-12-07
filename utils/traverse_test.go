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

package utils

import (
	"fmt"
	"reflect"
	"testing"
)

func TestGet(t *testing.T) {
	cases := []struct {
		name   string
		json   map[string]interface{}
		path   string
		output interface{}
	}{
		{
			"map",
			map[string]interface{}{
				"hello": map[string]interface{}{
					"darkness": map[string]interface{}{
						"my": map[string]interface{}{
							"old": "friend",
						},
					},
				},
			},
			"hello.darkness.my",
			map[string]interface{}{
				"old": "friend",
			},
		},
		{
			"array",
			map[string]interface{}{
				"hello": map[string]interface{}{
					"darkness": map[string]interface{}{
						"my": []interface{}{
							map[string]interface{}{
								"old": "friend",
							},
							map[string]interface{}{
								"ive": "come",
							},
							map[string]interface{}{
								"to": "talk",
							},
						},
					},
				},
			},
			"hello.darkness.my",
			[]interface{}{
				map[string]interface{}{
					"old": "friend",
				},
				map[string]interface{}{
					"ive": "come",
				},
				map[string]interface{}{
					"to": "talk",
				},
			},
		},
		{
			"within-array",
			map[string]interface{}{
				"hello": map[string]interface{}{
					"darkness": map[string]interface{}{
						"my": []interface{}{
							map[string]interface{}{
								"old": "friend",
							},
							map[string]interface{}{
								"ive": "come",
							},
							map[string]interface{}{
								"to": "talk",
							},
						},
					},
				},
			},
			"hello.darkness.my[2].to",
			"talk",
		},
		{
			"non-int-index",
			map[string]interface{}{
				"hello": map[string]interface{}{
					"darkness": map[string]interface{}{
						"my": []interface{}{
							map[string]interface{}{
								"old": "friend",
							},
							map[string]interface{}{
								"ive": "come",
							},
							map[string]interface{}{
								"to": "talk",
							},
						},
					},
				},
			},
			"hello.darkness.my[a].to",
			nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			out := Get(tc.json, tc.path)
			if !reflect.DeepEqual(out, tc.output) {
				t.Fatalf("Got:\n%+v\n\nExpected:\n%+v", out, tc.output)
			}
		})
	}
}

func TestGetAll(t *testing.T) {
	cases := []struct {
		name   string
		json   map[string]interface{}
		path   string
		output interface{}
	}{
		{
			"not-nested",
			map[string]interface{}{
				"hello": map[string]interface{}{
					"darkness": map[string]interface{}{
						"my": []interface{}{
							map[string]interface{}{
								"old": "friend",
							},
							map[string]interface{}{
								"ive": "come",
							},
							map[string]interface{}{
								"to": "talk",
							},
						},
					},
				},
			},
			"hello.darkness.my",
			[]interface{}{
				map[string]interface{}{
					"old": "friend",
				},
				map[string]interface{}{
					"ive": "come",
				},
				map[string]interface{}{
					"to": "talk",
				},
			},
		},
		{
			"nested",
			map[string]interface{}{
				"hello": map[string]interface{}{
					"darkness": map[string]interface{}{
						"buckets": []interface{}{
							map[string]interface{}{
								"key": "old",
								"ayy": map[string]interface{}{
									"buckets": []interface{}{
										map[string]interface{}{
											"key":   "greg",
											"hello": "world",
										},
										map[string]interface{}{
											"key":   "friend",
											"hello": "darkness",
										},
									},
								},
							},
							map[string]interface{}{
								"key": "yesterday",
								"ayy": map[string]interface{}{
									"buckets": []interface{}{
										map[string]interface{}{
											"key": "troubles",
											"far": "away",
										},
										map[string]interface{}{
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
			[]interface{}{
				map[string]interface{}{
					"key":   "old - greg",
					"hello": "world",
				},
				map[string]interface{}{
					"key":   "old - friend",
					"hello": "darkness",
				},
				map[string]interface{}{
					"key": "yesterday - troubles",
					"far": "away",
				},
				map[string]interface{}{
					"key": "yesterday - here",
					"to":  "stay",
				},
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			out := GetAll(tc.json, tc.path)
			if !reflect.DeepEqual(out, tc.output) {
				t.Fatalf("Got:\n%+v\n\nExpected:\n%+v", out, tc.output)
			}
		})
	}
}

func ExampleGet() {
	jsonData := map[string]interface{}{
		"hello": map[string]interface{}{
			"world": []interface{}{
				map[string]interface{}{
					"foo": "example-1",
					"bar": map[string]interface{}{
						"bim": "baz",
					},
				},
				map[string]interface{}{
					"foo": "example-2",
					"bar": map[string]interface{}{
						"ping": "pong",
					},
				},
			},
		},
	}

	v := Get(jsonData, "hello.world.bar")
	fmt.Printf("%v\n", v)

	v = Get(jsonData, "hello.world[0].foo")
	fmt.Printf("%v\n", v)

	v = Get(jsonData, "hello.world[0].bar")
	fmt.Printf("%#v\n", v)

	// Output:
	// <nil>
	// example-1
	// map[string]interface {}{"bim":"baz"}
}

func ExampleGetAll() {
	jsonData := map[string]interface{}{
		"hits": map[string]interface{}{
			"hits": []interface{}{
				map[string]interface{}{
					"_source": map[string]interface{}{
						"foo": "bar",
					},
				},
				map[string]interface{}{
					"_source": map[string]interface{}{
						"bim": "baz",
					},
				},
			},
		},
		"aggregations": map[string]interface{}{
			"hostname": map[string]interface{}{
				"buckets": []interface{}{
					map[string]interface{}{
						"key":   "foo",
						"count": 10,
						"program": map[string]interface{}{
							"buckets": []interface{}{
								map[string]interface{}{
									"key":   "bar",
									"count": 3,
								},
								map[string]interface{}{
									"key":   "bim",
									"count": 7,
								},
							},
						},
					},
					map[string]interface{}{
						"key":   "hello",
						"count": 6,
						"program": map[string]interface{}{
							"buckets": []interface{}{
								map[string]interface{}{
									"key":   "world",
									"count": 2,
								},
								map[string]interface{}{
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

	expected1 := []interface{}{
		map[string]interface{}{
			"key":   "foo",
			"count": 10,
			"program": map[string]interface{}{
				"buckets": []interface{}{
					map[string]interface{}{
						"key":   "bar",
						"count": 3,
					},
					map[string]interface{}{
						"key":   "bim",
						"count": 7,
					},
				},
			},
		},
		map[string]interface{}{
			"key":   "hello",
			"count": 6,
			"program": map[string]interface{}{
				"buckets": []interface{}{
					map[string]interface{}{
						"key":   "world",
						"count": 2,
					},
					map[string]interface{}{
						"key":   "darkness",
						"count": 4,
					},
				},
			},
		},
	}

	expected2 := []interface{}{
		map[string]interface{}{
			"key":   "foo - bar",
			"count": 3,
		},
		map[string]interface{}{
			"key":   "foo - bim",
			"count": 7,
		},
		map[string]interface{}{
			"key":   "hello - world",
			"count": 2,
		},
		map[string]interface{}{
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
