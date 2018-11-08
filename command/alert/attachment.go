package alert

const (
	defaultColor string  = "#36a64f"
	defaultShort bool    = true
	defaultFooter string = "#data"
)

type Field struct {
	Title string `json:"title"`
	Value string `json:"value"`
	Short bool   `json:"short"`
}

type Attachment struct {
	Fallback string   `json:"fallback"`
	Color    string   `json:"color"`
	Pretext  string   `json:"pretext"`
	Fields   []*Field `json:"fields"`
	Footer   string   `json:"footer"`
}