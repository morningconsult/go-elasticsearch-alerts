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

import "time"

const (
	defaultAttachmentColor      = "#36a64f"
	defaultAttachmentShort      = true
	defaultAttachmentFooter     = "Go Elasticsearch Alerts"
	defaultAttachmentFooterIcon = "https://www.elastic.co/static/images/elastic-logo-200.png"
)

type Field struct {
	Title string `json:"title"`
	Value string `json:"value"`
	Short bool   `json:"short"`
}

type AttachmentConfig struct {
	Fallback   string
	Color      string
	Title      string
	Pretext    string
	Fields     []*Field
	Text       string
	AuthorName string
	AuthorLink string
	Footer     string
	FooterIcon string
	Timestamp  int64
	MarkdownIn []string
}

type Attachment struct {
	Fallback   string   `json:"fallback"`
	Color      string   `json:"color,omitempty"`
	Title      string   `json:"title,omitempty"`
	Pretext    string   `json:"pretext,omitempty"`
	Fields     []*Field `json:"fields,omitempty"`
	Text       string   `json:"text,omitempty"`
	AuthorName string   `json:"author_name,omitempty"`
	AuthorLink string   `json:"author_link,omitempty"`
	Footer     string   `json:"footer,omitempty"`
	FooterIcon string   `json:"footer_icon,omitempty"`
	Timestamp  int64    `json:"ts,omitempty"`
	MarkdownIn []string `json:"mrkdwn_in,omitempty"`
}

func NewAttachment(config *AttachmentConfig) *Attachment {
	if config.Color == "" {
		config.Color = defaultAttachmentColor
	}

	if config.Footer == "" {
		config.Footer = defaultAttachmentFooter
	}

	if config.FooterIcon == "" {
		config.FooterIcon = defaultAttachmentFooterIcon
	}

	if config.Timestamp == 0 {
		config.Timestamp = time.Now().Unix()
	}

	return &Attachment{
		Fallback:   config.Fallback,
		Color:      config.Color,
		Title:      config.Title,
		Pretext:    config.Pretext,
		Fields:     config.Fields,
		Text:       config.Text,
		AuthorName: config.AuthorName,
		AuthorLink: config.AuthorLink,
		Footer:     config.Footer,
		FooterIcon: config.FooterIcon,
		Timestamp:  config.Timestamp,
		MarkdownIn: config.MarkdownIn,
	}
}
