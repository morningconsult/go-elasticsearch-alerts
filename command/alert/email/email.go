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

package email

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"net/smtp"
	"strings"
	"html/template"

	"gitlab.morningconsult.com/mci/go-elasticsearch-alerts/command/alert"
)

const EnvEmailAuthPassword = "GO_ELASTICSEARCH_ALERTS_SMTP_PASSWORD"

var _ alert.AlertMethod = (*EmailAlertMethod)(nil)

type EmailAlertMethodConfig struct {
	Address  string   `mapstructure:"address"`
	From     string   `mapstructure:"from"`
	To       []string `mapstructure:"to"`
	AuthHost string   `mapstructure:"auth_host"`
	Password string   `mapstructure:"password"`
}

func NewEmailAlertMethod(config *EmailAlertMethodConfig) (*EmailAlertMethod, error) {
	errors := []string{}
	if config.Address == "" {
		errors = append(errors, "no SMTP host provided (must be in <host>:<port> format)")
	}

	if config.From == "" {
		errors = append(errors, "no sender address provided")
	}

	if len(config.To) < 1 {
		errors = append(errors, "no recipient address(es) provided")
	}

	if config.AuthHost == "" {
		errors = append(errors, "no authentication host provided")
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
		address:  config.Address,
		from:     config.From,
		to:       config.To,
		authHost: config.AuthHost,
		password: config.Password,
	}, nil
}

type EmailAlertMethod struct {
	address  string
	from     string
	password string
	to       []string
	authHost string
}

func (e *EmailAlertMethod) Write(ctx context.Context, rule string, records []*alert.Record) error {
	body, err := e.buildMessage(rule, records)
	if err != nil {
		return fmt.Errorf("error creating email message: %v", err)
	}
	auth := smtp.PlainAuth("", e.from, e.password, e.authHost)
	return smtp.SendMail(e.address, auth, e.from, e.to, []byte(body))
}

func (e *EmailAlertMethod) buildMessage(rule string, records []*alert.Record) (string, error) {
	alert := struct {
		Name string
		Records []*alert.Record
	}{
		rule,
		records,
	}

	funcs := template.FuncMap{
		"tabsAndLines": func(text string) template.HTML {
			return template.HTML(strings.Replace(strings.Replace(template.HTMLEscapeString(text), "\n", "<br>",  -1), " ", "&nbsp;", -1))
		},
	}

	tpl := `Content-Type: text/html
Subject: Go ElasticSearch Alerts: {{ .Name }}

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
{{ range .Records }}<h4>Filter path: {{ .Title }}</h4>{{ if .Fields }}
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
