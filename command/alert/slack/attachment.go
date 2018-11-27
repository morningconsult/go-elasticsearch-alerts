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

const (
	defaultAttachmentColor string = "#36a64f"
	defaultAttachmentShort bool = true
	defaultAttachmentFooter string = "#data"
)

type Field struct {
	Title string `json:"title"`
	Value string `json:"value"`
	Short bool   `json:"short"`
}

type AttachmentConfig struct {
	Fallback string
	Color    string
	Pretext  string
	Fields   []*Field
	Text     string
	Footer   string
}

type Attachment struct {
	Fallback string   `json:"fallback"`
	Color    string   `json:"color,omitempty"`
	Pretext  string   `json:"pretext,omitempty"`
	Fields   []*Field `json:"fields,omitempty"`
	Text     string   `json:"text,omitempty"`
	Footer   string   `json:"footer,omitempty"`
}

func NewAttachment(config *AttachmentConfig) *Attachment {
	if config.Color == "" {
		config.Color = defaultAttachmentColor
	}

	if config.Footer == "" {
		config.Footer = defaultAttachmentFooter
	}

	return &Attachment{
		Fallback: config.Fallback,
		Color:    config.Color,
		Pretext:  config.Pretext,
		Fields:   config.Fields,
		Text:     config.Text,
		Footer:   config.Footer,
	}
}