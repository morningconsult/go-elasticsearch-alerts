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

package slack

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	// "time"

	"github.com/hashicorp/go-cleanhttp"
	"github.com/hashicorp/vault/helper/jsonutil"
	"github.com/morningconsult/go-elasticsearch-alerts/command/alert"
)

const defaultTextLimit = 6000

// Ensure SlackAlertMethod adheres to the alert.AlertMethod interface
var _ alert.AlertMethod = (*SlackAlertMethod)(nil)

type SlackAlertMethodConfig struct {
	WebhookURL string `mapstructure:"webhook"`
	Channel    string `mapstructure:"channel"`
	Username   string `mapstructure:"username"`
	Text       string `mapstructure:"text"`
	Emoji      string `mapstructure:"emoji"`
	TextLimit  int    `mapstructure:"text_limit"`
	Client     *http.Client
}

type SlackAlertMethod struct {
	webhookURL string
	client     *http.Client
	channel    string
	username   string
	text       string
	emoji      string
	textLimit  int
}

type Payload struct {
	Channel     string        `json:"channel,omitempty"`
	Username    string        `json:"username,omitempty"`
	Text        string        `json:"text,omitempty"`
	Emoji       string        `json:"icon_emoji,omitempty"`
	Attachments []*Attachment `json:"attachments,omitempty"`
}

func NewSlackAlertMethod(config *SlackAlertMethodConfig) (*SlackAlertMethod, error) {
	if config == nil {
		config = &SlackAlertMethodConfig{}
	}

	if config.WebhookURL == "" {
		return nil, fmt.Errorf("field 'output.config.webhook' must not be empty when using the Slack output method")
	}

	if config.Client == nil {
		config.Client = cleanhttp.DefaultClient()
	}

	if config.TextLimit == 0 {
		config.TextLimit = defaultTextLimit
	}

	return &SlackAlertMethod{
		channel:    config.Channel,
		webhookURL: config.WebhookURL,
		client:     config.Client,
		text:       config.Text,
		emoji:      config.Emoji,
		textLimit:  config.TextLimit,
	}, nil
}

func (s *SlackAlertMethod) Write(ctx context.Context, rule string, records []*alert.Record) error {
	if records == nil || len(records) < 1 {
		return nil
	}
	return s.post(ctx, s.BuildPayload(rule, records))
}

func (s *SlackAlertMethod) BuildPayload(rule string, records []*alert.Record) *Payload {
	payload := &Payload{
		Channel:  s.channel,
		Username: s.username,
		Text:     s.text,
		Emoji:    s.emoji,
	}

	records = s.preprocess(records)

	for _, record := range records {
		config := &AttachmentConfig{
			Title:      rule,
			Text:       record.Title,
			MarkdownIn: []string{"text"},
		}
		if record.BodyField {
			config.Text = config.Text+"\n```\n"+record.Text+"\n```"
			config.Color = "#ff0000"
			config.MarkdownIn = []string{"text"}
		}

		att := NewAttachment(config)

		for _, field := range record.Fields {
			short := false
			if len(field.Key) <= 35 {
				short = true
			}
			f := &Field{
				Title: field.Key,
				Value: fmt.Sprintf("%d", field.Count),
				Short: short,
			}
			att.Fields = append(att.Fields, f)
		}
		payload.Attachments = append(payload.Attachments, att)
	}
	return payload
}

func (s *SlackAlertMethod) post(ctx context.Context, payload *Payload) error {
	data, err := jsonutil.EncodeJSON(payload)
	if err != nil {
		return err
	}
	req, err := http.NewRequest("POST", s.webhookURL, bytes.NewBuffer(data))
	req.Header.Add("Content-Type", "application/json")
	req = req.WithContext(ctx)

	resp, err := s.client.Do(req)
	if err != nil {
		return fmt.Errorf("error making HTTP request: %v", err)
	}
	resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("received non-200 status code: %s", resp.Status)
	}

	return err
}

// preprocess breaks attachments with text greater than s.textLimit
// into multiple attachments in order to prevent trucation
func (s *SlackAlertMethod) preprocess(records []*alert.Record) []*alert.Record {
	var output []*alert.Record
	for _, record := range records {
		n := len(record.Text)/s.textLimit
		if n < 1 {
			output = append(output, record)
			continue
		}
		var i int
		for i = 0; i < n; i++ {
			chopped := fmt.Sprintf("(part %d of %d)\n\n%s\n\n(continued)", i+1, n+1, record.Text[s.textLimit*i:s.textLimit*(i+1)])
			record := &alert.Record{
				Title:     fmt.Sprintf("%s (%d of %d)", record.Title, i+1, n+1),
				Text:      chopped,
				BodyField: record.BodyField,
			}
			output = append(output, record)
		}
		chopped := fmt.Sprintf("(part %d of %d)\n\n%s", i+1, n+1, record.Text[s.textLimit*i:])
		record := &alert.Record{
			Title:     fmt.Sprintf("%s (%d of %d)", record.Title, i+1, n+1),
			Text:      chopped,
			BodyField: record.BodyField,
		}
		output = append(output, record)
	}
	return output
}
