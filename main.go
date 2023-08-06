package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"

	"github.com/gebn/bmc"
	"github.com/gebn/bmc/pkg/ipmi"
	"github.com/kelseyhightower/envconfig"
)

type config struct {
	ServerHostPort string
	Username       string
	Password       string
	IPMIHost       string
}

type server struct {
	cfg *config
}

func (srv *server) sessionOpts() *bmc.SessionOpts {
	return &bmc.SessionOpts{
		Username:          srv.cfg.Username,
		Password:          []byte(srv.cfg.Password),
		MaxPrivilegeLevel: ipmi.PrivilegeLevelUser,
	}
}

type GetPowerStatusResponse struct {
	PoweredOn         bool
	PowerControlFault bool
}

func (srv *server) GetPowerStatus(w http.ResponseWriter, r *http.Request) {
	transport, err := bmc.Dial(context.Background(), srv.cfg.IPMIHost)
	if err != nil {
		http.Error(w, "unable to connect to ipmi", http.StatusFailedDependency)
		log.Printf("unable to connect to ipmi: %s", err.Error())
		return
	}
	defer transport.Close()

	session, err := transport.NewSession(context.Background(), srv.sessionOpts())
	if err != nil {
		http.Error(w, "error creating new session", http.StatusFailedDependency)
		log.Printf("error creating new session: %s", err.Error())
		return
	}
	defer session.Close(context.Background())

	chassisStatus, err := session.GetChassisStatus(context.Background())
	if err != nil {
		http.Error(w, "command failed", http.StatusFailedDependency)
		log.Printf("command failed: %s", err.Error())
		return
	}

	err = json.NewEncoder(w).Encode(GetPowerStatusResponse{
		PoweredOn:         chassisStatus.PoweredOn,
		PowerControlFault: chassisStatus.PowerControlFault,
	})
	if err != nil {
		http.Error(w, "error encoding json", http.StatusInternalServerError)
		return
	}
}

func main() {
	var cfg config
	err := envconfig.Process("wol", &cfg)
	if err != nil {
		log.Fatal(err.Error())
	}

	srv := &server{
		cfg: &cfg,
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/power/status", srv.GetPowerStatus)

	err = http.ListenAndServe(cfg.ServerHostPort, mux)
	if err != nil {
		log.Fatal(err.Error())
	}
}
