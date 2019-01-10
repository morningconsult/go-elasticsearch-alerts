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
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/go-uuid"
	"github.com/hashicorp/vault/helper/jsonutil"
)

// Ensure FileAlertMethod adheres to the AlertMethod interface
var _ AlertMethod = (*fileAlertMethod)(nil)

type OutputJSON struct {
	RuleName   string    `json:"rule_name"`
	ReceivedAt time.Time `json:"received_at"`
	Records    []*Record `json:"results"`
}

// dilealertMethod is defined here rather than importing
// gitlab.morningconsult.com/mci/go-elasticsearch-alerts/command/alert/file
// to avoid import cycle
type fileAlertMethod struct {
	outputFilepath string
}

func (f *fileAlertMethod) Write(ctx context.Context, rule string, records []*Record) error {
	outfile, err := os.OpenFile(f.outputFilepath, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		return fmt.Errorf("error opening new file: %v", err)
	}
	defer outfile.Close()

	entry := &OutputJSON{
		RuleName:   rule,
		ReceivedAt: time.Now(),
		Records:    records,
	}
	data, err := jsonutil.EncodeJSON(entry)
	if err != nil {
		return fmt.Errorf("error JSON-encoding data: %v", err)
	}

	return write(outfile, data)
}

// errorAlertMethod is a mock alert.AlertMethod used to simulate an
// even where AlertMethod.Write() return an error
type errorAlertMethod struct{}

func (e *errorAlertMethod) Write(ctx context.Context, rule string, records []*Record) error {
	return fmt.Errorf("test error")
}

func write(writer io.Writer, data []byte) error {
	start := 0
	for {
		if start >= len(data) {
			break
		}

		n, err := writer.Write(data[start:])
		if err != nil {
			return fmt.Errorf("error writing data: %v", err)
		}

		start += n
	}
	return nil
}

func TestRun(t *testing.T) {
	var wg sync.WaitGroup
	outputCh := make(chan *Alert, 1)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)

	ah := NewAlertHandler(&AlertHandlerConfig{
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
		Methods:  []AlertMethod{fm},
		Records: []*Record{
			&Record{
				Filter: "test.rule.1",
				Text:   "test text",
				Fields: []*Field{
					&Field{
						Key:   "hello",
						Count: 10,
					},
					&Field{
						Key:   "world",
						Count: 3,
					},
				},
			},
			&Record{
				Filter: "test.rule.2",
				Text:   "test text",
			},
		},
	}

	outputCh <- a

	wg.Add(1)

	go ah.Run(ctx, outputCh)

	defer func() {
		cancel()
		<-ah.DoneCh
	}()
	for {
		select {
		case <-ctx.Done():
			t.Fatal("context timed out")
		case <-time.After(500 * time.Millisecond):
			// check for file
			logfile, err := os.Open(filename)
			if err != nil {
				t.Fatal(err)
			}

			json := new(OutputJSON)
			if err = jsonutil.DecodeJSONFromReader(logfile, json); err != nil {
				t.Fatal(err)
			}

			if json.RuleName != a.RuleName {
				t.Fatalf("rule name mismatch (got %q, expected %q)", json.RuleName, a.RuleName)
			}

			if len(json.Records) != len(a.Records) {
				t.Fatalf("received unexpected number of records (got %d, expected %d)", len(a.Records), len(json.Records))
			}
			return
		}
	}
}

func TestRunError(t *testing.T) {
	var wg sync.WaitGroup
	outputCh := make(chan *Alert, 1)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)

	buf := new(bytes.Buffer)
	logger := hclog.New(&hclog.LoggerOptions{
		Output: buf,
	})
	ah := NewAlertHandler(&AlertHandlerConfig{
		Logger: logger,
	})

	em := &errorAlertMethod{}

	a := &Alert{
		ID:       randomUUID(t),
		RuleName: "test-rule",
		Methods:  []AlertMethod{em},
		Records: []*Record{
			&Record{
				Filter: "test.rule.1",
				Text:   "test text",
				Fields: []*Field{
					&Field{
						Key:   "hello",
						Count: 10,
					},
					&Field{
						Key:   "world",
						Count: 3,
					},
				},
			},
			&Record{
				Filter: "test.rule.2",
				Text:   "test text",
			},
		},
	}

	outputCh <- a

	wg.Add(1)

	go ah.Run(ctx, outputCh)

	defer func() {
		cancel()
		<-ah.DoneCh
	}()

	time.Sleep(10 * time.Second)

	// Should attempt to execute Write() 3 times (see logs)
	expected := `[ERROR] [Alert Handler] error returned by alert function: error="test error" remaining_retries=0`
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
