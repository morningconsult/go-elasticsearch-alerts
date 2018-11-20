package config

import (
	"testing"
)

func TestNewESClient(t *testing.T) {
	cases := []struct{
		name   string
		config *Config
		err    bool
	}{
		{
			"tls-disabled",
			&Config{
				ElasticSearch: &ESConfig{
					Client: &ClientConfig{
						TLSEnabled: false,
					},
				},
			},
			false,
		},
		{
			"no-ca-cert",
			&Config{
				ElasticSearch: &ESConfig{
					Client: &ClientConfig{
						TLSEnabled: true,
					},
				},
			},
			true,
		},
		{
			"no-client-cert",
			&Config{
				ElasticSearch: &ESConfig{
					Client: &ClientConfig{
						TLSEnabled: true,
						CACert:     "testdata/certs/cacert.pem",
					},
				},
			},
			true,
		},
		{
			"no-client-key",
			&Config{
				ElasticSearch: &ESConfig{
					Client: &ClientConfig{
						TLSEnabled: true,
						CACert:     "testdata/certs/cacert.pem",
						ClientCert: "testdata/certs/cert.pem",
					},
				},
			},
			true,
		},
		{
			"error-loading-pair",
			&Config{
				ElasticSearch: &ESConfig{
					Client: &ClientConfig{
						TLSEnabled: true,
						CACert:     "testdata/certs/cacert.pem",
						ClientCert: "testdata/certs/i-dont-exist.pem",
						ClientKey:  "testdata/certs/key.pem",
					},
				},
			},
			true,
		},
		{
			"error-reading-ca-cert",
			&Config{
				ElasticSearch: &ESConfig{
					Client: &ClientConfig{
						TLSEnabled: true,
						CACert:     "testdata/certs/i-dont-exist.pem",
						ClientCert: "testdata/certs/cert.pem",
						ClientKey:  "testdata/certs/key.pem",
					},
				},
			},
			true,
		},
		{
			"success",
			&Config{
				ElasticSearch: &ESConfig{
					Client: &ClientConfig{
						TLSEnabled: true,
						CACert:     "testdata/certs/cacert.pem",
						ClientCert: "testdata/certs/cert.pem",
						ClientKey:  "testdata/certs/key.pem",
					},
				},
			},
			false,
		},
	}

	for _, tc := range cases {
		_, err := tc.config.NewESClient()
		t.Run(tc.name, func(t *testing.T) {
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