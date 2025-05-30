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
	"cmp"
	"os"
	"strings"

	consul "github.com/hashicorp/consul/api"

	"github.com/morningconsult/go-elasticsearch-alerts/config"
)

var consulEnvVars = []string{
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

func newConsulClient(config config.ConsulConfig) (*consul.Client, error) {
	for _, env := range consulEnvVars {
		if os.Getenv(strings.ToUpper(env)) != "" {
			continue
		}

		v := cmp.Or(config[env], config[strings.ToLower(env)])
		if v == "" {
			continue
		}

		reset := setEnvTemp(env, v)
		defer reset()
	}

	client, err := consul.NewClient(&consul.Config{})
	if err != nil {
		return nil, err
	}
	return client, nil
}

// setEnvTemp sets an environment variable and returns a function that restores
// the previous state.
func setEnvTemp(k, v string) (reset func()) {
	old, ok := os.LookupEnv(k)
	os.Setenv(k, v)

	if !ok {
		return func() { os.Unsetenv(k) }
	}

	return func() { os.Setenv(k, old) }
}
