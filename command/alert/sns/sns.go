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

package sns

import (
	"bytes"
	"context"
	"text/template"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sns"
	"github.com/morningconsult/go-elasticsearch-alerts/command/alert"
	"golang.org/x/xerrors"
)

// AlertMethodConfig configures where AWS SNS alerts will be
// published and what the published messages should look like.
type AlertMethodConfig struct {
	Region   string `mapstructure:"region"`
	TopicARN string `mapstructure:"topic_arn"`
	Message  string `mapstructure:"message"`
}

// AlertMethod implements the alert.AlertMethod interface
// for publishing new alerts to an AWS SNS topic.
type AlertMethod struct {
	client   *sns.SNS
	topicARN string
	template *template.Template
}

// NewAlertMethod creates a new *AlertMethod or a
// non-nil error if there was an error.
func NewAlertMethod(config *AlertMethodConfig) (alert.Method, error) {
	if config == nil {
		return nil, xerrors.New("no config provided")
	}
	if config.Region == "" {
		return nil, xerrors.New("field 'output.config.region' must not be empty when using the SNS output method")
	}
	if config.TopicARN == "" {
		return nil, xerrors.New("field 'output.config.topic_arn' must not be empty when using the SNS output method")
	}
	if config.Message == "" {
		return nil, xerrors.New("field 'output.config.message' must not be empty when using the SNS output method")
	}
	tmpl, err := template.New("sns").Parse(config.Message)
	if err != nil {
		return nil, xerrors.Errorf("error parsing SNS message template: %w", err)
	}
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(config.Region),
	})
	if err != nil {
		return nil, xerrors.Errorf("error creating new SNS alert method: %w", err)
	}
	return &AlertMethod{
		client:   sns.New(sess),
		topicARN: config.TopicARN,
		template: tmpl,
	}, nil
}

// Write renders the pre-defined message template and publishes
// the message to an AWS SNS topic.
func (a *AlertMethod) Write(ctx context.Context, rule string, records []*alert.Record) error {
	if records == nil || len(records) < 1 {
		return nil
	}
	out := bytes.Buffer{}
	if err := a.template.Execute(&out, records); err != nil {
		return xerrors.Errorf("error executing SNS message template: %w", err)
	}
	input := &sns.PublishInput{
		Message:  aws.String(out.String()),
		TopicArn: aws.String(a.topicARN),
	}
	_, err := a.client.PublishWithContext(ctx, input)
	return err
}
