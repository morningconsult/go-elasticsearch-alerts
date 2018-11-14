package main

import (
	"fmt"
	"log"
	// "time"

	"github.com/hashicorp/consul/api"
)

func main() {
	client, err := api.NewClient(&api.Config{
		Address: "http://127.0.0.1:8500",
	})
	if err != nil {
		log.Fatal(err)
	}

	lock, err := client.LockKey("go-elasticsearch-alerts/leader")
	if err != nil {
		log.Fatal(err)
	}

	stopCh := make(chan struct{})
	lockCh, err := lock.Lock(stopCh)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("acquired lock")
	select {
	case <-lockCh:
		fmt.Println("lost lock or an error")
	}
}