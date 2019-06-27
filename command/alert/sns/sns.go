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
	"fmt"
	"text/template"

	"github.com/Masterminds/sprig"
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
	Template string `mapstructure:"template"`
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
	if config.Template == "" {
		return nil, xerrors.New("field 'output.config.template' must not be empty when using the SNS output method")
	}
	tmpl, err := template.New("sns").Funcs(template.FuncMap(sprig.FuncMap())).Parse(config.Template)
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
	msg, err := a.renderTemplate(rule, records)
	if err != nil {
		return err
	}
	input := &sns.PublishInput{
		Message:  aws.String(msg),
		TopicArn: aws.String(a.topicARN),
	}
	_, err = a.client.PublishWithContext(ctx, input)
	if err != nil {
		return xerrors.Errorf("error publishing alert to SNS: %w", err)
	}
	return nil
}

func (a *AlertMethod) renderTemplate(rule string, records []*alert.Record) (string, error) {
	out := bytes.Buffer{}
	if err := a.template.Execute(&out, records); err != nil {
		return "", xerrors.Errorf("error executing SNS message template: %w", err)
	}
	if out.String() == "" {
		out.WriteString("New alerts detected. See logs.")
	}
	return fmt.Sprintf("[%s]\n%s", rule, out.String()), nil
}
