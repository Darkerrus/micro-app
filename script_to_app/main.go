package main

import (
	"fmt"
	"io/ioutil"
	"log"

	"github.com/nats-io/stan.go"
)

func main() {
	data, err := ioutil.ReadFile("model.json")

	if err != nil {
		fmt.Println(err)
	}
	sc, err := stan.Connect("test-cluster", "sb", stan.NatsURL("nats://localhost:14222"))
	if err != nil {
		log.Fatalf("can't connect to Nats: %v", err)
	}
	sc.Publish("foo", data)

}
