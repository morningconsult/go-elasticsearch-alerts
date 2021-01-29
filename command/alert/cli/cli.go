package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/morningconsult/go-elasticsearch-alerts/command/alert"
	"github.com/morningconsult/go-elasticsearch-alerts/utils"
	"os/exec"
	"strings"
	"time"
)

type comand struct {
	name, arg string
}

// AlertMethodConfig configures which command will be executed and with which parameters.
type AlertMethodConfig struct {
	Comand string `mapstructure:"comand"`
	// command arguments, separated by "|"
	Args string `mapstructure:"args"`
}

// AlertMethod implements the alert.AlertMethod interface
// for to execute cli commands
type AlertMethod struct {
	comand, args string
}

// NewAlertMethod creates a new *AlertMethod or a
// non-nil error if there was an error.
func NewAlertMethod(config *AlertMethodConfig) (alert.Method, error) {
	return &AlertMethod{
		comand: config.Comand,
		args:   config.Args,
	}, nil
}

func (e *AlertMethod) Write(ctx context.Context, rule string, records []*alert.Record) error {
	runctx, _ := context.WithTimeout(context.WithValue(ctx, "rule", rule), time.Minute*5)
	ch := make(chan comand)

	go e.run(runctx, ch)

	for _, item := range records {
		if !item.BodyField && item.Elements != nil {
			for _, element := range item.Elements {

				value := utils.Get(element, "timelock.value")
				doc_count := utils.Get(element, "doc_count")
				key := utils.Get(element, "key")

				msg := e.args
				if doc_count != nil {
					msg = strings.Replace(msg, "%count%", string(doc_count.(json.Number)), -1)
				}
				if value != nil {
					msg = strings.Replace(msg, "%value%", string(value.(json.Number)), -1)
				}
				if key != nil {
					msg = strings.Replace(msg, "%key%", key.(string), -1)
				}
				ch <- comand{
					name: e.comand,
					arg:  msg,
				}
			}
		}
	}
	close(ch)

	return nil
}

func (e *AlertMethod) run(ctx context.Context, c chan comand) {
	rule := ctx.Value("rule").(string)
	for comand := range c {
		cmd := exec.CommandContext(ctx, comand.name, strings.Split(comand.arg, "|")...)
		cmd.Stdout = new(bytes.Buffer)
		//cmd.Stderr = new(bytes.Buffer)

		if err := cmd.Run(); err != nil {
			fmt.Printf("Error executing the role %q:\n%v\n", rule, err)
		}
		fmt.Println(cmd.Stdout.(*bytes.Buffer).String())
		//fmt.Println(cmd.Stderr.(*bytes.Buffer).String())
	}
}
