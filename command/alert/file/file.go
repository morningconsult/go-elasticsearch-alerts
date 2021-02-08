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
	"os"
	"time"

	multierror "github.com/hashicorp/go-multierror"
	homedir "github.com/mitchellh/go-homedir"
	"github.com/morningconsult/go-elasticsearch-alerts/command/alert"
	"golang.org/x/xerrors"
)

// Ensure AlertMethod adheres to the alert.Method interface.
var _ alert.Method = (*AlertMethod)(nil)

type outputJSON struct {
	RuleName   string          `json:"rule_name"`
	ReceivedAt time.Time       `json:"received_at"`
	Records    []*alert.Record `json:"results"`
}

// AlertMethodConfig configures to what file alerts will be written.
type AlertMethodConfig struct {
	// OutputFilepath is the file where logs will be written
	OutputFilepath string `mapstructure:"file"`
}

// AlertMethod implements the alert.AlertMethod interface
// for writing new alerts to a file.
type AlertMethod struct {
	outputFilepath string
}

// NewAlertMethod returns a new *AlertMethod or a non-nil
// error if there was an error.
func NewAlertMethod(config *AlertMethodConfig) (alert.Method, error) {
	if err := validateConfig(config); err != nil {
		return nil, err
	}

	expanded, err := homedir.Expand(config.OutputFilepath)
	if err != nil {
		return nil, xerrors.Errorf("error expanding file path %q: %v", config.OutputFilepath, err)
	}

	return &AlertMethod{
		outputFilepath: expanded,
	}, nil
}

func validateConfig(config *AlertMethodConfig) error {
	var allErrors *multierror.Error
	if config == nil {
		allErrors = multierror.Append(xerrors.New("no config provided"))
	} else if config.OutputFilepath == "" {
		allErrors = multierror.Append(xerrors.New("no file path provided"))
	}
	return allErrors.ErrorOrNil()
}

// Write creates JSON-formatted logs from the records and writes
// them to the file specified at the creation of the AlertMethod.
// If there was an error writing logs to disk, it returns a
// non-nil error.
func (f *AlertMethod) Write(ctx context.Context, rule string, records []*alert.Record) error {
	outfile, err := os.OpenFile(f.outputFilepath, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0600)
	if err != nil {
		return xerrors.Errorf("error opening new file: %v", err)
	}
	defer outfile.Close()

	entry := outputJSON{
		RuleName:   rule,
		ReceivedAt: time.Now(),
		Records:    records,
	}
	return json.NewEncoder(outfile).Encode(&entry)
}
