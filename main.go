package main

import (
	"context"
	"fmt"
	"github.com/uzzeet/uzzeet-gateway/controller"
	"github.com/uzzeet/uzzeet-gateway/libs"
	"github.com/uzzeet/uzzeet-gateway/service/handler"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/cors"
	redis "github.com/go-redis/redis/v7"
	"github.com/joho/godotenv"
	"go.elastic.co/apm/module/apmchi"

	"github.com/uzzeet/uzzeet-gateway/controller/auth"
	redisreg "github.com/uzzeet/uzzeet-gateway/controller/redis"
	"github.com/uzzeet/uzzeet-gateway/libs/helper"
	"github.com/uzzeet/uzzeet-gateway/libs/helper/logger"
	"github.com/uzzeet/uzzeet-gateway/libs/helper/serror"
	"github.com/uzzeet/uzzeet-gateway/service"
)

var (
	httpServer         http.Server
	redisConn          *redisreg.Connection
	authService        auth.Service
	strictAuthService  auth.Service
	privateAuthService auth.Service

	fwd service.Forwarder
	reg service.RegistryReader
)

func init() {
	err := godotenv.Load()
	if err != nil {
		logger.Err(err)
		os.Exit(1)
	}
}

func init() {
	var (
		err          error
		useSignature bool
	)

	redisConn, err = redisreg.NewConnection(&redis.Options{
		Addr:     helper.Env(libs.AppRegistryAddr, "127.0.0.1:6379"),
		Password: helper.Env(libs.AppRegistryPwd, ""),
	})
	if err != nil {
		logger.Err(err)
		os.Exit(1)
	}

	redisReg := redisreg.NewRegistry(controller.DefaultRegistryKey, redisConn)
	if helper.Env(libs.AppEnv, libs.EnvProduction) == libs.EnvProduction {
		useSignature = true
	}

	reg = redisReg
	authService = auth.NewService("protect", helper.Env("APP_SECRET", "um_phrase"), redisReg, useSignature)
	strictAuthService = auth.NewService("strict", helper.Env("APP_STRICT_SECRET", "um_phrase"), redisReg, useSignature)
	privateAuthService = auth.NewService("private", helper.Env("APP_PRIVATE_SECRET", "um_phrase"), redisReg, useSignature)
}

func init() {
	mux := chi.NewMux()
	mux.Use(middleware.RequestID)
	mux.Use(middleware.RealIP)
	mux.Use(middleware.Logger)
	mux.Use(middleware.NoCache)
	mux.Use(middleware.Recoverer)
	mux.Use(middleware.DefaultCompress)
	mux.Use(middleware.Timeout(3 * time.Minute))
	mux.Use(apmchi.Middleware())
	mux.Use(cors.New(cors.Options{
		AllowedOrigins: []string{"https://*", "http://*"},
		AllowedMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders: []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token", "Accept-Encoding", "Cookie", "Origin", "X-Api-Key"},
	}).Handler)

	fwd = handler.NewChiForwarder(authService, strictAuthService, privateAuthService, mux.Route(helper.Env(libs.AppEndpoint, "/"), nil))
	httpServer = http.Server{
		Addr:         fmt.Sprintf(":%s", helper.Env(libs.AppPort, "9000")),
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 60 * time.Second,
		Handler:      mux,
	}
}

func main() {
	done := make(chan bool, 1)
	sc := make(chan os.Signal)
	g := controller.New(fwd, reg)

	err := g.Open()
	if err != nil {
		logger.Err(err)
		os.Exit(1)
	}

	go func() {
		logger.Infof("HTTP server is running and listening on %s", httpServer.Addr)
		err := httpServer.ListenAndServe()
		if err != nil {
			logger.Err(err)
			os.Exit(1)
		}
	}()

	go func() {
		for {
			select {
			case sc := <-sc:
				logger.Infof("receiving %s signal\nshutting down http server", sc.String())

				err := httpServer.Shutdown(context.Background())
				if err != nil {
					logger.Err(serror.NewFromErrorc(err, "while shutting down http server"))
				}

				logger.Info("closing redis connection")
				err = redisConn.Close()
				if err != nil {
					logger.Err(serror.NewFromErrorc(err, "while closing redis connection"))
				}

				done <- true
			}
		}
	}()

	signal.Notify(sc, os.Interrupt, os.Kill)

	<-done
}
