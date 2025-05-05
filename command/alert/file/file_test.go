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
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/morningconsult/go-elasticsearch-alerts/command/alert"
)

func TestNewAlertMethod(t *testing.T) {
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
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			a, err := NewAlertMethod(&AlertMethodConfig{
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
			f, ok := a.(*AlertMethod)
			if !ok {
				t.Fatalf("Expected type *AlertMethod")
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
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			defer os.Remove(tc.filename)
			f, err := NewAlertMethod(&AlertMethodConfig{
				OutputFilepath: tc.filename,
			})
			if err != nil {
				t.Fatal(err)
			}
			records := []*alert.Record{
				{
					Filter: "hits.hits._source",
					Text:   "{\n    \"ayy\": \"lmao\"\n}",
				},
			}
			ctx := t.Context()
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

			data := outputJSON{}
			if err = json.NewDecoder(jsonfile).Decode(&data); err != nil {
				t.Fatal(err)
			}

			if !reflect.DeepEqual(data.Records, records) {
				t.Fatalf("Got:%+v\n\nExpected:\n%+v", data.Records, records)
			}
		})
	}
}

func ExampleAlertMethod_Write() {
	records := []*alert.Record{
		{
			Filter: "hits.hits._source",
			Text:   `Lorem ipsum dolor sit amet...`,
		},
		{
			Filter: "aggregation.hostname.buckets",
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
	}

	fm, err := NewAlertMethod(&AlertMethodConfig{
		OutputFilepath: "testdata/results.log",
	})
	if err != nil {
		fmt.Printf("error creating new *AlertMethod: %v", err)
		return
	}

	err = fm.Write(context.Background(), "Test Rule", records)
	if err != nil {
		fmt.Printf("error writing data to file: %v", err)
		return
	}
}
