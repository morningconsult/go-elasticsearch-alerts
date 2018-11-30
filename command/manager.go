package command

import (
	"fmt"
	"net/http"
	"sync"

	"github.com/hashicorp/go-hclog"
	"github.com/morningconsult/go-elasticsearch-alerts/config"
	"github.com/morningconsult/go-elasticsearch-alerts/command/query"
	"github.com/morningconsult/go-elasticsearch-alerts/command/alert"
	"github.com/morningconsult/go-elasticsearch-alerts/command/alert/email"
	"github.com/morningconsult/go-elasticsearch-alerts/command/alert/file"
	"github.com/morningconsult/go-elasticsearch-alerts/command/alert/slack"
	"github.com/morningconsult/go-elasticsearch-alerts/command/query"
)

type queryManager struct {
	logger hclog.Logger
	client *http.Client
}

func newQueryManager(cfg *config.Config) (*queryManager, error) {
	client, err := cfg.NewESClient()
	if err != nil {
		return nil, fmt.Errorf("error creating new HTTP client: %v", err)
	}
}

func (qm *queryManager) buildQueryHandlers(cfg *config.Config) ([]*query.QueryHandler, error) {
	var queryHandlers []*query.QueryHandler
	for _, rule := range config.Rules {
		var methods []alert.AlertMethod
		for _, output := range rule.Outputs {
			method, err := qm.buildMethod(output)
			if err != nil {
				return nil, fmt.Errorf("error creating alert.AlertMethod: %v", err)
			}
			methods = append(methods, method)
		}
		handler, err := query.NewQueryHandler(&query.QueryHandlerConfig{
			Name:         rule.Name,
			Logger:       qm.logger,
			Distributed:  config.Distributed,
			AlertMethods: methods,
			Client:       qm.esClient,
			ESUrl:        config.ElasticSearch.Server.ElasticSearchURL,
			QueryData:    rule.ElasticSearchBody,
			QueryIndex:   rule.ElasticSearchIndex,
			Schedule:     rule.CronSchedule,
			Filters:      rule.Filters,
		})
		if err != nil {
			logger.Error("error creating new job handler", "error", err)
			return 1
		}
		queryHandlers = append(queryHandlers, handler)
	}
}

func (qm *queryManager) buildMethod(output *config.OutputConfig) (alert.AlertMethod, error) {
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