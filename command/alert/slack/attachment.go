package slack

const (
	defaultAttachmentColor string = "#36a64f"
	defaultAttachmentShort bool = true
	defaultAttachmentFooter string = "#data"
)

type Field struct {
	Title string `json:"title"`
	Value string `json:"value"`
	Short bool   `json:"short"`
}

type AttachmentConfig struct {
	Fallback string
	Color    string
	Pretext  string
	Fields   []*Field
	Text     string
	Footer   string
}

type Attachment struct {
	Fallback string   `json:"fallback"`
	Color    string   `json:"color,omitempty"`
	Pretext  string   `json:"pretext,omitempty"`
	Fields   []*Field `json:"fields,omitempty"`
	Text     string   `json:"text,omitempty"`
	Footer   string   `json:"footer,omitempty"`
}

func NewAttachment(config *AttachmentConfig) *Attachment {
	if config.Color == "" {
		config.Color = defaultAttachmentColor
	}

	if config.Footer == "" {
		config.Footer = defaultAttachmentFooter
	}

	return &Attachment{
		Fallback: config.Fallback,
		Color:    config.Color,
		Pretext:  config.Pretext,
		Fields:   config.Fields,
		Text:     config.Text,
		Footer:   config.Footer,
	}
}