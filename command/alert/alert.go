package alert

import (
	"context"
	"fmt"
	"math/rand"
	"sync"
	"time"

	"github.com/hashicorp/go-hclog"
)

type Field struct {
	Key   string `json:"key" mapstructure:"key`
	Count int    `json:"doc_count" mapstructure:"doc_count"`
}

type Record struct {
	Title  string   `json:"filter,omitempty"`
	Text   string   `json:"text,omitempty"`
	Fields []*Field `json:"fields,omitempty"`
}

type Alert struct {
	ID      string
	Methods []AlertMethod
	Records []*Record
}

type AlertMethod interface {
	Write(context.Context, []*Record) error
}

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

func (a *AlertHandler) Run(ctx context.Context, outputCh <-chan *Alert, wg *sync.WaitGroup) {
	defer func() {
		wg.Done()
	}()

	alertCh := make(chan func() error, 8)
	active := NewInventory()

	alertFunc := func(ctx context.Context, alertID string, method AlertMethod, records []*Record) func() error {
		return func() error {
			if active.remaining(alertID) < 1 {
				active.deregister(alertID)
				return nil
			}
			active.decrement(alertID)
			return method.Write(ctx, records)
		}
	}

	for {
		select {
		case <-ctx.Done():
			return
		case alert := <-outputCh:
			for i, method := range alert.Methods {
				alertMethodID := fmt.Sprintf("%d|%s", i, alert.ID)
				active.register(alertMethodID)
				alertCh <- alertFunc(ctx, alertMethodID, method, alert.Records)
			}
		case writeAlert := <-alertCh:
			select {
			case <-ctx.Done():
				return
			default:
			}

			if err := writeAlert(); err != nil {
				backoff := a.newBackoff()
				a.logger.Error("error returned by sink function, retrying", "error", err.Error(), "backoff", backoff.String())
				select {
				case <-ctx.Done():
					return
				case <-time.After(backoff):
					alertCh <- writeAlert
				}
			}
		}
	}
}

func (a *AlertHandler) newBackoff() time.Duration {
	random := rand.New(rand.NewSource(int64(time.Now().Nanosecond())))
	return 2*time.Second + time.Duration(random.Int63()%int64(time.Second*2)-int64(time.Second))
}
