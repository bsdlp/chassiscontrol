package main

import (
	"log"
	"net/http"

	"github.com/bsdlp/chassiscontrol/configuration"
	"github.com/bsdlp/chassiscontrol/internal/server"
	"github.com/bsdlp/chassiscontrol/rpc/chassis"
	"github.com/kelseyhightower/envconfig"
)

func main() {
	var cfg configuration.Object
	err := envconfig.Process("wol", &cfg)
	if err != nil {
		log.Fatal(err.Error())
	}

	srv := &server.Server{
		Targets: cfg.Targets,
	}

	err = http.ListenAndServe(cfg.ServerHostPort, chassis.NewChassisControlServer(srv))
	if err != nil {
		log.Fatal(err.Error())
	}
}
