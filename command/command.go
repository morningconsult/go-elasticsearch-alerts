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

package command

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/hashicorp/consul/api"
	"github.com/hashicorp/go-hclog"
	"github.com/morningconsult/go-elasticsearch-alerts/config"
	"github.com/morningconsult/go-elasticsearch-alerts/command/alert"
)

func Run() int {

	logger := hclog.Default()
	ctx, cancel := context.WithCancel(context.Background())

	shutdownCh := makeShutdownCh()
	reloadCh := makeReloadCh(ctx)

	cfg, err := config.ParseConfig()
	if err != nil {
		logger.Error("Error loading config file", "error", err)
		return 1
	}

	esClient, err := cfg.NewESClient()
	if err != nil {
		logger.Error("Error creating new ElasticSearch HTTP client", "error", err)
		return 1
	}

	qhs, err := buildQueryHandlers(cfg.Rules, cfg.ElasticSearch.Server.ElasticSearchURL, esClient, logger)
	if err != nil {
		logger.Error("Error creating query handlers from rules", "error", err)
		return 1
	}

	controller, err := newController(&controllerConfig{
		queryHandlers: qhs,
		alertHandler:  alert.NewAlertHandler(&alert.AlertHandlerConfig{
			Logger: logger,
		}),
	})
	if err != nil {
		logger.Error("Error creating new controller", "error", err)
		return 1
	}

	syncDoneCh := make(chan struct{})
	if cfg.Distributed {
		consulClient, err := newConsulClient(cfg.Consul)
		if err != nil {
			logger.Error("Error creating Consul API client", "error", err)
			return 1
		}

		k, ok := cfg.Consul["consul_lock_key"]
		if !ok || k == "" {
			logger.Error("No 'consul_lock_key' value found")
			return 1
		}

		lock, err := consulClient.LockKey(k)
		if err != nil {
			logger.Error("Error creating a Consul API lock", "error", err)
			return 1
		}

		go func(ctx context.Context) {
			defer func() {
				close(syncDoneCh)
			}()

			for {
				lockCh, err := lock.Lock(ctx.Done())
				if err != nil {
					logger.Error("Error attempting to acquire lock, exiting", "error", err)
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
						logger.Info("This process is now the leader")
						controller.distLock.Set(true)
					}

					select {
					case <-ctx.Done():
						lock.Unlock()
						return
					case <-lockCh:
						logger.Info("This process is no longer the leader")
						controller.distLock.Set(false)
						lock.Unlock()
						break UnlockedLoop
					}
				}
			}
		}(ctx)
	} else {
		close(syncDoneCh)
		controller.distLock.Set(true)
	}

	qh := controller.queryHandlers[0]
	err = qh.PutTemplate(ctx)
	if err != nil {
		logger.Error(fmt.Sprintf("Error creating template %q", qh.StateAliasURL()), "error", err)
	} else {
		logger.Info(fmt.Sprintf("Successfully created template %q", qh.StateAliasURL()))
	}

	go controller.run(ctx)

	defer func() {
		<-syncDoneCh
		<-controller.doneCh
		<-reloadCh
	}()

	for {
		select {
		case <-shutdownCh:
			logger.Info("SIGKILL received. Cleaning up goroutines...")
			cancel()
			return 0
		case <-reloadCh:
			logger.Info("SIGHUP received. Updating rules.")

			rules, err := config.ParseRules()
			if err != nil {
				logger.Error("Error parsing rules. Exiting", "error", err)
				cancel()
				return 1
			}
			qhs, err := buildQueryHandlers(rules, cfg.ElasticSearch.Server.ElasticSearchURL, esClient, logger)
			if err != nil {
				logger.Error("Error creating query handlers from rules. Exiting", "error", err)
				cancel()
				return 1
			}
			controller.updateHandlersCh <- qhs
		}
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

func makeReloadCh(ctx context.Context) chan struct{} {
	resultCh := make(chan struct{})

	reloadCh := make(chan os.Signal, 1)
	signal.Notify(reloadCh, syscall.SIGHUP)
	go func() {
		for {
			select {
			case <-ctx.Done():
				close(resultCh)
				return
			case <-reloadCh:
				resultCh <- struct{}{}
			}
		}
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
