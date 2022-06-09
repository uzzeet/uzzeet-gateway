package controller

import (
	"fmt"
	redis "github.com/go-redis/redis/v7"
	"github.com/uzzeet/uzzeet-gateway/controller/resolver"
	"github.com/uzzeet/uzzeet-gateway/libs"
	"go.elastic.co/apm/module/apmgrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/balancer/roundrobin"

	redisreg "github.com/uzzeet/uzzeet-gateway/controller/redis"
	"github.com/uzzeet/uzzeet-gateway/libs/helper"
	"github.com/uzzeet/uzzeet-gateway/libs/helper/logger"
	"github.com/uzzeet/uzzeet-gateway/service"
)

const (
	DefaultRegistryKey = "gateway"
)

type RegistryConfig struct {
	Address  string
	Password string
}

type Registry struct {
	conn     *redisreg.Connection
	reader   service.RegistryReader
	writer   service.RegistryWriter
	resolver resolver.Resolver
}

type Connection interface {
	Kind() string
}

func InitRegistry(cfg RegistryConfig) (*Registry, error) {
	logger.Infof("initializing registry with controller %v", cfg)
	redisConn, err := redisreg.NewConnection(&redis.Options{
		Addr:     cfg.Address,
		Password: cfg.Password,
	})
	if err != nil {
		return nil, err
	}
	reg := redisreg.NewRegistry(helper.Env(libs.AppKeyGateway, DefaultRegistryKey), redisConn)

	resolv, errx := service.NewResolver()
	if errx != nil {
		return nil, errx.Cause()
	}

	return &Registry{redisConn, reg, reg, resolv}, nil
}

func (reg Registry) GetConnection(key string) (*grpc.ClientConn, error) {
	logger.Infof("get service controller %s on registry", key)
	cfg, err := reg.reader.GetByKey(key)
	if err != nil {
		return nil, fmt.Errorf("while fetching controller from registry: %v", err)
	}

	connURL := reg.resolver.GenerateURL(cfg.Host, helper.IntToString(cfg.Port))
	logger.Infof("try dialing to service %s(%s) via gRPC", key, connURL)

	conn, err := grpc.Dial(
		connURL,
		grpc.WithInsecure(),
		grpc.WithBalancerName(roundrobin.Name),
		grpc.WithDefaultCallOptions(
			grpc.MaxCallRecvMsgSize(150*1024*1024),
		),
		grpc.WithUnaryInterceptor(apmgrpc.NewUnaryClientInterceptor()),
	)
	if err != nil {
		return nil, fmt.Errorf("while dialing server: %v", err)
	}
	logger.Infof("connected with %s", key)

	return conn, nil
}

func (reg Registry) ForkConnection() Connection {
	return reg.conn
}
