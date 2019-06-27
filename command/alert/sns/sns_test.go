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

	"github.com/Masterminds/sprig"
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
			"more-custom-templating",
			"{{range .}}{{if ne .Text \"\"}}{{range $_, $v := regexFindAll \"\\\\[ERROR\\\\].*\" .Text -1 | uniq }}* {{$v | trimAll \"\\\",\"}}\n{{end}}{{end}}{{end}}",
			[]*alert.Record{
				{
					Filter: "aggregations.programs.buckets",
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
				{
					Filter: "hits.hits._source",
					Text: `
{
  "system" : {
    "syslog" : {
      "hostname" : "ip-172-31-71-249",
      "pid" : "1317",
      "program" : "update",
      "message" : """2019-06-27T17:30:05.710Z#011test/main.go:10#011[ERROR] Error launching: bad stuff happened""",
      "timestamp" : "Jun 27 17:30:05"
    }
  },
    "host" : {
    "name" : "ip-172-31-71-249"
  }
}
----------------------------------------
{
  "system" : {
    "syslog" : {
      "hostname" : "ip-172-31-71-249",
      "pid" : "1317",
      "program" : "update",
      "message" : """2019-06-27T17:30:05.710Z#011test/main.go:10#011[ERROR] Error launching: more bad stuff happened""",
      "timestamp" : "Jun 27 17:30:05"
    }
  },
    "host" : {
    "name" : "ip-172-31-71-249"
  }
}
----------------------------------------
{
  "system" : {
    "syslog" : {
      "hostname" : "ip-172-31-71-249",
      "pid" : "1317",
      "program" : "update",
      "message" : """2019-06-27T17:30:05.710Z#011test/main.go:10#011[ERROR] Error launching: more bad stuff happened""",
      "timestamp" : "Jun 27 17:30:05"
    }
  },
    "host" : {
    "name" : "ip-172-31-71-249"
  }
}
----------------------------------------
{
    "host": {
        "name": "ip-172-31-54-5"
    },
    "system": {
        "syslog": {
            "hostname": "ip-172-31-54-5",
            "message": "2019-06-27T07:01:02.395Z#011warn#011test/main.go:256#011[WARN ] Not gonna do anything.",
            "pid": "1267",
            "program": "close",
            "timestamp": "Jun 27 07:01:02"
        }
    }
}
`,
					BodyField: true,
				},
			},
			false,
			"[TEST ERROR ALERT]\n* [ERROR] Error launching: bad stuff happened\n* [ERROR] Error launching: more bad stuff happened\n",
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
		{
			"no-records-matching-template-logic",
			"{{range .}}{{if ne .Text \"\"}}{{range $_, $v := regexFindAll \"\\\\[ERROR\\\\].*\" .Text -1 | uniq }}* {{$v | trimAll \"\\\",\"}}\n{{end}}{{end}}{{end}}",
			defaultRecords,
			false,
			"[TEST ERROR ALERT]\nNew alerts detected. See logs.",
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			a := &AlertMethod{
				template: template.Must(template.New("test").Funcs(template.FuncMap(sprig.FuncMap())).Parse(tc.template)),
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
