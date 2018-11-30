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
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/hashicorp/vault/helper/jsonutil"
	"github.com/mitchellh/go-homedir"
)

const (
	envConfigFile     string = "GO_ELASTICSEARCH_ALERTS_CONFIG_FILE"
	envRulesDir       string = "GO_ELASTICSEARCH_ALERTS_RULES_DIR"
	defaultConfigFile string = "/etc/go-elasticsearch-alerts/config.json"
	defaultRulesDir   string = "/etc/go-elasticsearch-alerts/rules"
)

type OutputConfig struct {
	Type   string                 `json:"type"`
	Config map[string]interface{} `json:"config"`
}

type RuleConfig struct {
	Name                 string                 `json:"name"`
	ElasticSearchIndex   string                 `json:"index"`
	CronSchedule         string                 `json:"schedule"`
	ElasticSearchBodyRaw interface{}            `json:"body"`
	ElasticSearchBody    map[string]interface{} `json:"-"`
	Filters              []string               `json:"filters"`
	Outputs              []*OutputConfig        `json:"outputs"`
}

type DistributedConfig struct {
	ConsulAddr    string `json:"consul_address"`
	ConsulLockKey string `json:"consul_lock_key"`
}

type ServerConfig struct {
	ElasticSearchURL        string `json:"url"`
}

type ESConfig struct {
	Server *ServerConfig `json:"server"`
	Client *ClientConfig `json:"client"`
}

type Config struct {
	ElasticSearch *ESConfig         `json:"elasticsearch"`
	Distributed   bool              `json:"distributed"`
	Consul        map[string]string `json:"consul"`
	Rules         []*RuleConfig     `json:"-"`
}

func ParseConfig() (*Config, error) {
	configFile := defaultConfigFile
	if v := os.Getenv(envConfigFile); v != "" {
		d, err := homedir.Expand(v)
		if err != nil {
			return nil, err
		}
		configFile = d
	}

	file, err := os.Open(configFile)
	if err != nil {
		return nil, err
	}

	cfg := new(Config)
	if err := jsonutil.DecodeJSONFromReader(file, cfg); err != nil {
		file.Close()
		return nil, err
	}
	file.Close()

	if cfg.ElasticSearch == nil {
		return nil, errors.New("no 'elasticsearch' field found in main configuration file")
	}

	if cfg.ElasticSearch.Server == nil {
		return nil, errors.New("no 'elasticsearch.server' field found in main configuration file")
	}

	if cfg.ElasticSearch.Server.ElasticSearchURL == "" {
		return nil, errors.New("field 'elasticsearch.server.url' of main configuration file is empty")
	}

	if cfg.Distributed {
		if cfg.Consul == nil || len(cfg.Consul) < 1 {
			return nil, errors.New("field 'consul' of main configuration file is empty (required when 'distributed' is true)")
		}

		if _, ok := cfg.Consul["consul_http_addr"]; !ok {
			return nil, errors.New("field 'consul.consul_http_addr' of main configuration file is empty (required when 'distributed' is true)")
		}

		if _, ok := cfg.Consul["consul_lock_key"]; !ok {
			return nil, errors.New("field 'consul.consul_lock_key' of main configuration file is empty (required when 'distributed' is true)")
		}
	}

	rules, err := parseRules()
	if err != nil {
		return nil, err
	}

	cfg.Rules = rules
	return cfg, nil
}

func parseRules() ([]*RuleConfig, error) {
	rulesDir := defaultRulesDir
	if v := os.Getenv(envRulesDir); v != "" {
		d, err := homedir.Expand(v)
		if err != nil {
			return nil, fmt.Errorf("error expanding rules directory: %v", err)
		}
		rulesDir = d
	}

	ruleFiles, err := filepath.Glob(filepath.Join(rulesDir, "*.json"))
	if err != nil {
		return nil, err
	}

	var rules []*RuleConfig
	for _, ruleFile := range ruleFiles {
		file, err := os.Open(ruleFile)
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return nil, err
		}

		rule := new(RuleConfig)
		if err = jsonutil.DecodeJSONFromReader(file, rule); err != nil {
			file.Close()
			return nil, err
		}
		file.Close()

		switch b := rule.ElasticSearchBodyRaw.(type) {
		case map[string]interface{}:
			rule.ElasticSearchBody = b
		case string:
			var body map[string]interface{}
			if err = jsonutil.DecodeJSON([]byte(b), &body); err != nil {
				return nil, fmt.Errorf("error JSON-decoding 'body' field of file %s: %v", file.Name(), err)
			}
			rule.ElasticSearchBody = body
		default:
			return nil, fmt.Errorf("'body' field of file %s must be valid JSON", file.Name())
		}
		rule.ElasticSearchBodyRaw = nil

		if rule.Name == "" {
			return nil, fmt.Errorf("no 'name' field found in rule file %s", file.Name())
		}

		if rule.ElasticSearchIndex == "" {
			return nil, fmt.Errorf("no 'index' field found in rule file %s", file.Name())
		}

		if rule.CronSchedule == "" {
			return nil, fmt.Errorf("no 'schedule' field found in rule file %s", file.Name())
		}

		if rule.Filters == nil {
			rule.Filters = []string{}
		}

		if rule.Outputs == nil {
			return nil, fmt.Errorf("no 'output' field found in rule file %s", file.Name())
		}

		if len(rule.Outputs) < 1 {
			return nil, fmt.Errorf("at least one output must be specified ('outputs') in file %s", file.Name())
		}

		for _, output := range rule.Outputs {
			if output.Type == "" {
				return nil, fmt.Errorf("all outputs must have a type specified ('output.type') in file %s", file.Name())
			}

			if output.Config == nil || len(output.Config) < 1 {
				return nil, fmt.Errorf("all outputs must have a config field ('output.config') in file %s", file.Name())
			}
		}
		rules = append(rules, rule)
	}

	return rules, nil
}
