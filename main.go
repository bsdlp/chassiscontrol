package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"

	"github.com/bsdlp/chassiscontrol/configuration"
	"github.com/bsdlp/chassiscontrol/internal/server"
	"github.com/bsdlp/chassiscontrol/rpc/chassis"
)

func main() {
	// optional config file path
	configFilePath, ok := os.LookupEnv("WOL_CONFIG_FILE")
	if !ok {
		configFilePath = "/config"
	}

	// read config file
	var cfg configuration.Object
	if configFh, err := os.Open(configFilePath); err == nil {
		err = json.NewDecoder(configFh).Decode(&cfg)
		if err != nil {
			log.Fatalf("error decoding config: %s", err.Error())
		}
		configFh.Close()
	} else {
		log.Fatalf("error reading config file: %s", err.Error())
	}

	srv := &server.Server{
		Targets: cfg.Targets,
	}

	err := http.ListenAndServe(cfg.ServerHostPort, chassis.NewChassisControlServer(srv))
	if err != nil {
		log.Fatal(err.Error())
	}
}
