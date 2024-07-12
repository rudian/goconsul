package main

import (
	"github.com/rudian/goconsul/consul"
	"log"
	"time"
)

func main() {
	client, err := consul.NewService("test", "http://localhost:8500")
	defer client.DeregisterAllService()
	if err != nil {
		return
	}

	err1 := client.RegisterService(consul.RegisterService{
		ServiceName:   "hello_service",
		Address:       "127.0.0.1",
		Port:          3000,
		HeathCheckTTL: 3 * time.Second,
	})
	if err1 != nil {
		log.Fatalln(err1)
		return
	}

	time.Sleep(10 * time.Minute)
}
