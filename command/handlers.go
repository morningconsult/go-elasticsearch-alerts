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
	"net/http"

	hclog "github.com/hashicorp/go-hclog"
	"github.com/mitchellh/mapstructure"
	"github.com/morningconsult/go-elasticsearch-alerts/command/alert"
	"github.com/morningconsult/go-elasticsearch-alerts/command/alert/email"
	"github.com/morningconsult/go-elasticsearch-alerts/command/alert/file"
	"github.com/morningconsult/go-elasticsearch-alerts/command/alert/slack"
	"github.com/morningconsult/go-elasticsearch-alerts/command/alert/sns"
	"github.com/morningconsult/go-elasticsearch-alerts/command/query"
	"github.com/morningconsult/go-elasticsearch-alerts/config"
	"golang.org/x/xerrors"
)

func buildQueryHandlers(
	rules []config.RuleConfig,
	esURL string,
	esClient *http.Client,
	logger hclog.Logger,
) ([]*query.QueryHandler, error) {
	if len(rules) < 1 {
		return nil, xerrors.New("at least one rule must be provided")
	}
	if logger == nil {
		return nil, xerrors.New("no logger provided")
	}
	if esClient == nil {
		return nil, xerrors.New("no HTTP client provided")
	}
	if esURL == "" {
		return nil, xerrors.New("no URL provided")
	}

	queryHandlers := make([]*query.QueryHandler, 0, len(rules))
	for _, rule := range rules {
		var methods []alert.Method
		for _, output := range rule.Outputs {
			method, err := buildMethod(output)
			if err != nil {
				return nil, xerrors.Errorf("error creating alert.AlertMethod: %v", err)
			}
			methods = append(methods, method)
		}
		handler, err := query.NewQueryHandler(&query.QueryHandlerConfig{
			Name:         rule.Name,
			Logger:       logger,
			AlertMethods: methods,
			Client:       esClient,
			ESUrl:        esURL,
			QueryData:    rule.ElasticsearchBody,
			QueryIndex:   rule.ElasticsearchIndex,
			Schedule:     rule.CronSchedule,
			BodyField:    rule.BodyField,
			Filters:      rule.Filters,
			Conditions:   rule.Conditions,
		})
		if err != nil {
			return nil, xerrors.Errorf("error creating new *query.QueryHandler: %v", err)
		}
		queryHandlers = append(queryHandlers, handler)
	}
	return queryHandlers, nil
}

func buildMethod(output config.OutputConfig) (alert.Method, error) {
	var method alert.Method
	var err error

	switch output.Type {
	case "slack":
		slackConfig := new(slack.AlertMethodConfig)
		if err = mapstructure.Decode(output.Config, slackConfig); err != nil {
			return nil, xerrors.Errorf("error decoding Slack output configuration: %v", err)
		}
		method, err = slack.NewAlertMethod(slackConfig)
	case "file":
		fileConfig := new(file.AlertMethodConfig)
		if err = mapstructure.Decode(output.Config, fileConfig); err != nil {
			return nil, xerrors.Errorf("error decoding file output configuration: %v", err)
		}
		method, err = file.NewAlertMethod(fileConfig)
	case "email":
		emailConfig := new(email.AlertMethodConfig)
		if err = mapstructure.Decode(output.Config, emailConfig); err != nil {
			return nil, xerrors.Errorf("error decoding email output configuration: %v", err)
		}
		method, err = email.NewAlertMethod(emailConfig)
	case "sns":
		snsConfig := new(sns.AlertMethodConfig)
		if err = mapstructure.Decode(output.Config, snsConfig); err != nil {
			return nil, xerrors.Errorf("error decoding SNS output configuration: %v", err)
		}
		method, err = sns.NewAlertMethod(snsConfig)
	default:
		return nil, xerrors.Errorf("output type %q is not supported", output.Type)
	}
	if err != nil {
		return nil, xerrors.Errorf("error creating new %s output method: %v", output.Type, err)
	}
	return method, nil
}
