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
	ID       string
	RuleName string
	Methods  []AlertMethod
	Records  []*Record
}

type AlertMethod interface {
	Write(context.Context, string, []*Record) error
}

type AlertHandlerConfig struct {
	Logger hclog.Logger
}

type AlertHandler struct {
	logger hclog.Logger
	rand   *rand.Rand
}

func NewAlertHandler(config *AlertHandlerConfig) *AlertHandler {
	return &AlertHandler{
		logger: config.Logger,
		rand:   rand.New(rand.NewSource(int64(time.Now().Nanosecond()))),
	}
}

func (a *AlertHandler) Run(ctx context.Context, outputCh <-chan *Alert, wg *sync.WaitGroup) {
	defer func() {
		wg.Done()
	}()

	a.logger.Info("starting alert handler")

	alertCh := make(chan func() (int, error), 8)
	active := NewInventory()

	alertFunc := func(ctx context.Context, alertID, rule string, method AlertMethod, records []*Record) func() (int, error) {
		return func() (int, error) {
			if active.remaining(alertID) < 1 {
				active.deregister(alertID)
				return 0, nil
			}
			active.decrement(alertID)
			err := method.Write(ctx, rule, records)
			return active.remaining(alertID), err
		}
	}

	for {
		select {
		case <-ctx.Done():
			return
		case alert := <-outputCh:
			a.logger.Info("new query results received from rule", "\""+alert.RuleName+"\"")
			for i, method := range alert.Methods {
				alertMethodID := fmt.Sprintf("%d|%s", i, alert.ID)
				active.register(alertMethodID)
				alertCh <- alertFunc(ctx, alertMethodID, alert.RuleName, method, alert.Records)
			}
		case writeAlert := <-alertCh:
			select {
			case <-ctx.Done():
				return
			default:
			}

			n, err := writeAlert()
			if err != nil {
				backoff := a.newBackoff()
				a.logger.Error("error returned by alert function", "error", err, "remaining_retries", n, "backoff", backoff.String())
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
	return 2*time.Second + time.Duration(a.rand.Int63()%int64(time.Second*2)-int64(time.Second))
}
