package command

import (
	"context"
	"errors"
	"sync"

	// "github.com/morningconsult/go-elasticsearch-alerts/config"
	"github.com/morningconsult/go-elasticsearch-alerts/utils/lock"
	"github.com/morningconsult/go-elasticsearch-alerts/command/query"
	"github.com/morningconsult/go-elasticsearch-alerts/command/alert"
)

type controllerConfig struct {
	alertHandler  *alert.AlertHandler
	queryHandlers []*query.QueryHandler
}

type controller struct {
	doneCh           chan struct{}
	outputCh         chan *alert.Alert
	updateHandlersCh chan []*query.QueryHandler
	distLock         *lock.Lock
	queryHandlerWG   *sync.WaitGroup
	alertHandler     *alert.AlertHandler
	queryHandlers    []*query.QueryHandler
}

func newController(config *controllerConfig) (*controller, error) {
	if config.alertHandler == nil {
		return nil, errors.New("no *alert.AlertHandler provided")
	}
	if len(config.queryHandlers) < 1 {
		return nil, errors.New("at least one *query.QueryHandler must be provided")
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
		close(qh.StopCh)
	}
	ctrl.queryHandlerWG.Wait()
}

func (ctrl *controller) stopAlertHandler() {
	close(ctrl.alertHandler.StopCh)
	<-ctrl.alertHandler.DoneCh
}
