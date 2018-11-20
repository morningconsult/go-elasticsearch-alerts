package command

import (
	"context"
	// "fmt"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	// "time"

	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/consul/api"
	"github.com/hashicorp/go-cleanhttp"
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
		logger.Error("error loading config file", "error", err)
		return 1
	}

	esClient, err := config.NewESClient()
	if err != nil {
		logger.Error("error creating new HTTP client", "error", err)
		return 1
	}
	apiClient := cleanhttp.DefaultClient()

	ah := alert.NewAlertHandler(&alert.AlertHandlerConfig{
		Logger: logger,
	})

	var queryHandlers []*query.QueryHandler
	for _, rule := range config.Rules {
		var methods []alert.AlertMethod
		for _, output := range rule.Outputs {
			var method alert.AlertMethod
			switch output.Type {
			case "slack":
				slackConfig := new(slack.SlackAlertMethodConfig)
				if err = mapstructure.Decode(output.Config, slackConfig); err != nil {
					logger.Error("error decoding Slack output configuration", "error", err)
					return 1
				}
				slackConfig.Client = apiClient

				method, err = slack.NewSlackAlertMethod(slackConfig)
				if err != nil {
					logger.Error("error creating new Slack output method", "error", err)
					return 1
				}
			case "file":
				fileConfig := new(file.FileAlertMethodConfig)
				if err = mapstructure.Decode(output.Config, fileConfig); err != nil {
					logger.Error("error decoding file output configuration", "error", err)
					return 1
				}

				method, err = file.NewFileAlertMethod(fileConfig)
				if err != nil {
					logger.Error("error creating new file output method", "error", err)
					return 1
				}
			default:
				logger.Error("output type is not valid", "'output.type'", output.Type)
				return 1
			}
			methods = append(methods, method)
		}
		handler, err := query.NewQueryHandler(&query.QueryHandlerConfig{
			Name:         rule.Name,
			Logger:       logger,
			Distributed:  config.Distributed,
			AlertMethods: methods,
			Client:       esClient,
			ESUrl:        config.ElasticSearch.Server.ElasticSearchURL,
			QueryData:    rule.ElasticSearchBody,
			QueryIndex:   rule.ElasticSearchIndex,
			Schedule:     rule.CronSchedule,
			StateIndex:   config.ElasticSearch.Server.ElasticSearchStateIndex,
			Filters:      rule.Filters,
		})
		if err != nil {
			logger.Error("error creating new job handler", "error", err)
			return 1
		}
		queryHandlers = append(queryHandlers, handler)
	}

	if config.Distributed {
		consulClient, err := newConsulClient(config.Consul)
		if err != nil {
			logger.Error("error creating Consul API client", "error", err)
			return 1
		}

		k, ok := config.Consul["consul_lock_key"]
		if !ok || k == "" {
			logger.Error("no 'consul_lock_key' value found")
			return 1
		}

		lock, err := consulClient.LockKey(k)
		if err != nil {
			logger.Error("error creating a Consul API lock", "error", err)
			return 1
		}

		wg.Add(1)

		go func(ctx context.Context) {
			defer func() {
				wg.Done()
			}()

			for {
				lockCh, err := lock.Lock(ctx.Done())
				if err != nil {
					logger.Error("error attempting to acquire lock, exiting", "error", err)
					close(shutdownCh)
					return
				}

			UnlockedLoop:
				for {
					select {
					case <-ctx.Done():
						lock.Unlock()
						return
					default:
						logger.Info("this process is now the leader")
						for _, handler := range queryHandlers {
							handler.HaveLockCh <- true
						}
					}

					select {
					case <-ctx.Done():
						lock.Unlock()
						return
					case <-lockCh:
						logger.Info("this process is no longer the leader")
						for _, handler := range queryHandlers {
							handler.HaveLockCh <- false
						}
						lock.Unlock()
						break UnlockedLoop
					}
				}
			}
		}(ctx)
	}

	outputCh := make(chan *alert.Alert, len(queryHandlers))

	wg.Add(len(queryHandlers) + 1)

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
		logger.Info("SIGKILL received. Cleaning up goroutines...")
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

func newConsulClient(config map[string]string) (*api.Client, error) {
	consulEnvVars := []string{
		api.HTTPAddrEnvName,
		api.HTTPTokenEnvName,
		api.HTTPSSLEnvName,
		api.HTTPCAFile,
		api.HTTPCAPath,
		api.HTTPClientCert,
		api.HTTPClientKey,
		api.HTTPTLSServerName,
		api.HTTPSSLVerifyEnvName,
	}

	for _, env := range consulEnvVars {
		v, ok := config[env]
		if !ok {
			v, ok = config[strings.ToLower(env)]
		}
		if ok && v != "" && os.Getenv(env) == "" {
			os.Setenv(env, v)
			defer os.Unsetenv(env)
		}
	}

	return api.NewClient(&api.Config{})
}