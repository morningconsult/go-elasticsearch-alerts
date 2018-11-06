package alert

import (
	"context"
	"fmt"
	"sync"

	"github.com/hashicorp/go-hclog"
)

type AlertHandlerConfig struct {
	Logger hclog.Logger
}

type AlertHandler struct {
	logger hclog.Logger
}

func NewAlertHandler(config *AlertHandlerConfig) *AlertHandler {
	return &AlertHandler{
		logger: config.Logger,
	}
}

func (a *AlertHandler) Run(ctx context.Context, outputCh <-chan interface{}, wg *sync.WaitGroup) {
	defer func() {
		wg.Done()
	}()

	for {
		select {
		case <-ctx.Done():
			return
		case msg := <-outputCh:
			m, ok := msg.(string)
			if !ok {
				continue
			}
			fmt.Printf("Received message: %q\n", m)
		}
	}
}