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
	"testing"
	"text/template"

	"github.com/morningconsult/go-elasticsearch-alerts/command/alert"
)

func TestAlertMethod_renderTemplate(t *testing.T) {
	defaultRecords := []*alert.Record{
		{
			Filter: "foo.bar.bim",
			Fields: []*alert.Field{
				{
					Key:   "test-1",
					Count: 2,
				},
				{
					Key:   "test-2",
					Count: 4,
				},
			},
		},
	}

	cases := []struct {
		name      string
		template  string
		records   []*alert.Record
		expectErr bool
		expectMsg string
	}{
		{
			"valid-template",
			"{{range .}}{{.Filter}}:\n{{range .Fields}}* {{.Key}}: {{.Count}}\n{{end}}\n{{end}}",
			defaultRecords,
			false,
			"[TEST ERROR ALERT]\nfoo.bar.bim:\n* test-1: 2\n* test-2: 4\n\n",
		},
		{
			"invalid-template",
			"Filter: {{.Filter}}", // this will cause template.Execute to fail
			defaultRecords,
			true,
			"",
		},
		{
			"no-templating-logic",
			"Hit the deck!", // No templating logic in template string
			defaultRecords,
			false,
			"[TEST ERROR ALERT]\nHit the deck!",
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			a := &AlertMethod{
				template: template.Must(template.New("test").Parse(tc.template)),
			}
			msg, err := a.renderTemplate("TEST ERROR ALERT", tc.records)
			if tc.expectErr {
				if err == nil {
					t.Fatal("Expected an error")
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if msg != tc.expectMsg {
				t.Errorf("Expected message:\n%s\nGot:\n%s", tc.expectMsg, msg)
			}
		})
	}
}
