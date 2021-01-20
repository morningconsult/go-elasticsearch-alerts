package cli

import (
	"bytes"
	"context"
	"fmt"
	"github.com/morningconsult/go-elasticsearch-alerts/command/alert"
	"os/exec"
	"strconv"
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
		if !item.BodyField {

			for _, field := range item.Fields {
				msg := strings.Replace(e.args, "%key%", field.Key, -1)
				msg = strings.Replace(msg, "%count%", strconv.Itoa(field.Count), -1)
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
