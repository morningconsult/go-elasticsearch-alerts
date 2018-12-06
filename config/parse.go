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

// OutputConfig maps to each element of 'output' field of
// a rule configuration file.
type OutputConfig struct {
	// Type is the type of output method. Some examples include
	// 'email', 'file', and 'slack'. Additional output methods
	// may be added in the future
	Type string `json:"type"`

	// Config is used to configure the chosen type of output method.
	// The content of this field is specific to the output type.
	// Please refer to the README for more detailed information
	// on this field
	Config map[string]interface{} `json:"config"`
}

// RuleConfig represents a rule configuration file
type RuleConfig struct {
	// Name is the name of the rule. This value should come
	// from the 'name' field of the rule configuration file
	Name string `json:"name"`

	// ElasticsearchIndex is the index that this rule should
	// query. This value should come from the 'index' field
	// of the rule configuration file
	ElasticsearchIndex string `json:"index"`

	// CronSchedule is the interval at which the
	// *github.com/morningconsult/go-elasticsearch-alerts/command/query.QueryHandler
	// will execute the query. This value should come from
	// the 'schedule' field of the rule configuration file
	CronSchedule string `json:"schedule"`

	// BodyField is the field on which the application should
	// group query responses before sending alerts. This value
	// should come from the 'body_field' field of the rule
	// configuration file
	BodyField string `json:"body_field"`

	// ElasticsearchBodyRaw is the untyped query that this
	// alert should send when querying Elasticsearch. This
	// value should come from the 'body' field of the
	// rule configuration file
	ElasticsearchBodyRaw interface{} `json:"body"`

	// ElasticsearchBody is the typed query that this alert
	// will send when querying Elasticsearch
	ElasticsearchBody map[string]interface{} `json:"-"`

	// Filters are the additional fields on which the application
	// should group query responses before sending alerts. This
	// value should come from the 'filters' field of the rule
	// configuration file
	Filters []string `json:"filters"`

	// Outputs are the methods by which alerts should be sent
	Outputs []*OutputConfig `json:"outputs"`
}

// ServerConfig represents the 'elasticsearch.server'
// field of the main configuration file.
type ServerConfig struct {
	// ElasticsearchURL is the URL of your Elasticsearch instance.
	// This value should come from the 'elasticsearch.server.url'
	// field of the main configuration file
	ElasticsearchURL string `json:"url"`
}

// ESConfig represents the 'elasticsearch' field of the
// main configuration file.
type ESConfig struct {
	// Server represents the 'elasticsearch.server' field
	// of the main configuration file
	Server *ServerConfig `json:"server"`

	// Client represents the 'elasticsearch.client' field
	// of the main configuration file
	Client *ClientConfig `json:"client"`
}

// Config represents the main configuration file.
type Config struct {
	// Elasticsearch is the Elasticsearch client and server
	// configuration. This value should come from the
	// 'elasticsearch' field of the main configuration file
	Elasticsearch *ESConfig `json:"elasticsearch"`

	// Distributed is whether or not this process will be run
	// in a distributed fashion. This value should come from
	// the 'distributed' field of the main configuration file
	Distributed bool `json:"distributed"`

	// Consul is the configuration of your Consul server. This
	// field is required if the process shall be run in a
	// distributed fashion. This value should come from the
	// 'consul' field of the main configuration file
	Consul map[string]string `json:"consul"`

	// Rules are the methods by which any alerts should be sent.
	Rules []*RuleConfig `json:"-"`
}

// ParseConfig parses the main configuration file and returns a
// *Config instance or a non-nil error if there was an error.
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

	if cfg.Elasticsearch == nil {
		return nil, errors.New("no 'elasticsearch' field found in main configuration file")
	}

	if cfg.Elasticsearch.Server == nil {
		return nil, errors.New("no 'elasticsearch.server' field found in main configuration file")
	}

	if cfg.Elasticsearch.Server.ElasticsearchURL == "" {
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

	rules, err := ParseRules()
	if err != nil {
		return nil, err
	}

	if len(rules) < 1 {
		return nil, errors.New("at least one rule must be specified")
	}

	cfg.Rules = rules
	return cfg, nil
}

// ParseRules parses the rule configuration files and returns an
// array of *RuleConfig or a non-nil error if there was an error.
func ParseRules() ([]*RuleConfig, error) {
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

		switch b := rule.ElasticsearchBodyRaw.(type) {
		case map[string]interface{}:
			rule.ElasticsearchBody = b
		case string:
			var body map[string]interface{}
			if err = jsonutil.DecodeJSON([]byte(b), &body); err != nil {
				return nil, fmt.Errorf("error JSON-decoding 'body' field of file %s: %v", file.Name(), err)
			}
			rule.ElasticsearchBody = body
		default:
			return nil, fmt.Errorf("'body' field of file %s must be valid JSON", file.Name())
		}
		rule.ElasticsearchBodyRaw = nil

		if rule.Name == "" {
			return nil, fmt.Errorf("no 'name' field found in rule file %s", file.Name())
		}

		if rule.ElasticsearchIndex == "" {
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
