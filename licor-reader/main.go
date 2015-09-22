package main

import (
	"encoding/json"
	"github.com/kf8a/li820"
	zmq "github.com/pebbe/zmq4"
	"github.com/prometheus/client_golang/prometheus"
	"log"
	"net/http"
)

func readLicor() {
	licor := li820.NewLicor("li820", "glbrc", "/dev/ttyS1")

	socket, err := zmq.NewSocket(zmq.PUB)
	if err != nil {
		log.Fatal(err)
	}
	defer socket.Close()
	socket.Bind("tcp://*:5556")
	socket.Bind("ipc://weather.ipc")

	c := make(chan li820.Datum)
	go licor.Sampler(c)
	for {
		sample := <-c
		jsonString, err := json.Marshal(sample)
		if err != nil {
			log.Print(err)
		}
		s := string(jsonString)
		log.Print(s)
		socket.Send(s, 0)
	}
}

func main() {
	go readLicor()

	http.Handle("/metrics", prometheus.Handler())
	http.ListenAndServe(":9092", nil)
}
