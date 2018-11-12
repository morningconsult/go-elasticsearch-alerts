package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	// "context"

	"github.com/hashicorp/go-hclog"
	"gitlab.morningconsult.com/mci/go-elasticsearch-alerts/command/poll"
	"gitlab.morningconsult.com/mci/go-elasticsearch-alerts/command/alert/slack"
	"github.com/hashicorp/vault/helper/jsonutil"
)

func main() {
	file, err := os.Open("/home/dilan/Downloads/result.js")
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	var data map[string]interface{}
	if err := jsonutil.DecodeJSONFromReader(file, &data); err != nil {
		log.Fatal(err)
	}

	qh, err := poll.NewQueryHandler(&poll.QueryHandlerConfig{
		Logger: hclog.NewNullLogger(),
		Schedule: "* * * * * *",
		Filters: []string{"aggregations.hostname.buckets.program.buckets"},
	})
	if err != nil {
		log.Fatal(err)
	}
	results, err := qh.Transform(data)
	if err != nil {
		log.Fatal(err)
	}

	// ctx := context.Background()
	

	sh := slack.NewSlackAlertMethod(&slack.SlackAlertMethodConfig{
		Channel: "#ayylmao",
		Username: "dbellinghoven",
		Emoji: ":hankey:",
		Text: "hello!",
	})
	payload := sh.BuildPayload(results)

	buf, err := json.MarshalIndent(payload, "", "    ")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(string(buf))


	// for _, result := range results {
	// 	for _, field := range result.Fields {
	// 		fmt.Printf("%+v\n", field)
	// 	}
	// }

}