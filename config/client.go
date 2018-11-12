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

func (c *Config) NewClient() (*http.Client, error) {
	client := cleanhttp.DefaultClient()
	if c.Client == nil || !c.Client.TLSEnabled {
		return client, nil
	}

	if c.Client.CACert == "" {
		return nil, fmt.Errorf("no path to CA certificate")
	}
	if c.Client.ClientCert == "" {
		return nil, fmt.Errorf("no path to client certificate")
	}
	if c.Client.ClientKey == "" {
		return nil, fmt.Errorf("no path to client key")
	}

	// Load client certificate
	cert, err := tls.LoadX509KeyPair(c.Client.ClientCert, c.Client.ClientKey)
	if err != nil {
		return nil, fmt.Errorf("error loading X509 key pair: %v", err)
	}

	// Load CA certificate
	caCert, err := ioutil.ReadFile(c.Client.CACert)
	if err != nil {
		return nil, fmt.Errorf("error reading CA certificate file: %v", err)
	}
	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(caCert)

	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{cert},
		RootCAs:      caCertPool,
		ServerName:   c.Client.ServerName,
	}
	tlsConfig.BuildNameToCertificate()
	client.Transport.(*http.Transport).TLSClientConfig = tlsConfig
	return client, nil
}