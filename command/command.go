package command

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
	// "time"

	"github.com/hashicorp/go-hclog"
	// "github.com/hashicorp/go-cleanhttp"

	"gitlab.morningconsult.com/mci/go-elasticsearch-alerts/command/poll"
	"gitlab.morningconsult.com/mci/go-elasticsearch-alerts/command/alert"
)

func Run() int {
	var wg sync.WaitGroup

	logger := hclog.Default()
	ctx, cancel := context.WithCancel(context.Background())

	shutdownCh := makeShutdownCh()

	client := cleanhttp.DefaultClient()
	// set up TLS?

	// config, err := config.LoadConfig(logger)
	// if err != nil {
	// 	logger.Error("error loading config file", err.Error())
	// 	return 1
	// }

	ah := alert.NewAlertHandler(&alert.AlertHandlerConfig{logger})

	var queryHandlers []*poll.QueryHandler
	// for _, a := range config.Alerts {
	// 	handler, err := poll.NewQueryHandler(&poll.QueryHandlerConfig{})
	// 	if err != nil {
	// 		logger.Error("error creating new job handler", err.Error())
	// 	}
	// 	queryHandlers = append(queryHandlers, handler)
	// }

	for _, a := range []int{10, 12, 14, 16, 18} {
		handler, err := poll.NewQueryHandler(&poll.QueryHandlerConfig{
			Name:     fmt.Sprintf("%d seconds", a),
			Schedule: fmt.Sprintf("*/%d * * * * *", a),
		})
		if err != nil {
			logger.Error("error creating new query handler", err.Error())
		}
		queryHandlers = append(queryHandlers, handler)
	}

	wg.Add(len(queryHandlers) + 1)

	outputCh := make(chan interface{}, len(queryHandlers))

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
		fmt.Println("SIGKILL received")
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
