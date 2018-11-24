package file

import (
	"context"
	"path/filepath"
	"os"
	"reflect"
	"testing"

	"github.com/hashicorp/vault/helper/jsonutil"
	"gitlab.morningconsult.com/mci/go-elasticsearch-alerts/command/alert"
)

func TestNewFileAlertMethod(t *testing.T) {
	cases := []struct{
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
	cases := []struct{
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
					Title: "hits.hits._source",
					Text:  "{\n    \"ayy\": \"lmao\"\n}",
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
