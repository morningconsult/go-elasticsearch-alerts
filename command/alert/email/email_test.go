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
	"os"
	"testing"

	"github.com/morningconsult/go-elasticsearch-alerts/command/alert"
)

func TestNewEmailAlertMethod(t *testing.T) {
	cases := []struct {
		name   string
		config *EmailAlertMethodConfig
		err    bool
	}{
		{
			"success",
			&EmailAlertMethodConfig{
				Host: "smtp.gmail.com",
				Port: 587,
				From: "test@gmail.com",
				To: []string{
					"test_recipient_1@gmail.com",
					"test_recipient_2@gmail.com",
				},
				Username: "test@gmail.com",
				Password: "password",
			},
			false,
		},
		{
			"password-set-in-env",
			&EmailAlertMethodConfig{
				Host: "smtp.gmail.com",
				Port: 587,
				From: "test@gmail.com",
				To: []string{
					"test_recipient_1@gmail.com",
					"test_recipient_2@gmail.com",
				},
				Username: "test@gmail.com",
			},
			false,
		},
		{
			"username-set-in-env",
			&EmailAlertMethodConfig{
				Host: "smtp.gmail.com",
				Port: 587,
				From: "test@gmail.com",
				To: []string{
					"test_recipient_1@gmail.com",
					"test_recipient_2@gmail.com",
				},
				Password: "test",
			},
			false,
		},
		{
			"missing-required-fields",
			&EmailAlertMethodConfig{},
			true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			switch tc.name {
			case "password-set-in-env":
				os.Setenv(EnvEmailAuthPassword, "random-password")
				defer os.Unsetenv(EnvEmailAuthPassword)
			case "username-set-in-env":
				os.Setenv(EnvEmailAuthUsername, "test@gmail.com")
				defer os.Unsetenv(EnvEmailAuthUsername)
			default:
			}
			_, err := NewEmailAlertMethod(tc.config)
			if tc.err {
				if err == nil {
					t.Fatal("expected an error but didn't receive one")
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestBuildMessage(t *testing.T) {
	records := []*alert.Record{
		&alert.Record{
			Filter: "aggregations.hostname.buckets",
			Text:  "",
			Fields: []*alert.Field{
				&alert.Field{
					Key:   "foo",
					Count: 10,
				},
				&alert.Field{
					Key:   "bar",
					Count: 8,
				},
			},
		},
		&alert.Record{
			Filter: "aggregations.hostname.buckets.program.buckets",
			Text:  "",
			Fields: []*alert.Field{
				&alert.Field{
					Key:   "foo - bim",
					Count: 3,
				},
				&alert.Field{
					Key:   "foo - baz",
					Count: 7,
				},
				&alert.Field{
					Key:   "bar - hello",
					Count: 6,
				},
				&alert.Field{
					Key:   "bar - world",
					Count: 2,
				},
			},
		},
		&alert.Record{
			Filter: "hits.hits._source",
			Text:  "{\n   \"ayy\": \"lmao\"\n}\n----------------------------------------\n{\n    \"hello\": \"world\"\n}",
		},
	}

	expected := `Content-Type: text/html
Subject: Go Elasticsearch Alerts: Test Error

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
<h4>Filter path: aggregations.hostname.buckets</h4>
<table>
  <tr>
    <th>Key</th>
    <th>Count</th>
  </tr>
  <tr>
    <td>foo</td>
    <td>10</td>
  </tr>
  <tr>
    <td>bar</td>
    <td>8</td>
  </tr>
</table>

<br><h4>Filter path: aggregations.hostname.buckets.program.buckets</h4>
<table>
  <tr>
    <th>Key</th>
    <th>Count</th>
  </tr>
  <tr>
    <td>foo - bim</td>
    <td>3</td>
  </tr>
  <tr>
    <td>foo - baz</td>
    <td>7</td>
  </tr>
  <tr>
    <td>bar - hello</td>
    <td>6</td>
  </tr>
  <tr>
    <td>bar - world</td>
    <td>2</td>
  </tr>
</table>

<br><h4>Filter path: hits.hits._source</h4>
{<br>&nbsp;&nbsp;&nbsp;&#34;ayy&#34;:&nbsp;&#34;lmao&#34;<br>}<br>----------------------------------------<br>{<br>&nbsp;&nbsp;&nbsp;&nbsp;&#34;hello&#34;:&nbsp;&#34;world&#34;<br>}
<br>
</body>
</html>`

	eh := &EmailAlertMethod{}
	msg, err := eh.buildMessage("Test Error", records)
	if err != nil {
		t.Fatal(err)
	}
	if msg != expected {
		t.Errorf("Got:\n%s\n\nExpected:\n%s", msg, expected)
	}
}
