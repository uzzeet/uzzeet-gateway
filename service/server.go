package service

import (
	"fmt"
	"github.com/uzzeet/uzzeet-gateway/libs"
	"net"

	"github.com/uzzeet/uzzeet-gateway/libs/helper"
	"github.com/uzzeet/uzzeet-gateway/libs/helper/logger"
	"go.elastic.co/apm/module/apmgrpc"
	"google.golang.org/grpc"

	"github.com/uzzeet/uzzeet-gateway/packets"
)

type Server struct {
	cfg      Config
	svc      *Service
	instance *grpc.Server
	listener net.Listener
	reg      RegistryWriter
}

func NewServer(cfg Config, reg RegistryWriter) (*Server, error) {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", cfg.Port))
	if err != nil {
		return nil, fmt.Errorf("while opening listener: %v", err)
	}

	return &Server{
		cfg:      cfg,
		instance: grpc.NewServer(grpc.UnaryInterceptor(apmgrpc.NewUnaryServerInterceptor(apmgrpc.WithRecovery()))),
		listener: listener,
		reg:      reg,
	}, nil
}

func NewServerHttp(cfg Config, reg RegistryWriter) (*Server, error) {
	return &Server{
		cfg:      cfg,
		instance: grpc.NewServer(grpc.UnaryInterceptor(apmgrpc.NewUnaryServerInterceptor(apmgrpc.WithRecovery()))),
		reg:      reg,
	}, nil
}

func (svr *Server) AsGatewayService(baseEndpoint string) *Service {
	svr.cfg.gatewayEndpoint = baseEndpoint
	svc := &Service{
		key:          helper.Env(libs.AppName, libs.AppName),
		namespace:    helper.Env(libs.AppNamespace, libs.NamespaceDefault),
		baseEndpoint: baseEndpoint,
		checksum:     svr.cfg.checksum(),
		router: router{
			routes:          make(map[string][]Route),
			protectedRoutes: make(map[string][]protectedRoute),
		},
	}

	packets.RegisterServiceServer(svr.instance, svc)

	return svc
}

func (svr Server) Instance() *grpc.Server {
	return svr.instance
}

func (svr Server) Start() error {
	err := svr.reg.Write(svr.cfg)
	if err != nil {
		return fmt.Errorf("while writing controller: %v", err)
	}

	if svr.cfg.HasGatewayEndpoint() {
		logger.Infof("send notify to gateway with controller %v", svr.cfg)
		err := svr.reg.Publish(svr.cfg)
		if err != nil {
			return fmt.Errorf("while publishing controller: %v", err)
		}
	}

	logger.Infof("service is listening on %s.", fmt.Sprintf("%s:%d", svr.cfg.Host, svr.cfg.Port))
	return svr.instance.Serve(svr.listener)
}

func (svr Server) Write() error {
	err := svr.reg.Write(svr.cfg)
	if err != nil {
		return fmt.Errorf("while writing controller: %v", err)
	}

	if svr.cfg.HasGatewayEndpoint() {
		logger.Infof("send notify to gateway with controller %v", svr.cfg)
		err := svr.reg.Publish(svr.cfg)
		if err != nil {
			return fmt.Errorf("while publishing controller: %v", err)
		}
	}

	return nil
}

func (svr Server) Stop() error {
	svr.Stop()

	err := svr.listener.Close()
	if err != nil {
		return fmt.Errorf("while closing listener: %v", err)
	}

	return nil
}
