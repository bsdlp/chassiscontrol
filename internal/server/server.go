package server

import (
	"context"
	"errors"
	"log"

	"github.com/bsdlp/dell-wol/configuration"
	"github.com/bsdlp/dell-wol/rpc/chassis"
	"github.com/gebn/bmc"
	"github.com/gebn/bmc/pkg/ipmi"
	"github.com/twitchtv/twirp"
)

type Server struct {
	Targets map[string]configuration.IPMIHostConfig
}

type bmcSession struct {
	bmc.Session
	transport bmc.SessionlessTransport
}

func (session *bmcSession) close() {
	session.transport.Close()
	session.Session.Close(context.TODO())
}

func (srv *Server) dialBMC(ctx context.Context, targetConfig configuration.IPMIHostConfig) (*bmcSession, error) {
	transport, err := bmc.Dial(context.Background(), targetConfig.Address)
	if err != nil {
		log.Printf("unable to connect to ipmi: %s", err.Error())
		return nil, errors.New("error dialing bmc")
	}

	session, err := transport.NewSession(context.Background(), &bmc.SessionOpts{
		Username:          targetConfig.Username,
		Password:          []byte(targetConfig.Password),
		MaxPrivilegeLevel: ipmi.PrivilegeLevelAdministrator,
	})
	if err != nil {
		log.Printf("error creating new session: %s", err.Error())
		return nil, errors.New("error creating new session")
	}
	return &bmcSession{
		transport: transport,
		Session:   session,
	}, nil
}

func (srv *Server) GetChassisStatus(ctx context.Context, req *chassis.GetChassisStatusRequest) (*chassis.GetChassisStatusResponse, error) {
	targetConfig, ok := srv.Targets[req.Target]
	if !ok {
		return nil, twirp.NotFoundError("target not configured")
	}

	session, err := srv.dialBMC(ctx, targetConfig)
	if err != nil {
		return nil, twirp.FailedPrecondition.Error(err.Error())
	}
	defer session.close()

	chassisStatus, err := session.GetChassisStatus(context.Background())
	if err != nil {
		log.Printf("command failed: %s", err.Error())
		return nil, twirp.FailedPrecondition.Error("chassis control error")
	}

	return &chassis.GetChassisStatusResponse{
		Target:                     req.Target,
		PowerControlFault:          chassisStatus.PowerControlFault,
		PowerFault:                 chassisStatus.PowerFault,
		PowerOverload:              chassisStatus.PowerOverload,
		PoweredOn:                  chassisStatus.PoweredOn,
		PoweredOnByIpmi:            chassisStatus.PoweredOnByIPMI,
		LastPowerDownFault:         chassisStatus.LastPowerDownFault,
		LastPowerDownInterlock:     chassisStatus.LastPowerDownInterlock,
		LastPowerDownOverload:      chassisStatus.LastPowerDownOverload,
		LastPowerDownSupplyFailure: chassisStatus.LastPowerDownSupplyFailure,
		ResetButtonDisabled:        chassisStatus.ResetButtonDisabled,
		PowerOffButtonDisabled:     chassisStatus.PowerOffButtonDisabled,
	}, nil
}

var chassisControlCommandToIPMIEnum = map[chassis.ChassisControlCommand]ipmi.ChassisControl{
	chassis.ChassisControlCommand_OFF:                  ipmi.ChassisControlPowerOff,
	chassis.ChassisControlCommand_ON:                   ipmi.ChassisControlPowerOn,
	chassis.ChassisControlCommand_CYCLE:                ipmi.ChassisControlPowerCycle,
	chassis.ChassisControlCommand_RESET:                ipmi.ChassisControlHardReset,
	chassis.ChassisControlCommand_DIAGNOSTIC_INTERRUPT: ipmi.ChassisControlDiagnosticInterrupt,
	chassis.ChassisControlCommand_SOFT_POWER_OFF:       ipmi.ChassisControlSoftPowerOff,
}

func (srv *Server) IssueChassisControlCommand(ctx context.Context, req *chassis.ChassisControlRequest) (*chassis.ChassisControlResponse, error) {
	targetConfig, ok := srv.Targets[req.Target]
	if !ok {
		return nil, twirp.NotFoundError("target not configured")
	}

	session, err := srv.dialBMC(ctx, targetConfig)
	if err != nil {
		return nil, twirp.FailedPrecondition.Error(err.Error())
	}
	defer session.close()

	ipmiControlCommand, ok := chassisControlCommandToIPMIEnum[req.GetChassisControlCommand()]
	if !ok {
		return nil, twirp.InvalidArgumentError("ChassisControlCommand", "invalid command")
	}

	err = session.ChassisControl(ctx, ipmiControlCommand)
	if err != nil {
		return nil, twirp.FailedPrecondition.Error(err.Error())
	}

	return &chassis.ChassisControlResponse{
		Target: req.Target,
	}, nil
}
