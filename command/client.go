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

package command

import (
	"os"
	"strings"

	consul "github.com/hashicorp/consul/api"
	"github.com/morningconsult/go-elasticsearch-alerts/config"
)

func newConsulClient(config config.ConsulConfig) (*consul.Client, error) {
	consulEnvVars := []string{
		consul.HTTPAddrEnvName,
		consul.HTTPTokenEnvName,
		consul.HTTPSSLEnvName,
		consul.HTTPCAFile,
		consul.HTTPCAPath,
		consul.HTTPClientCert,
		consul.HTTPClientKey,
		consul.HTTPTLSServerName,
		consul.HTTPSSLVerifyEnvName,
	}

	for _, env := range consulEnvVars {
		if os.Getenv(strings.ToUpper(env)) != "" {
			continue
		}

		v, ok := config[env]
		if !ok {
			v, ok = config[strings.ToLower(env)]
			if !ok {
				continue
			}
		}

		if v != "" {
			os.Setenv(env, v)
			defer os.Unsetenv(env)
		}
	}

	client, err := consul.NewClient(&consul.Config{})
	if err != nil {
		return nil, err
	}
	return client, nil
}
