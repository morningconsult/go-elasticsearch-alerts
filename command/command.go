package command

import (
	"context"
	// "fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
	// "time"

	"github.com/hashicorp/go-hclog"
	"github.com/mitchellh/mapstructure"
	"gitlab.morningconsult.com/mci/go-elasticsearch-alerts/config"
	"gitlab.morningconsult.com/mci/go-elasticsearch-alerts/command/query"
	"gitlab.morningconsult.com/mci/go-elasticsearch-alerts/command/alert"
	"gitlab.morningconsult.com/mci/go-elasticsearch-alerts/command/alert/slack"
	"gitlab.morningconsult.com/mci/go-elasticsearch-alerts/command/alert/file"
)

func Run() int {
	var wg sync.WaitGroup

	logger := hclog.Default()
	ctx, cancel := context.WithCancel(context.Background())

	shutdownCh := makeShutdownCh()

	config, err := config.ParseConfig()
	if err != nil {
		logger.Error("error loading config file", err.Error())
		return 1
	}

	client, err := config.NewClient()
	if err != nil {
		logger.Error("error creating new HTTP client", err.Error())
		return 1
	}
	ah := alert.NewAlertHandler(&alert.AlertHandlerConfig{
		Logger: logger,
	})

	var queryHandlers []*query.QueryHandler
	for _, rule := range config.Rules {

		// Build alert.AlertMethod array for this rule
		var methods []alert.AlertMethod
		for _, output := range rule.Outputs {
			var method alert.AlertMethod
			switch output.Type {
			case "slack":
				slackConfig := new(slack.SlackAlertMethodConfig)
				if err = mapstructure.Decode(output.Config, slackConfig); err != nil {
					logger.Error("error decoding Slack output configuration", err.Error())
					return 1
				}
				slackConfig.Client = client

				method, err = slack.NewSlackAlertMethod(slackConfig)
				if err != nil {
					logger.Error("error creating new Slack output method", err.Error())
					return 1
				}
			case "file":
				fileConfig := new(file.FileAlertMethodConfig)
				if err = mapstructure.Decode(output.Config, fileConfig); err != nil {
					logger.Error("error decoding file output configuration", err.Error())
					return 1
				}
				fileConfig.RuleName = rule.Name

				method, err = file.NewFileAlertMethod(fileConfig)
				if err != nil {
					logger.Error("error creating new file output method", err.Error())
					return 1
				}
			default:
				logger.Error("output type %q is not a valid type", output.Type)
				return 1
			}
			methods = append(methods, method)
		}
		handler, err := query.NewQueryHandler(&query.QueryHandlerConfig{
			Name:         rule.Name,
			Logger:       logger,
			AlertMethods: methods,
			Client:       client,
			ESUrl:        config.Server.ElasticSearchURL,
			QueryData:    rule.ElasticSearchBody,
			QueryIndex:   rule.ElasticSearchIndex,
			Schedule:     rule.CronSchedule,
			StateIndex:   config.Server.ElasticSearchStateIndex,
			Filters:      rule.Filters,
		})
		if err != nil {
			logger.Error("error creating new job handler", err.Error())
			return 1
		}
		queryHandlers = append(queryHandlers, handler)
	}

	wg.Add(len(queryHandlers) + 1)

	outputCh := make(chan *alert.Alert, len(queryHandlers))

	go ah.Run(ctx, outputCh, &wg)
	for _, qh := range queryHandlers {
		go qh.Run(ctx, outputCh, &wg)
	}

	go func() {
		wg.Wait()
		close(outputCh)
	}()

	select {
	case <-shutdownCh:
		logger.Info("SIGKILL received")
		cancel()
		// Wait for goroutines to cleanup
		<-outputCh
	}
	return 0
}

// makeShutdownCh returns a channel that can be used for shutdown
// notifications for commands. This channel will send a message for every
// SIGINT or SIGTERM received.
func makeShutdownCh() chan struct{} {
	resultCh := make(chan struct{})

	shutdownCh := make(chan os.Signal, 4)
	signal.Notify(shutdownCh, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-shutdownCh
		close(resultCh)
	}()
	return resultCh
}
