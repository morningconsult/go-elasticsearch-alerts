package utils

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/hashicorp/go-cleanhttp"
)

type ClientOptions struct {
	TLSEnabled bool
	CACert     string
	ClientCert string
	ClientKey  string
	ServerName string
}

func NewClient(opts *ClientOptions) (*http.Client, error) {
	client := cleanhttp.DefaultClient()
	if !opts.TLSEnabled {
		return client, nil
	}

	if opts.CACert == "" {
		return nil, fmt.Errorf("no path to CA certificate")
	}
	if opts.ClientCert == "" {
		return nil, fmt.Errorf("no path to client certificate")
	}
	if opts.ClientKey == "" {
		return nil, fmt.Errorf("no path to client key")
	}

	// Load client certificate
	cert, err := tls.LoadX509KeyPair(opts.ClientCert, opts.ClientKey)
	if err != nil {
		return nil, fmt.Errorf("error loading X509 key pair: %v", err)
	}

	// Load CA certificate
	caCert, err := ioutil.ReadFile(opts.CACert)
	if err != nil {
		return nil, fmt.Errorf("error reading CA certificate file: %v", err)
	}
	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(caCert)

	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{cert},
		RootCAs:      caCertPool,
		ServerName:   opts.ServerName,
	}
	tlsConfig.BuildNameToCertificate()
	client.Transport.(*http.Transport).TLSClientConfig = tlsConfig
	return client, nil
}