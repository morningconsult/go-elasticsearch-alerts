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

package file

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/hashicorp/vault/helper/jsonutil"
	"github.com/mitchellh/go-homedir"
	"github.com/morningconsult/go-elasticsearch-alerts/command/alert"
)

// Ensure FileAlertMethod adheres to the alert.AlertMethod interface
var _ alert.AlertMethod = (*FileAlertMethod)(nil)

type OutputJSON struct {
	RuleName   string          `json:"rule_name"`
	ReceivedAt time.Time       `json:"received_at"`
	Records    []*alert.Record `json:"results"`
}

type FileAlertMethodConfig struct {
	OutputFilepath string `mapstructure:"file"`
}

type FileAlertMethod struct {
	outputFilepath string
}

func NewFileAlertMethod(config *FileAlertMethodConfig) (*FileAlertMethod, error) {
	if config.OutputFilepath == "" {
		return nil, errors.New("no file path provided")
	}

	expanded, err := homedir.Expand(config.OutputFilepath)
	if err != nil {
		return nil, fmt.Errorf("error expanding file path %q: %v", config.OutputFilepath, err)
	}

	return &FileAlertMethod{
		outputFilepath: expanded,
	}, nil
}

func (f *FileAlertMethod) Write(ctx context.Context, rule string, records []*alert.Record) error {
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
