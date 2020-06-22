package main

import (
	"context"
	"log"

	"./src/core"
	"./src/service"
)


func main() {
	conf_file := "ic1101.yaml"
	conf := core.Config{}

	core.DefaultConfig(&conf)
	err := core.ReadConfig(&conf, conf_file)

	if nil != err {
		log.Print("Cannot read config file, ", err)
	}

	mg := core.ConnectMongo(&conf)
	defer mg.Disconnect(context.TODO())
	service.Install(&conf, mg)
}