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

package config

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/hashicorp/go-cleanhttp"
)

// ClientConfig is used to create new configured new
// *http.Client instances.
type ClientConfig struct {
	// TLSEnabled is used to inform NewESClient whether to
	// communicate with Elasticsearch via TLS. This value
	// should come from the 'elasticsearch.client.tls_enabled'
	// field of the main configuration file
	TLSEnabled bool `json:"tls_enabled"`

	// CACert is the path to a PEM-encoded CA certificate file.
	// This value should come from the 'elasticsearch.client.ca_cert'
	// field of the main configuration file
	CACert string `json:"ca_cert"`

	// ClientCert is the path to a PEM-encoded client
	// certificate file when connecting via TLS. This value
	// should come from the 'elasticsearch.client.client_cert'
	// field of the main configuration file
	ClientCert string `json:"client_cert"`

	// ClientKey is the path to a PEM-encoded client key file
	// when connecting via TLS. This value should come from the
	// 'elasticsearch.client.client_key' field of the main
	// configuration file
	ClientKey string `json:"client_key"`

	// ServerName is the server name to use as the SNI host when
	// connecting via TLS. This value should come from the
	// 'elasticsearch.client.server_name' field of the main
	// configuration file
	ServerName string `json:"server_name"`
}

// NewESClient creates a new HTTP client based on the
// values of ClientConfig's fields. This client should
// be used to communicate with Elasticsearch.
func (c *Config) NewESClient() (*http.Client, error) {
	client := cleanhttp.DefaultClient()
	if c.Elasticsearch.Client == nil || !c.Elasticsearch.Client.TLSEnabled {
		return client, nil
	}

	if c.Elasticsearch.Client.CACert == "" {
		return nil, fmt.Errorf("no path to CA certificate")
	}
	if c.Elasticsearch.Client.ClientCert == "" {
		return nil, fmt.Errorf("no path to client certificate")
	}
	if c.Elasticsearch.Client.ClientKey == "" {
		return nil, fmt.Errorf("no path to client key")
	}

	// Load client certificate
	cert, err := tls.LoadX509KeyPair(c.Elasticsearch.Client.ClientCert, c.Elasticsearch.Client.ClientKey)
	if err != nil {
		return nil, fmt.Errorf("error loading X509 key pair: %v", err)
	}

	// Load CA certificate
	caCert, err := ioutil.ReadFile(c.Elasticsearch.Client.CACert)
	if err != nil {
		return nil, fmt.Errorf("error reading CA certificate file: %v", err)
	}
	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(caCert)

	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{cert},
		RootCAs:      caCertPool,
		ServerName:   c.Elasticsearch.Client.ServerName,
	}
	tlsConfig.BuildNameToCertificate()
	client.Transport.(*http.Transport).TLSClientConfig = tlsConfig
	return client, nil
}
