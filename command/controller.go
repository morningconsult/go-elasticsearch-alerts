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
	"sync"

	"golang.org/x/xerrors"

	"github.com/morningconsult/go-elasticsearch-alerts/command/alert"
	"github.com/morningconsult/go-elasticsearch-alerts/command/query"
	"github.com/morningconsult/go-elasticsearch-alerts/utils/lock"
)

type controllerConfig struct {
	alertHandler  *alert.Handler
	queryHandlers []*query.QueryHandler
}

type controller struct {
	doneCh           chan struct{}
	outputCh         chan *alert.Alert
	updateHandlersCh chan []*query.QueryHandler
	distLock         *lock.Lock
	queryHandlerWG   *sync.WaitGroup
	alertHandler     *alert.Handler
	queryHandlers    []*query.QueryHandler
}

func newController(config *controllerConfig) (*controller, error) {
	if config.alertHandler == nil {
		return nil, xerrors.New("no *alert.Handler provided")
	}
	if len(config.queryHandlers) < 1 {
		return nil, xerrors.New("at least one *query.QueryHandler must be provided")
	}

	return &controller{
		doneCh:           make(chan struct{}),
		outputCh:         make(chan *alert.Alert, 4),
		updateHandlersCh: make(chan []*query.QueryHandler, 1),
		distLock:         lock.NewLock(),
		queryHandlerWG:   new(sync.WaitGroup),
		alertHandler:     config.alertHandler,
		queryHandlers:    config.queryHandlers,
	}, nil
}

func (ctrl *controller) run(ctx context.Context) {
	ctrl.startAlertHandler(ctx)
	ctrl.startQueryHandlers(ctx)

	for {
		select {
		case <-ctx.Done():
			<-ctrl.alertHandler.DoneCh
			ctrl.queryHandlerWG.Wait()
			close(ctrl.doneCh)
			return
		case qhs := <-ctrl.updateHandlersCh:
			ctrl.stopQueryHandlers()
			ctrl.queryHandlers = qhs
			ctrl.startQueryHandlers(ctx)
		}
	}
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
		qh.StopCh <- struct{}{}
	}
	ctrl.queryHandlerWG.Wait()
}
