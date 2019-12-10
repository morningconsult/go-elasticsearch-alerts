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

package slack

const (
	defaultAttachmentColor      = "#36a64f"
	defaultAttachmentFooter     = "Go Elasticsearch Alerts"
	defaultAttachmentFooterIcon = "https://www.elastic.co/static/images/elastic-logo-200.png"
)

// field corresponds to the 'attachment.field'
// field of a Slack message payload.
type field struct {
	Title string `json:"title"`
	Value string `json:"value"`
	Short bool   `json:"short"`
}

// attachment corresponds to the 'attachment' field
// of a Slack message payload.
type attachment struct {
	Fallback   string   `json:"fallback"`
	Color      string   `json:"color,omitempty"`
	Title      string   `json:"title,omitempty"`
	Pretext    string   `json:"pretext,omitempty"`
	Fields     []field  `json:"fields,omitempty"`
	Text       string   `json:"text,omitempty"`
	AuthorName string   `json:"author_name,omitempty"`
	AuthorLink string   `json:"author_link,omitempty"`
	Footer     string   `json:"footer,omitempty"`
	FooterIcon string   `json:"footer_icon,omitempty"`
	Timestamp  int64    `json:"ts,omitempty"`
	MarkdownIn []string `json:"mrkdwn_in,omitempty"`
}
