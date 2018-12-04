package command

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/hashicorp/go-hclog"
	"github.com/mitchellh/mapstructure"
	"github.com/morningconsult/go-elasticsearch-alerts/config"
	"github.com/morningconsult/go-elasticsearch-alerts/command/query"
	"github.com/morningconsult/go-elasticsearch-alerts/command/alert"
	"github.com/morningconsult/go-elasticsearch-alerts/command/alert/email"
	"github.com/morningconsult/go-elasticsearch-alerts/command/alert/file"
	"github.com/morningconsult/go-elasticsearch-alerts/command/alert/slack"
)

func buildQueryHandlers(rules []*config.RuleConfig, esURL string, esClient *http.Client, logger hclog.Logger) ([]*query.QueryHandler, error) {
	if len(rules) < 1 {
		return nil, errors.New("at least one rule must be provided")
	}
	if logger == nil {
		return nil, errors.New("no logger provided")
	}
	if esClient == nil {
		return nil, errors.New("no HTTP client provided")
	}
	if esURL == "" {
		return nil, errors.New("no URL provided")
	}

	var queryHandlers []*query.QueryHandler
	for _, rule := range rules {
		var methods []alert.AlertMethod
		for _, output := range rule.Outputs {
			method, err := buildMethod(output)
			if err != nil {
				return nil, fmt.Errorf("error creating alert.AlertMethod: %v", err)
			}
			methods = append(methods, method)
		}
		handler, err := query.NewQueryHandler(&query.QueryHandlerConfig{
			Name:         rule.Name,
			Logger:       logger,
			AlertMethods: methods,
			Client:       esClient,
			ESUrl:        esURL,
			QueryData:    rule.ElasticSearchBody,
			QueryIndex:   rule.ElasticSearchIndex,
			Schedule:     rule.CronSchedule,
			BodyField:    rule.BodyField,
			Filters:      rule.Filters,
		})
		if err != nil {
			return nil, fmt.Errorf("error creating new job handler: %v", err)
		}
		queryHandlers = append(queryHandlers, handler)
	}
	return queryHandlers, nil
}

func buildMethod(output *config.OutputConfig) (alert.AlertMethod, error) {
	var method alert.AlertMethod
	var err error

	switch output.Type {
	case "slack":
		slackConfig := new(slack.SlackAlertMethodConfig)
		if err = mapstructure.Decode(output.Config, slackConfig); err != nil {
			return nil, fmt.Errorf("error decoding Slack output configuration: %v", err)
		}

		method, err = slack.NewSlackAlertMethod(slackConfig)
		if err != nil {
			return nil, fmt.Errorf("error creating new Slack output method: %v", err)
		}
	case "file":
		fileConfig := new(file.FileAlertMethodConfig)
		if err = mapstructure.Decode(output.Config, fileConfig); err != nil {
			return nil, fmt.Errorf("error decoding file output configuration: %v", err)
		}

		method, err = file.NewFileAlertMethod(fileConfig)
		if err != nil {
			return nil, fmt.Errorf("error creating new file output method: %v", err)
		}
	case "email":
		emailConfig := new(email.EmailAlertMethodConfig)
		if err = mapstructure.Decode(output.Config, emailConfig); err != nil {
			return nil, fmt.Errorf("error decoding email output configuration: %v", err)
		}

		method, err = email.NewEmailAlertMethod(emailConfig)
		if err != nil {
			return nil, fmt.Errorf("error creating new email output method: %v", err)
		}
	default:
		return nil, fmt.Errorf("output type %q is not supported", output.Type)
	}
	return method, nil
}