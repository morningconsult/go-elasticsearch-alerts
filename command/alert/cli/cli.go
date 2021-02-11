package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/morningconsult/go-elasticsearch-alerts/command/alert"
	"github.com/morningconsult/go-elasticsearch-alerts/utils"
	"github.com/shopspring/decimal"
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

	params := map[string]interface{}{}

	for _, item := range records {
		msg := e.args
		getParamsFromMsg(msg, params)

		if !item.BodyField && item.Elements != nil {
			for _, element := range item.Elements {
				for k, _ := range params {
					params[k] = utils.Get(element, k)

					switch  v := params[k].(type) {
					case json.Number:
						msg = strings.Replace(msg, "%"+ k +"%", requireFromString(string(v)), -1)
					case string:
						msg = strings.Replace(msg, "%"+ k +"%", v, -1)
					}
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
		fmt.Println(comand.arg)

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

func requireFromString(in string) string {
	d := decimal.RequireFromString(string(in))
	return d.String()
}

func getParamsFromMsg(msg string, out map[string]interface{}) {
	start := strings.Index(msg, "%") + 1
	end := strings.Index(msg[start:], "%") + start

	if start < 0 || end < 0 || start > end {
		return
	}

	out[msg[start:end]] = nil
	getParamsFromMsg(msg[end+1:], out)
}
