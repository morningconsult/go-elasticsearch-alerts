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

package alert

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	hclog "github.com/hashicorp/go-hclog"
	uuid "github.com/hashicorp/go-uuid"
	"golang.org/x/xerrors"
)

// Ensure Method adheres to the Method interface.
var _ Method = (*fileAlertMethod)(nil)

type OutputJSON struct {
	RuleName   string    `json:"rule_name"`
	ReceivedAt time.Time `json:"received_at"`
	Records    []*Record `json:"results"`
}

// filealertMethod is defined here rather than importing
// gitlab.morningconsult.com/mci/go-elasticsearch-alerts/command/alert/file
// to avoid import cycle.
type fileAlertMethod struct {
	outputFilepath string
}

func (f *fileAlertMethod) Write(ctx context.Context, rule string, records []*Record) error {
	outfile, err := os.OpenFile(f.outputFilepath, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0o600)
	if err != nil {
		return xerrors.Errorf("error opening new file: %v", err)
	}
	defer outfile.Close()

	entry := OutputJSON{
		RuleName:   rule,
		ReceivedAt: time.Now(),
		Records:    records,
	}
	return json.NewEncoder(outfile).Encode(&entry)
}

// errorAlertMethod is a mock alert.AlertMethod used to simulate an
// even where AlertMethod.Write() return an error.
type errorAlertMethod struct{}

func (e *errorAlertMethod) Write(ctx context.Context, rule string, records []*Record) error {
	return xerrors.Errorf("test error")
}

func TestRun(t *testing.T) {
	outputCh := make(chan *Alert, 1)

	ctx, cancel := context.WithTimeout(t.Context(), 10*time.Second)

	ah := NewHandler(&HandlerConfig{
		Logger: hclog.NewNullLogger(),
	})

	filename := filepath.Join("testdata", "testdata.log")
	defer os.Remove(filename)

	fm := &fileAlertMethod{
		outputFilepath: filename,
	}

	a := &Alert{
		ID:       randomUUID(t),
		RuleName: "test-rule",
		Methods:  []Method{fm},
		Records: []*Record{
			{
				Filter: "test.rule.1",
				Text:   "test text",
				Fields: []*Field{
					{
						Key:   "hello",
						Count: 10,
					},
					{
						Key:   "world",
						Count: 3,
					},
				},
			},
			{
				Filter: "test.rule.2",
				Text:   "test text",
			},
		},
	}

	outputCh <- a

	go ah.Run(ctx, outputCh)

	defer func() {
		cancel()
		<-ah.DoneCh
	}()

	select {
	case <-ctx.Done():
		t.Fatal("context timed out")
	case <-time.After(500 * time.Millisecond):
		// check for file
		logfile, err := os.Open(filepath.Clean(filename))
		if err != nil {
			t.Fatal(err)
		}
		defer logfile.Close()

		data := OutputJSON{}
		if err = json.NewDecoder(logfile).Decode(&data); err != nil {
			t.Fatal(err)
		}

		if data.RuleName != a.RuleName {
			t.Fatalf("rule name mismatch (got %q, expected %q)", data.RuleName, a.RuleName)
		}

		if len(data.Records) != len(a.Records) {
			t.Fatalf("received unexpected number of records (got %d, expected %d)", len(a.Records), len(data.Records))
		}
	}
}

func TestRunError(t *testing.T) {
	outputCh := make(chan *Alert, 1)

	ctx, cancel := context.WithTimeout(t.Context(), 10*time.Second)

	buf := new(bytes.Buffer)
	logger := hclog.New(&hclog.LoggerOptions{
		Output: buf,
	})
	ah := NewHandler(&HandlerConfig{
		Logger: logger,
	})

	em := &errorAlertMethod{}

	a := &Alert{
		ID:       randomUUID(t),
		RuleName: "test-rule",
		Methods:  []Method{em},
		Records: []*Record{
			{
				Filter: "test.rule.1",
				Text:   "test text",
				Fields: []*Field{
					{
						Key:   "hello",
						Count: 10,
					},
					{
						Key:   "world",
						Count: 3,
					},
				},
			},
			{
				Filter: "test.rule.2",
				Text:   "test text",
			},
		},
	}

	outputCh <- a

	go ah.Run(ctx, outputCh)

	defer func() {
		cancel()
		<-ah.DoneCh
	}()

	time.Sleep(7 * time.Second)

	// Should attempt to execute Write() 3 times (see logs)
	expected := `[ERROR] error returned by alert function: error="test error" remaining_retries=0`
	if !strings.Contains(buf.String(), expected) {
		t.Fatalf("Expected errors to contain:\n\t%s\nGot:\n\t%s", expected, buf.String())
	}
}

func randomUUID(t *testing.T) string {
	id, err := uuid.GenerateUUID()
	if err != nil {
		t.Fatal(err)
	}
	return id
}
