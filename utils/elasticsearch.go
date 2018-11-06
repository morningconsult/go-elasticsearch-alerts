package utils

import (
	"bytes"
	"context"
	"fmt"
	"net/http"

	"github.com/hashicorp/vault/helper/jsonutil"
	"gitlab.morningconsult.com/mci/go-elasticsearch-alert/config"
)

func Query(ctx context.Context, client *http.Client, input *config.Input) (map[string]interface{}, error) {
	if client == nil {
		return nil, fmt.Errorf("error making ElasticSearch request: input is nil")
	}
	if client == nil {
		return nil, fmt.Errorf("error making ElasticSearch request: HTTP client is nil")
	}

	data, err := jsonutil.EncodeJSON(input.Query)
	if err != nil {
		return nil, fmt.Errorf("error JSON-encoding ElasticSearch query: %v", err)
	}

	proto := "http"
	if input.TLSEnabled {
		proto = "https"
	}

	req, err := http.NewRequest("GET", fmt.Sprintf("%s://%s/%s", proto, input.Host, input.Index), bytes.NewBuffer(data)).WithContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("error creating *http.Request: %v", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error making HTTP request: %v", err)
	}
	defer resp.Body.Close()

	var output = make(map[string]interface{})
	if err = jsonutil.DecodeJSONFromReader(resp.Body, &output); err != nil {
		return nil, fmt.Errorf("error JSON-decoding response from ElasticSearch: %v", err)
	}

	return output, nil
}
