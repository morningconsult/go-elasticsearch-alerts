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

package file

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/hashicorp/vault/helper/jsonutil"
	"github.com/morningconsult/go-elasticsearch-alerts/command/alert"
)

func TestNewFileAlertMethod(t *testing.T) {
	cases := []struct {
		name     string
		filename string
		err      bool
	}{
		{
			"success",
			"testdata/test.log",
			false,
		},
		{
			"no-file",
			"",
			true,
		},
		{
			"homedir-error",
			"~testdata",
			true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			f, err := NewFileAlertMethod(&FileAlertMethodConfig{
				OutputFilepath: tc.filename,
			})
			if tc.err {
				if err == nil {
					t.Fatal("expected an error but didn't receive one")
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if f.outputFilepath != tc.filename {
				t.Fatalf("unexpected filename (got %q, expected %q)", f.outputFilepath, tc.filename)
			}
		})
	}
}

func TestWrite(t *testing.T) {
	cases := []struct {
		name     string
		filename string
		err      bool
	}{
		{
			"success",
			filepath.Join("testdata", "test.log"),
			false,
		},
		{
			"open-error",
			"testdata/does/not/exist/test.log",
			true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			defer os.Remove(tc.filename)
			f, err := NewFileAlertMethod(&FileAlertMethodConfig{
				OutputFilepath: tc.filename,
			})
			if err != nil {
				t.Fatal(err)
			}
			records := []*alert.Record{
				&alert.Record{
					Filter: "hits.hits._source",
					Text:   "{\n    \"ayy\": \"lmao\"\n}",
				},
			}
			ctx := context.Background()
			err = f.Write(ctx, "test-rule", records)
			if tc.err {
				if err == nil {
					t.Fatal("expected an erorr but didn't receive one")
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			jsonfile, err := os.Open(tc.filename)
			if err != nil {
				t.Fatal(err)
			}
			defer jsonfile.Close()

			json := new(OutputJSON)
			if err = jsonutil.DecodeJSONFromReader(jsonfile, json); err != nil {
				t.Fatal(err)
			}

			if !reflect.DeepEqual(json.Records, records) {
				t.Fatalf("Got:%+v\n\nExpected:\n%+v", json.Records, records)
			}
		})
	}
}

func ExampleFileAlertMethod_Write() {
	records := []*alert.Record{
		&alert.Record{
			Filter: "hits.hits._source",
			Text:   `Lorem ipsum dolor sit amet...`,
		},
		&alert.Record{
			Filter: "aggregation.hostname.buckets",
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
	}

	fm, err := NewFileAlertMethod(&FileAlertMethodConfig{
		OutputFilepath: "testdata/results.log",
	})
	if err != nil {
		fmt.Printf("error creating new *FileAlertMethod: %v", err)
		return
	}

	err = fm.Write(context.Background(), "Test Rule", records)
	if err != nil {
		fmt.Printf("error writing data to file: %v", err)
		return
	}
}
