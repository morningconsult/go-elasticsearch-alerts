package file

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/mitchellh/go-homedir"
	"github.com/hashicorp/vault/helper/jsonutil"
	"gitlab.morningconsult.com/mci/go-elasticsearch-alerts/command/alert"
)

// Ensure FileAlertMethod adheres to the alert.AlertMethod interface
var _ alert.AlertMethod = (*FileAlertMethod)(nil)

type OutputJSON struct {
	RuleName   string          `json:"rule_name"`
	ReceivedAt time.Time       `json:"received_at"`
	Records    []*alert.Record `json:"results"`
}

type FileAlertMethodConfig struct {
	RuleName       string
	OutputFilepath string `mapstructure:"file"`
}

type FileAlertMethod struct {
	ruleName       string
	outputFilepath string
}

func NewFileAlertMethod(config *FileAlertMethodConfig) (*FileAlertMethod, error) {
	if config.RuleName == "" {
		return nil, errors.New("no rule name provided")
	}

	if config.OutputFilepath == "" {
		return nil, errors.New("no file path provided")
	}

	expanded, err := homedir.Expand(config.OutputFilepath)
	if err != nil {
		return nil, fmt.Errorf("error expanding file path %q: %v", config.OutputFilepath, err)
	}

	return &FileAlertMethod{
		ruleName:       config.RuleName,
		outputFilepath: expanded,
	}, nil
}

func (f *FileAlertMethod) Write(ctx context.Context, records []*alert.Record) error {
	outfile, err := os.OpenFile(f.outputFilepath, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		return fmt.Errorf("error opening new file: %v", err)
	}
	defer outfile.Close()

	entry := &OutputJSON{
		RuleName:   f.ruleName,
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