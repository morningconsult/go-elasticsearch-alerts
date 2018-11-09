package slack

import (
	"context"
	"net/http"

	"github.com/hashicorp/go-cleanhttp"
	"gitlab.morningconsult.com/mci/go-elasticsearch-alerts/command/alert"
)

const (
	defaultChannel  string = "#error-alerts"
	defaultUsername string = "go-alerts"
	defaultEmoji    string = ":robot_face:"
)

// Ensure Helper adheres to the alert.AlertHandler interface
var _ alert.AlertMethod = (*slackAlertMethod)(nil)

type SlackAlertMethodConfig {
	WebhookURL string
	Client     *http.Client
	Channel    string
	Username   string
	Text       string
	Emoji      string
}

type slackAlertMethod struct {
	webhookURL string
	client     *http.Client
	channel    string
	username   string
	text       string
	emoji      string
}

type Payload struct {
	Channel     string        `json:"channel"`
	Username    string        `json:"username,omitempty"`
	Text        string        `json:"text,omitempty"`
	Emoji       string        `json:"icon_emoji,omitempty"`
	Attachments []*Attachment `json:"attachments,omitempty"`
}

func NewSlackAlertMethod(config *SlackAlertHandlerConfig) *slackAlertMethod {
	if config.Client == nil {
		config.Client = cleanhttp.DefaultClient()
	}

	if config.Channel == "" {
		config.channel = defaultChannel
	}

	if config.Username == "" {
		config.username = defaultUsername
	}

	if config.Emoji == "" {
		config.emoji = defaultEmoji
	}

	return &slackAlertMethod{
		webhookURL: config.WebhookURL,
		client:     config.Client,
		text:       config.Text,
		emoji:      config.Emoji,
	}
}

func (s *slackAlertMethod) Send(ctx context.Context, records []*alert.Record) error {
	payload := &Payload{
		Channel:  s.channel,
		Username: s.username,
		Text:     s.text,
		Emoji:    s.emoji,
	}

	for _, record := range records {
		att := NewAttachment(&config.AttachmentConfig{
			Fallback: record.Title,
			Pretext:  record.Title,
			Text:     record.Text,
		})

		for _, field := range record.Fields {
			f := &Field{
				Title: field.Key,
				Value: field.Count,
				Short: true,
			}
			att.Fields = append(att.Fields, f)
		}

		payload.Attachments = append(payload.Attachments, att)
	}

	req, err := http.NewRequest("POST", )
}