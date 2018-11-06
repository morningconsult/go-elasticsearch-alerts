package config

type Input struct {
	Host       string
	Index      string
	Query      map[string]interface{}
	TLSEnabled bool
}
