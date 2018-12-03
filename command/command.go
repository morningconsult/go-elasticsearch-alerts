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
)

func Run() int {

	logger := hclog.Default()
	ctx, cancel := context.WithCancel(context.Background())

	shutdownCh := makeShutdownCh()
	reloadCh := makeReloadCh(ctx)

	config, err := config.ParseConfig()
	if err != nil {
		logger.Error("error loading config file", "error", err)
		return 1
	}

	esClient, err := config.NewESClient()
	if err != nil {
		logger.Error("error creating new ElasticSearch HTTP client", "error", err)
		return 1
	}

	controller, err := newController(&controllerConfig{
		logger:              logger,
		rules:               config.Rules,
		elasticSearchURL:    config.ElasticSearch.Server.ElasticSearchURL,
		elasticSearchClient: esClient,
	})
	if err != nil {
		logger.Error("error creating new controller", "error", err)
		return 1
	}

	syncDoneCh := make(chan struct{})
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

		go func(ctx context.Context) {
			defer func() {
				close(syncDoneCh)
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
						controller.distLock.Set(true)
					}

					select {
					case <-ctx.Done():
						lock.Unlock()
						return
					case <-lockCh:
						logger.Info("this process is no longer the leader")
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
		logger.Error(fmt.Sprintf("error creating template %q", qh.StateAliasURL()), "error", err)
	} else {
		logger.Info(fmt.Sprintf("successfully created template %q", qh.StateAliasURL()))
	}

	controller.run(ctx)

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
			if err = controller.reload(ctx); err != nil {
				logger.Error("error updating handlers. Exiting", "error", err)
				cancel()
				return 1
			}
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
