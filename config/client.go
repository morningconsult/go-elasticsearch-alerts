// Copyright 2018 The Morning Consult, LLC or its affiliates. All Rights Reserved.
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

type ClientConfig struct {
	TLSEnabled bool   `json:"tls_enabled"`
	CACert     string `json:"ca_cert"`
	ClientCert string `json:"client_cert"`
	ClientKey  string `json:"client_key"`
	ServerName string `json:"server_name"`
}

func (c *Config) NewESClient() (*http.Client, error) {
	client := cleanhttp.DefaultClient()
	if c.ElasticSearch.Client == nil || !c.ElasticSearch.Client.TLSEnabled {
		return client, nil
	}

	if c.ElasticSearch.Client.CACert == "" {
		return nil, fmt.Errorf("no path to CA certificate")
	}
	if c.ElasticSearch.Client.ClientCert == "" {
		return nil, fmt.Errorf("no path to client certificate")
	}
	if c.ElasticSearch.Client.ClientKey == "" {
		return nil, fmt.Errorf("no path to client key")
	}

	// Load client certificate
	cert, err := tls.LoadX509KeyPair(c.ElasticSearch.Client.ClientCert, c.ElasticSearch.Client.ClientKey)
	if err != nil {
		return nil, fmt.Errorf("error loading X509 key pair: %v", err)
	}

	// Load CA certificate
	caCert, err := ioutil.ReadFile(c.ElasticSearch.Client.CACert)
	if err != nil {
		return nil, fmt.Errorf("error reading CA certificate file: %v", err)
	}
	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(caCert)

	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{cert},
		RootCAs:      caCertPool,
		ServerName:   c.ElasticSearch.Client.ServerName,
	}
	tlsConfig.BuildNameToCertificate()
	client.Transport.(*http.Transport).TLSClientConfig = tlsConfig
	return client, nil
}