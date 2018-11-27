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
	"reflect"
	"testing"
)

func TestGet(t *testing.T) {
	cases := []struct{
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
	cases := []struct{
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
