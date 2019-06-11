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

import "testing"

func TestNewESClient(t *testing.T) {
	cases := []struct {
		name   string
		config *Config
		err    bool
	}{
		{
			"tls-disabled",
			&Config{
				Elasticsearch: &ESConfig{
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
				Elasticsearch: &ESConfig{
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
				Elasticsearch: &ESConfig{
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
				Elasticsearch: &ESConfig{
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
				Elasticsearch: &ESConfig{
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
				Elasticsearch: &ESConfig{
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
				Elasticsearch: &ESConfig{
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
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			_, err := tc.config.NewESClient()
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
