package email

import (
	"os"
	"testing"

	"gitlab.morningconsult.com/mci/go-elasticsearch-alerts/command/alert"
)

func TestNewEmailAlertMethod(t *testing.T) {
	cases := []struct{
		name string
		config *EmailAlertMethodConfig
		err bool
	}{
		{
			"success",
			&EmailAlertMethodConfig{
				Address: "smtp.gmail.com:587",
				From:    "test@gmail.com",
				To:      []string{
					"test_recipient_1@gmail.com",
					"test_recipient_2@gmail.com",
				},
				AuthHost: "smtp.gmail.com",
				Password: "password",
			},
			false,
		},
		{
			"password-set-in-env",
			&EmailAlertMethodConfig{
				Address: "smtp.gmail.com:587",
				From:    "test@gmail.com",
				To:      []string{
					"test_recipient_1@gmail.com",
					"test_recipient_2@gmail.com",
				},
				AuthHost: "smtp.gmail.com",
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
			Title:  "aggregations.hostname.buckets",
			Text:   "",
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
			Title:  "aggregations.hostname.buckets.program.buckets",
			Text:   "",
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
