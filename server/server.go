package main

import (
	"errors"
	"fmt"
	"grpc-test/logger"
	"grpc-test/pb/protogo"
	"net"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
)

const (
	port        = "16666"
	tcpFlag     = true
	MinInterval = time.Duration(1) * time.Minute
	// ConnectionTimeout connection timeout time
	ConnectionTimeout = 5 * time.Second
	MaxSendSize       = 10
	MaxRevSize        = 10
)

type MsgServer struct {
	Listener net.Listener
	Server   *grpc.Server
}

var Logger = logger.NewLogger("server", "INFO")

func NewMsgServer() (*MsgServer, error) {

	var listener net.Listener
	var err error

	if tcpFlag {
		endPoint := fmt.Sprintf(":%s", port)
		listener, err = net.Listen("tcp", endPoint)
		if err != nil {
			return nil, err
		}
	} else {
		serverAddress, _ := net.ResolveUnixAddr("unix", "/tmp/grpc_test.sock")
		listener, err = net.ListenUnix("unix", serverAddress)
		if err != nil {
			return nil, err
		}
	}

	//set up server options for keepalive and TLS
	var serverOpts []grpc.ServerOption

	// add keepalive
	serverKeepAliveParameters := keepalive.ServerParameters{
		Time:    1 * time.Minute,
		Timeout: 20 * time.Second,
	}
	serverOpts = append(serverOpts, grpc.KeepaliveParams(serverKeepAliveParameters))

	//set enforcement policy
	kep := keepalive.EnforcementPolicy{
		MinTime:             MinInterval,
		PermitWithoutStream: true,
	}
	serverOpts = append(serverOpts, grpc.KeepaliveEnforcementPolicy(kep))

	//set default connection timeout

	serverOpts = append(serverOpts, grpc.ConnectionTimeout(ConnectionTimeout))
	serverOpts = append(serverOpts, grpc.MaxSendMsgSize(MaxSendSize*1024*1024))
	serverOpts = append(serverOpts, grpc.MaxRecvMsgSize(MaxRevSize*1024*1024))

	server := grpc.NewServer(serverOpts...)

	return &MsgServer{
		Listener: listener,
		Server:   server,
	}, nil
}

// StartMsgServer Start the server
func (s *MsgServer) StartMsgServer(apiInstance *API) error {

	var err error

	if s.Listener == nil {
		return errors.New("nil listener")
	}

	if s.Server == nil {
		return errors.New("nil server")
	}

	protogo.RegisterMessageRPCServer(s.Server, apiInstance)

	Logger.Infof("start message server")

	err = s.Server.Serve(s.Listener)
	if err != nil {
		Logger.Errorf("message server fail to start: %s", err)
		panic(err)
	}

	return nil
}

// StopCDMServer Stop the server
func (s *MsgServer) StopCDMServer() {
	Logger.Infof("stop cdm server")
	if s.Server != nil {
		s.Server.Stop()
	}
}

func main() {
	api := NewAPI()
	server, err := NewMsgServer()
	if err != nil {
		panic(err)
	}
	err = server.StartMsgServer(api)
	if err != nil {
		panic(err)
	}
}
