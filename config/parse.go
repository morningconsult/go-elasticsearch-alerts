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
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	homedir "github.com/mitchellh/go-homedir"
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

func (o *OutputConfig) validate() error {
	if o.Type == "" {
		return errors.New("all outputs must have a type specified ('output.type')")
	}
	if o.Config == nil || len(o.Config) < 1 {
		return errors.New("all outputs must have a config field ('output.config')")
	}
	return nil
}

// ConsulConfig is used to configure the behavior of the
// Consul network lock required for distributed operation.
type ConsulConfig map[string]string

func (cc ConsulConfig) validate() error {
	if len(cc) < 1 {
		return errors.New("field 'consul' is empty (required when 'distributed' is true)")
	}
	if _, ok := cc["consul_http_addr"]; !ok {
		return errors.New("field 'consul.consul_http_addr' is empty (required when 'distributed' is true)")
	}
	if _, ok := cc["consul_lock_key"]; !ok {
		return errors.New("field 'consul.consul_lock_key' is empty (required when 'distributed' is true)")
	}
	return nil
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

func (rule *RuleConfig) validate() error {
	if rule.Name == "" {
		return errors.New("no 'name' field found")
	}
	if rule.ElasticsearchIndex == "" {
		return errors.New("no 'index' field found")
	}
	if rule.CronSchedule == "" {
		return errors.New("no 'schedule' field found")
	}
	if rule.Filters == nil {
		rule.Filters = []string{}
	}
	if rule.Outputs == nil {
		return errors.New("no 'output' field found")
	}
	if len(rule.Outputs) < 1 {
		return errors.New("at least one output must be specified ('outputs')")
	}
	for i, output := range rule.Outputs {
		if err := output.validate(); err != nil {
			return fmt.Errorf("error in output %d of rule %s: %v", i+1, rule.Name, err)
		}
	}
	return nil
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

func (es *ESConfig) validate() error {
	if es.Server == nil {
		return errors.New("no 'elasticsearch.server' field found")
	}
	if es.Server.ElasticsearchURL == "" {
		return errors.New("no 'elasticsearch.server.url' field found")
	}
	return nil
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
	Consul ConsulConfig `json:"consul"`

	// Rules are the definitions of the alerts
	Rules []*RuleConfig `json:"-"`
}

func decodeConfigFile(f string) (*Config, error) {
	var err error
	f, err = homedir.Expand(f)
	if err != nil {
		return nil, err
	}
	file, err := os.Open(filepath.Clean(f))
	if err != nil {
		return nil, err
	}
	defer file.Close()

	cfg := new(Config)
	err = json.NewDecoder(file).Decode(cfg)
	return cfg, err
}

// ParseConfig parses the main configuration file and returns a
// *Config instance or a non-nil error if there was an error.
func ParseConfig() (*Config, error) {
	configFile := defaultConfigFile
	if v := os.Getenv(envConfigFile); v != "" {
		configFile = v
	}

	cfg, err := decodeConfigFile(configFile)
	if err != nil {
		return nil, err
	}

	if cfg.Elasticsearch == nil {
		return nil, fmt.Errorf("no 'elasticsearch' field found in main configuration file %s", configFile)
	}
	if err = cfg.Elasticsearch.validate(); err != nil {
		return nil, fmt.Errorf("error in main configuration file %s: %v", configFile, err)
	}
	if cfg.Distributed {
		if cfg.Consul == nil {
			return nil, fmt.Errorf("no field 'consul' found in main configuration file %s (required when 'distributed' is true)", configFile) // nolint: lll
		}
		if err = cfg.Consul.validate(); err != nil {
			return nil, fmt.Errorf("error in main configuration file %s: %v", configFile, err)
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
func ParseRules() ([]*RuleConfig, error) { // nolint: gocyclo
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
		return nil, fmt.Errorf("error globbing rules dir: %v", err)
	}

	rules := make([]*RuleConfig, 0, len(ruleFiles))
	for _, ruleFile := range ruleFiles {
		file, err := os.Open(filepath.Clean(ruleFile))
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return nil, fmt.Errorf("error opening file %s: %v", file.Name(), err)
		}

		rule := new(RuleConfig)
		if err = json.NewDecoder(file).Decode(rule); err != nil {
			file.Close()
			return nil, fmt.Errorf("error JSON-decoding rule file %s: %v", file.Name(), err)
		}
		file.Close()

		rule.ElasticsearchBody, err = parseBody(rule.ElasticsearchBodyRaw)
		if err != nil {
			return nil, fmt.Errorf("error in rule file %s: %v", file.Name(), err)
		}
		rule.ElasticsearchBodyRaw = nil

		if err := rule.validate(); err != nil {
			return nil, fmt.Errorf("error in rule file %s: %v", file.Name(), err)
		}

		rules = append(rules, rule)
	}
	return rules, nil
}

func parseBody(v interface{}) (map[string]interface{}, error) {
	switch b := v.(type) {
	case map[string]interface{}:
		return b, nil
	case string:
		var body map[string]interface{}
		if err := json.NewDecoder(bytes.NewBufferString(b)).Decode(&body); err != nil {
			return nil, fmt.Errorf("error JSON-decoding 'body' field: %v", err)
		}
		return body, nil
	default:
		return nil, fmt.Errorf("'body' field must be valid JSON")
	}
}
