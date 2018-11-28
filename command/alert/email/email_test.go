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
			if tc.name == "password-set-in-env" {
				os.Setenv(EnvEmailAuthPassword, "random-password")
				defer os.Unsetenv(EnvEmailAuthPassword)
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
			Title: "aggregations.hostname.buckets",
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
			Title: "aggregations.hostname.buckets.program.buckets",
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
			Title: "hits.hits._source",
			Text:  "{\n   \"ayy\": \"lmao\"\n}\n----------------------------------------\n{\n    \"hello\": \"world\"\n}",
		},
	}
	eh := &EmailAlertMethod{}
	_, err := eh.buildMessage("Test Error", records)
	if err != nil {
		t.Fatal(err)
	}
}
