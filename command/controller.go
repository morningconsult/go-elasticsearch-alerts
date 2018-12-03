package command

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"sync"

	"github.com/mitchellh/mapstructure"
	"github.com/hashicorp/go-cleanhttp"
	"github.com/hashicorp/go-hclog"
	"github.com/morningconsult/go-elasticsearch-alerts/config"
	"github.com/morningconsult/go-elasticsearch-alerts/utils/lock"
	"github.com/morningconsult/go-elasticsearch-alerts/command/query"
	"github.com/morningconsult/go-elasticsearch-alerts/command/alert"
	"github.com/morningconsult/go-elasticsearch-alerts/command/alert/email"
	"github.com/morningconsult/go-elasticsearch-alerts/command/alert/file"
	"github.com/morningconsult/go-elasticsearch-alerts/command/alert/slack"
	// "github.com/morningconsult/go-elasticsearch-alerts/command/query"
)

const defaultElasticSearchURL = "http://127.0.0.1:9200"

type controllerConfig struct {
	logger              hclog.Logger
	rules               []*config.RuleConfig
	elasticSearchURL    string
	elasticSearchClient *http.Client
}

type controller struct {
	doneCh         chan struct{}
	stopCh         chan struct{}
	outputCh       chan *alert.Alert
	logger         hclog.Logger
	distLock       *lock.Lock
	esClient       *http.Client
	esURL          string
	alertHandler   *alert.AlertHandler
	queryHandlerWG *sync.WaitGroup
	queryHandlers  []*query.QueryHandler
}

func newController(config *controllerConfig) (*controller, error) {
	if config == nil {
		return nil, errors.New("no *controllerConfig provided")
	}

	if len(config.rules) < 1 {
		return nil, errors.New("at least one *config.RuleConfig must be provided")
	}

	if config.logger == nil {
		config.logger = hclog.Default()
	}

	if config.elasticSearchClient == nil {
		config.elasticSearchClient = cleanhttp.DefaultClient()
	}

	if config.elasticSearchURL == "" {
		config.elasticSearchURL = defaultElasticSearchURL
	}

	handlers, err := buildQueryHandlers(config.rules, config.elasticSearchURL, config.elasticSearchClient, config.logger)
	if err != nil {
		return nil, fmt.Errorf("error creating query handlers: %v", err)
	}

	return &controller{
		doneCh:         make(chan struct{}),
		outputCh:       make(chan *alert.Alert, 4),
		logger:         config.logger,
		distLock:       lock.NewLock(),
		esClient:       config.elasticSearchClient,
		esURL:          config.elasticSearchURL,
		alertHandler:   alert.NewAlertHandler(&alert.AlertHandlerConfig{
			Logger: config.logger,
		}),
		queryHandlerWG: new(sync.WaitGroup),
		queryHandlers:  handlers,
	}, nil
}

func (ctrl *controller) run(ctx context.Context) {
	ctrl.startAlertHandler(ctx)
	ctrl.startQueryHandlers(ctx)

	go func() {
		<-ctx.Done()
		<-ctrl.alertHandler.DoneCh
		ctrl.queryHandlerWG.Wait()
		close(ctrl.doneCh)
	}()
}

func (ctrl *controller) startAlertHandler(ctx context.Context) {
	go ctrl.alertHandler.Run(ctx, ctrl.outputCh)
}

func (ctrl *controller) startQueryHandlers(ctx context.Context) {
	ctrl.queryHandlerWG.Add(len(ctrl.queryHandlers))
	for _, qh := range ctrl.queryHandlers {
		go qh.Run(ctx, ctrl.outputCh, ctrl.queryHandlerWG, ctrl.distLock)
	}
}

func (ctrl *controller) stopQueryHandlers() {
	for _, qh := range ctrl.queryHandlers {
		close(qh.StopCh)
	}
	ctrl.queryHandlerWG.Wait()
}

func (ctrl *controller) stopAlertHandler() {
	close(ctrl.alertHandler.StopCh)
	<-ctrl.alertHandler.DoneCh
}

func (ctrl *controller) reload(ctx context.Context) error {
	select {
	case <-ctrl.doneCh:
		return errors.New("doneCh has already been closed")
	default:
	}

	ctrl.stopQueryHandlers()

	rules, err := config.ParseRules()
	if err != nil {
		return fmt.Errorf("error parsing rules: %v", err)
	}

	handlers, err := buildQueryHandlers(rules, ctrl.esURL, ctrl.esClient, ctrl.logger)
	if err != nil {
		return fmt.Errorf("error creating query handlers: %v", err)
	}
	ctrl.queryHandlers = handlers
	ctrl.startQueryHandlers(ctx)
	return nil
}

func buildQueryHandlers(rules []*config.RuleConfig, esURL string, esClient *http.Client, logger hclog.Logger) ([]*query.QueryHandler, error) {
	if len(rules) < 1 {
		return nil, errors.New("at least one rule must be provided")
	}
	if logger == nil {
		return nil, errors.New("no logger provided")
	}
	if esClient == nil {
		return nil, errors.New("no ElasticSearch HTTP client provided")
	}
	if esURL == "" {
		return nil, errors.New("no ElasticSearch URL provided")
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