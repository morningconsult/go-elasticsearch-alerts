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

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
	"time"

	uuid "github.com/hashicorp/go-uuid"

	"github.com/morningconsult/go-elasticsearch-alerts/command/alert"
)

func TestNewAlertMethod(t *testing.T) {
	cases := []struct {
		name   string
		config *AlertMethodConfig
		err    bool
	}{
		{
			"success",
			&AlertMethodConfig{
				WebhookURL: "https://example.com",
				Text:       "test",
				Channel:    "#test",
				Emoji:      ":robot:",
			},
			false,
		},
		{
			"nil-config",
			nil,
			true,
		},
		{
			"no-webhook",
			&AlertMethodConfig{
				Text: "text",
			},
			true,
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			a, err := NewAlertMethod(tc.config)
			if tc.err {
				if err == nil {
					t.Fatal("expected an error but didn't receive one")
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}

			s, ok := a.(*AlertMethod)
			if !ok {
				t.Fatalf("Expected type *AlertMethod")
			}

			if s.channel != tc.config.Channel {
				t.Fatalf("got unexpected channel (got %q, expected %q)", s.channel, tc.config.Channel)
			}

			if s.webhookURL != tc.config.WebhookURL {
				t.Fatalf("got unexpected webhook URL (got %q, expected %q)", s.webhookURL, tc.config.WebhookURL)
			}

			if s.text != tc.config.Text {
				t.Fatalf("got unexpected text value (got %q, expected %q)", s.text, tc.config.Text)
			}

			if s.emoji != tc.config.Emoji {
				t.Fatalf("got unexpected emoji value (got %q, expected %q)", s.emoji, tc.config.Emoji)
			}
		})
	}
}

func TestBuildPayload(t *testing.T) {
	rule := "Test Rule"
	filter := "test.filter"
	cases := []struct {
		name     string
		records  []*alert.Record
		expected payload
	}{
		{
			"pagination",
			[]*alert.Record{
				{
					Filter:    filter,
					Text:      `Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat. Duis aute irure dolor in reprehenderit in voluptate velit esse cillum dolore eu fugiat nulla pariatur. Excepteur sint occaecat cupidatat non proident, sunt in culpa qui officia deserunt mollit anim id est laborum.`,
					BodyField: true,
				},
			},
			payload{
				Attachments: []attachment{
					{
						Title:      rule,
						Text:       filter + " (1 of 3)\n```\n(part 1 of 3)\n\nLorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris nisi ut a\n\n(continued)\n```",
						MarkdownIn: []string{"text"},
						Color:      "#ff0000",
						Footer:     "Go Elasticsearch Alerts",
						FooterIcon: "https://www.elastic.co/static/images/elastic-logo-200.png",
						Timestamp:  time.Now().Unix(),
					},
					{
						Title:      rule,
						Text:       filter + " (2 of 3)\n```\n(part 2 of 3)\n\nliquip ex ea commodo consequat. Duis aute irure dolor in reprehenderit in voluptate velit esse cillum dolore eu fugiat nulla pariatur. Excepteur sint occaecat cupidatat non proident, sunt in culpa qui\n\n(continued)\n```",
						MarkdownIn: []string{"text"},
						Color:      "#ff0000",
						Footer:     "Go Elasticsearch Alerts",
						FooterIcon: "https://www.elastic.co/static/images/elastic-logo-200.png",
						Timestamp:  time.Now().Unix(),
					},
					{
						Title:      rule,
						Text:       filter + " (3 of 3)\n```\n(part 3 of 3)\n\n officia deserunt mollit anim id est laborum.\n```",
						MarkdownIn: []string{"text"},
						Color:      "#ff0000",
						Footer:     "Go Elasticsearch Alerts",
						FooterIcon: "https://www.elastic.co/static/images/elastic-logo-200.png",
						Timestamp:  time.Now().Unix(),
					},
				},
			},
		},
		{
			"builds-fields",
			[]*alert.Record{
				{
					Filter: filter,
					Fields: []*alert.Field{
						{
							Key:   "foo",
							Count: 8,
						},
						{
							Key:   "bar",
							Count: 2,
						},
					},
				},
			},
			payload{
				Attachments: []attachment{
					{
						Title:      rule,
						Text:       filter,
						MarkdownIn: []string{"text"},
						Footer:     "Go Elasticsearch Alerts",
						FooterIcon: "https://www.elastic.co/static/images/elastic-logo-200.png",
						Timestamp:  time.Now().Unix(),
						Color:      defaultAttachmentColor,
						Fields: []field{
							{
								Title: "foo",
								Value: "8",
								Short: true,
							},
							{
								Title: "bar",
								Value: "2",
								Short: true,
							},
						},
					},
				},
			},
		},
	}

	s := &AlertMethod{
		textLimit: 200,
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			payload := s.buildPayload(rule, tc.records)
			if !reflect.DeepEqual(tc.expected.Attachments, payload.Attachments) {
				t.Fatalf("Got Payload.Attachments:\n%+v\n\nExpected Payload.Attachments:\n%+v\n",
					prettyJSON(t, payload.Attachments),
					prettyJSON(t, tc.expected.Attachments),
				)
			}
		})
	}
}

func TestWrite(t *testing.T) {
	cases := []struct {
		name    string
		status  int
		records []*alert.Record
		err     bool
	}{
		{
			"success",
			200,
			[]*alert.Record{
				{
					Filter: "hits.hits._source",
					Text:   "{\n    \"ayy\": \"lmao\"\n}",
				},
				{
					Filter: "aggregations.hostname.buckets",
					Fields: []*alert.Field{
						{
							Key:   "foo",
							Count: 3,
						},
						{
							Key:   "bar",
							Count: 2,
						},
					},
				},
			},
			false,
		},
		{
			"no-records",
			200,
			[]*alert.Record{},
			false,
		},
		{
			"wrong-URL",
			200,
			[]*alert.Record{
				{
					Filter: "hits.hits._source",
					Text:   "{\n    \"ayy\": \"lmao\"\n}",
				},
			},
			true,
		},
		{
			"non-200-response",
			201,
			[]*alert.Record{
				{
					Filter: "hits.hits._source",
					Text:   "{\n    \"ayy\": \"lmao\"\n}",
				},
			},
			true,
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			ts := newMockSlackServer(tc.status)

			defer ts.Close()

			var url string
			if tc.name == "wrong-URL" {
				id, err := uuid.GenerateUUID()
				if err != nil {
					t.Fatal(err)
				}
				url = fmt.Sprintf("http://bad.%s.co.nz", id)
			} else {
				url = ts.URL
			}

			s, err := NewAlertMethod(&AlertMethodConfig{
				WebhookURL: url,
				Text:       "test",
			})
			if err != nil {
				t.Fatal(err)
			}

			ctx := t.Context()
			err = s.Write(ctx, "test-rule", tc.records)
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

func newMockSlackServer(status int) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "POST":
			ct := r.Header.Get("Content-Type")
			if ct == "" || ct != "application/json" {
				http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
				return
			}

			var data map[string]any
			if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
				http.Error(w, "Internal Server Error", 500)
				return
			}

			text, ok := data["text"].(string)
			if !ok || text == "" {
				http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
				return
			}
			w.WriteHeader(status)
			w.Write([]byte("OK"))
			return
		default:
			http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
			return
		}
	}))
}

func prettyJSON(t *testing.T, v any) string {
	data, err := json.MarshalIndent(v, "", "    ")
	if err != nil {
		t.Fatal(err)
	}
	return string(data)
}

func ExampleAlertMethod_buildPayload() {
	records := []*alert.Record{
		{
			Filter:    "hits.hits._source",
			Text:      `Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat. Duis aute irure dolor in reprehenderit in voluptate velit esse cillum dolore eu fugiat nulla pariatur. Excepteur sint occaecat cupidatat non proident, sunt in culpa qui officia deserunt mollit anim id est laborum.`,
			BodyField: true,
		},
		{
			Filter: "aggregation.hostname.buckets",
			Fields: []*alert.Field{
				{
					Key:   "foo",
					Count: 2,
				},
				{
					Key:   "bar",
					Count: 3,
				},
			},
		},
	}

	a, _ := NewAlertMethod(&AlertMethodConfig{
		WebhookURL: "https://hooks.slack.com/services/ABCDEFG",
		Channel:    "#alerts",
		Text:       "New alert!",
		Emoji:      ":robot",
		TextLimit:  200,
	})

	sm := a.(*AlertMethod)

	payload := sm.buildPayload("Test rule", records)

	// This loop is performed in order that tests will pass --
	// it is not necessary to perform this
	for i := range payload.Attachments {
		payload.Attachments[i].Timestamp = 1
	}
	data, err := json.MarshalIndent(payload, "", "    ")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(string(data))

	// Output:
	// {
	//     "channel": "#alerts",
	//     "text": "New alert!",
	//     "icon_emoji": ":robot",
	//     "attachments": [
	//         {
	//             "fallback": "",
	//             "color": "#ff0000",
	//             "title": "Test rule",
	//             "text": "hits.hits._source (1 of 3)\n```\n(part 1 of 3)\n\nLorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris nisi ut a\n\n(continued)\n```",
	//             "footer": "Go Elasticsearch Alerts",
	//             "footer_icon": "https://www.elastic.co/static/images/elastic-logo-200.png",
	//             "ts": 1,
	//             "mrkdwn_in": [
	//                 "text"
	//             ]
	//         },
	//         {
	//             "fallback": "",
	//             "color": "#ff0000",
	//             "title": "Test rule",
	//             "text": "hits.hits._source (2 of 3)\n```\n(part 2 of 3)\n\nliquip ex ea commodo consequat. Duis aute irure dolor in reprehenderit in voluptate velit esse cillum dolore eu fugiat nulla pariatur. Excepteur sint occaecat cupidatat non proident, sunt in culpa qui\n\n(continued)\n```",
	//             "footer": "Go Elasticsearch Alerts",
	//             "footer_icon": "https://www.elastic.co/static/images/elastic-logo-200.png",
	//             "ts": 1,
	//             "mrkdwn_in": [
	//                 "text"
	//             ]
	//         },
	//         {
	//             "fallback": "",
	//             "color": "#ff0000",
	//             "title": "Test rule",
	//             "text": "hits.hits._source (3 of 3)\n```\n(part 3 of 3)\n\n officia deserunt mollit anim id est laborum.\n```",
	//             "footer": "Go Elasticsearch Alerts",
	//             "footer_icon": "https://www.elastic.co/static/images/elastic-logo-200.png",
	//             "ts": 1,
	//             "mrkdwn_in": [
	//                 "text"
	//             ]
	//         },
	//         {
	//             "fallback": "",
	//             "color": "#36a64f",
	//             "title": "Test rule",
	//             "fields": [
	//                 {
	//                     "title": "foo",
	//                     "value": "2",
	//                     "short": true
	//                 },
	//                 {
	//                     "title": "bar",
	//                     "value": "3",
	//                     "short": true
	//                 }
	//             ],
	//             "text": "aggregation.hostname.buckets",
	//             "footer": "Go Elasticsearch Alerts",
	//             "footer_icon": "https://www.elastic.co/static/images/elastic-logo-200.png",
	//             "ts": 1,
	//             "mrkdwn_in": [
	//                 "text"
	//             ]
	//         }
	//     ]
	// }
}
