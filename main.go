package main

import (
	"os"
	cmd "gitlab.morningconsult.com/mci/go-elasticsearch-alerts/command"
)

func main() {
	os.Exit(cmd.Run())
}