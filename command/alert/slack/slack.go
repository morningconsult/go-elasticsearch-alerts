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

	"github.com/hashicorp/go-cleanhttp"
	"github.com/hashicorp/vault/helper/jsonutil"
	"gitlab.morningconsult.com/mci/go-elasticsearch-alerts/command/alert"
)

// Ensure SlackAlertMethod adheres to the alert.AlertMethod interface
var _ alert.AlertMethod = (*SlackAlertMethod)(nil)

type SlackAlertMethodConfig struct {
	WebhookURL string       `mapstructure:"webhook"`
	Channel    string       `mapstructure:"channel"`
	Username   string       `mapstructure:"username"`
	Text       string       `mapstructure:"text"`
	Emoji      string       `mapstructure:"emoji"`
	Client     *http.Client
}

type SlackAlertMethod struct {
	webhookURL string
	client     *http.Client
	channel    string
	username   string
	text       string
	emoji      string
}

type Payload struct {
	Channel     string        `json:"channel,omitempty"`
	Username    string        `json:"username,omitempty"`
	Text        string        `json:"text,omitempty"`
	Emoji       string        `json:"icon_emoji,omitempty"`
	Attachments []*Attachment `json:"attachments,omitempty"`
}

func NewSlackAlertMethod(config *SlackAlertMethodConfig) (*SlackAlertMethod, error) {
	if config.WebhookURL == "" {
		return nil, fmt.Errorf("field 'output.config.webhook' must not be empty when using the Slack output method")
	}

	if config.Client == nil {
		config.Client = cleanhttp.DefaultClient()
	}

	if config.Text == "" {
		return nil, fmt.Errorf("field 'config.test' must not be empty when using the Slack output method")
	}

	return &SlackAlertMethod{
		channel:    config.Channel,
		webhookURL: config.WebhookURL,
		client:     config.Client,
		text:       config.Text,
		emoji:      config.Emoji,
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

	for _, record := range records {
		att := NewAttachment(&AttachmentConfig{
			Fallback: rule,
			Pretext:  record.Title,
			Text:     record.Text,
		})

		for _, field := range record.Fields {
			f := &Field{
				Title: field.Key,
				Value: fmt.Sprintf("%d", field.Count),
				Short: true,
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