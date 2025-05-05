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
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	consul "github.com/hashicorp/consul/api"
	hclog "github.com/hashicorp/go-hclog"
	"golang.org/x/xerrors"

	"github.com/morningconsult/go-elasticsearch-alerts/command/alert"
	"github.com/morningconsult/go-elasticsearch-alerts/config"
)

// Run starts the daemon running. This function should be
// called directly within os.Exit() in your main.main()
// function.
func Run() int { //nolint:gocyclo,gocognit
	logger := hclog.Default()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	shutdownCh := makeShutdownCh()
	reloadCh := makeReloadCh(ctx)

	cfg, err := config.ParseConfig()
	if err != nil {
		logger.Error("Error loading main configuration file", "error", err)
		return 1
	}

	esClient, err := cfg.NewESClient()
	if err != nil {
		logger.Error("Error creating new Elasticsearch HTTP client", "error", err)
		return 1
	}

	qhs, err := buildQueryHandlers(cfg.Rules, cfg.Elasticsearch.Server.ElasticsearchURL, esClient, logger)
	if err != nil {
		logger.Error("Error creating query handlers from rules", "error", err)
		return 1
	}

	controller, err := newController(&controllerConfig{
		queryHandlers: qhs,
		alertHandler: alert.NewHandler(&alert.HandlerConfig{
			Logger: logger.Named("alert_handler"),
		}),
	})
	if err != nil {
		logger.Error("Error creating new controller", "error", err)
		return 1
	}

	syncDoneCh := make(chan struct{})
	syncErrCh := make(chan error)
	if cfg.Distributed {
		go handleDistOp(ctx, cfg.Consul, logger, controller, syncErrCh, syncDoneCh)
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
		close(syncErrCh)
		<-controller.doneCh
		<-reloadCh
	}()

	for {
		select {
		case err := <-syncErrCh:
			if err != nil {
				logger.Error("Error in distributed operation", "error", err)
				cancel()
				return 1
			}
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
			qhs, err := buildQueryHandlers(rules, cfg.Elasticsearch.Server.ElasticsearchURL, esClient, logger)
			if err != nil {
				logger.Error("Error creating query handlers from rules. Exiting", "error", err)
				cancel()
				return 1
			}
			controller.updateHandlersCh <- qhs
		}
	}
}

func handleDistOp(
	ctx context.Context,
	cfg config.ConsulConfig,
	logger hclog.Logger,
	ctrl *controller,
	errCh chan<- error,
	doneCh chan struct{},
) {
	defer close(doneCh)

	lock, err := newConsulLock(cfg)
	if err != nil {
		errCh <- err
		return
	}
	defer lock.Unlock()

	for {
		lockCh, err := lock.Lock(ctx.Done())
		if err != nil {
			errCh <- xerrors.Errorf("error attempting to acquire lock: %v", err)
			return
		}

		select {
		case <-ctx.Done():
			return
		default:
			logger.Info("This process is now the leader")
			ctrl.distLock.Set(true)
		}

	UnlockedLoop:
		for {
			select {
			case <-ctx.Done():
				return
			case <-lockCh:
				logger.Info("This process is no longer the leader")
				ctrl.distLock.Set(false)
				lock.Unlock()
				break UnlockedLoop
			}
		}
	}
}

func newConsulLock(cfg config.ConsulConfig) (*consul.Lock, error) {
	client, err := newConsulClient(cfg)
	if err != nil {
		return nil, xerrors.Errorf("error creating Consul API client: %v", err)
	}

	k, ok := cfg["consul_lock_key"]
	if !ok || k == "" {
		return nil, xerrors.Errorf("no 'consul_lock_key' value found in config")
	}

	lock, err := client.LockKey(k)
	if err != nil {
		return nil, xerrors.Errorf("error creating a Consul API lock: %v", err)
	}
	return lock, nil
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
