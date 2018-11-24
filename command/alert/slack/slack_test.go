package slack

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/hashicorp/go-uuid"
	"github.com/hashicorp/vault/helper/jsonutil"
	"gitlab.morningconsult.com/mci/go-elasticsearch-alerts/command/alert"
)

func TestNewSlackAlertMethod(t *testing.T) {
	cases := []struct{
		name   string
		config *SlackAlertMethodConfig
		err     bool
	}{
		{
			"success",
			&SlackAlertMethodConfig{
				WebhookURL: "https://example.com",
				Text:       "test",
				Channel:    "#test",
				Emoji:      ":robot:",
			},
			false,
		},
		{
			"no-webhook",
			&SlackAlertMethodConfig{
				Text: "text",
			},
			true,
		},
		{
			"no-text",
			&SlackAlertMethodConfig{
				WebhookURL: "https://example.com",
			},
			true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			s, err := NewSlackAlertMethod(tc.config)
			if tc.err {
				if err == nil {
					t.Fatal("expected an error but didn't receive one")
				}
				return
			}
			if err != nil {
				t.Fatal(err)
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

func TestWrite(t *testing.T) {
	cases := []struct{
		name    string
		status  int
		records []*alert.Record
		err     bool
	}{
		{
			"success",
			200,
			[]*alert.Record{
				&alert.Record{
					Title: "hits.hits._source",
					Text:  "{\n    \"ayy\": \"lmao\"\n}",
				},
				&alert.Record{
					Title:  "aggregations.hostname.buckets",
					Fields: []*alert.Field{
						&alert.Field{
							Key:   "foo",
							Count: 3,
						},
						&alert.Field{
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
				&alert.Record{
					Title: "hits.hits._source",
					Text:  "{\n    \"ayy\": \"lmao\"\n}",
				},
			},
			true,
		},
		{
			"non-200-response",
			201,
			[]*alert.Record{
				&alert.Record{
					Title: "hits.hits._source",
					Text:  "{\n    \"ayy\": \"lmao\"\n}",
				},
			},
			true,
		},
	}

	for _, tc := range cases {
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

			s, err := NewSlackAlertMethod(&SlackAlertMethodConfig{
				WebhookURL: url,
				Text:       "test",
			})
			if err != nil {
				t.Fatal(err)
			}

			ctx := context.Background()
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

			var data map[string]interface{}
			if err := jsonutil.DecodeJSONFromReader(r.Body, &data); err != nil {
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
