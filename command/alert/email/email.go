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

// Package email does stuff!
package email

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"html/template"
	"net/smtp"
	"os"
	"strings"

	"github.com/morningconsult/go-elasticsearch-alerts/command/alert"
)

const (
	EnvEmailAuthUsername = "GO_ELASTICSEARCH_ALERTS_SMTP_USERNAME"
	EnvEmailAuthPassword = "GO_ELASTICSEARCH_ALERTS_SMTP_PASSWORD"
)

var _ alert.AlertMethod = (*EmailAlertMethod)(nil)

type EmailAlertMethodConfig struct {
	Host     string   `mapstructure:"host"`
	Port     int      `mapstructure:"port"`
	From     string   `mapstructure:"from"`
	To       []string `mapstructure:"to"`
	Username string   `mapstructure:"username"`
	Password string   `mapstructure:"password"`
}

// NewEmailAlertMethod creates a new *EmailAlertMethod or a
// non-nil error if there was an error.
func NewEmailAlertMethod(config *EmailAlertMethodConfig) (*EmailAlertMethod, error) {
	if config == nil {
		return nil, errors.New("no config provided")
	}

	errors := []string{}
	if config.Host == "" {
		errors = append(errors, "no SMTP host provided")
	}

	if config.Port == 0 {
		errors = append(errors, "no SMTP port provided")
	}

	if config.From == "" {
		errors = append(errors, "no sender address provided")
	}

	if len(config.To) < 1 {
		errors = append(errors, "no recipient address(es) provided")
	}

	if u := os.Getenv(EnvEmailAuthUsername); u != "" {
		config.Username = u
	}

	if config.Username == "" {
		errors = append(errors, "no SMTP username provided in configuration file or environment")
	}

	if p := os.Getenv(EnvEmailAuthPassword); p != "" {
		config.Password = p
	}

	if config.Password == "" {
		errors = append(errors, "no SMTP password provided in configuration file or environment")
	}

	if len(errors) > 0 {
		return nil, fmt.Errorf("errors with your email output configuration:\n* %s", strings.Join(errors, "\n* "))
	}

	return &EmailAlertMethod{
		host: config.Host,
		port: config.Port,
		from: config.From,
		to:   config.To,
		auth: smtp.PlainAuth("", config.Username, config.Password, config.Host),
	}, nil
}

type EmailAlertMethod struct {
	host string
	port int
	from string
	auth smtp.Auth
	to   []string
}

// Write creates an email message from the records and sends
// it to the email address(es) specified at the creation of the
// EmailAlertMethod. If there was an error sending the email,
// it returns a non-nil error.
func (e *EmailAlertMethod) Write(ctx context.Context, rule string, records []*alert.Record) error {
	body, err := e.BuildMessage(rule, records)
	if err != nil {
		return fmt.Errorf("error creating email message: %v", err)
	}
	return smtp.SendMail(fmt.Sprintf("%s:%d", e.host, e.port), e.auth, e.from, e.to, []byte(body))
}

// BuildMessage creates an email message from the provided
// records. It will return a non-nil error if an error occurs.
func (e *EmailAlertMethod) BuildMessage(rule string, records []*alert.Record) (string, error) {
	alert := struct {
		Name    string
		Records []*alert.Record
	}{
		rule,
		records,
	}

	funcs := template.FuncMap{
		"tabsAndLines": func(text string) template.HTML {
			return template.HTML(strings.Replace(strings.Replace(template.HTMLEscapeString(text), "\n", "<br>", -1), " ", "&nbsp;", -1))
		},
	}

	tpl := `Content-Type: text/html
Subject: Go Elasticsearch Alerts: {{ .Name }}

<!DOCTYPE html>
<html>
<head>
<style>
table {
    font-family: arial, sans-serif;
    border-collapse: collapse;
}

td, th {
    border: 1px solid #dddddd;
    text-align: left;
    padding: 8px;
}

tr:nth-child(even) {
    background-color: #dddddd;
}
</style>
</head>
<body>
{{ range .Records }}<h4>Filter path: {{ .Filter }}</h4>{{ if .Fields }}
<table>
  <tr>
    <th>Key</th>
    <th>Count</th>
  </tr>{{ range .Fields }}
  <tr>
    <td>{{ .Key }}</td>
    <td>{{ .Count }}</td>
  </tr>{{ end }}
</table>{{ end }}
{{ tabsAndLines .Text }}
<br>{{ end }}
</body>
</html>`
	t, err := template.New("email").Funcs(funcs).Parse(tpl)
	if err != nil {
		return "", fmt.Errorf("error parsing email template: %v", err)
	}

	buf := &bytes.Buffer{}
	err = t.Execute(buf, alert)
	if err != nil {
		return "", fmt.Errorf("error executing email template: %v", err)
	}
	return buf.String(), nil
}
