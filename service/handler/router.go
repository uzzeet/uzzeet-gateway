package handler

import (
	redisreg "github.com/uzzeet/uzzeet-gateway/controller/redis"
	"sync"

	"github.com/go-chi/chi"

	"github.com/uzzeet/uzzeet-gateway/controller/auth"
	L "github.com/uzzeet/uzzeet-gateway/libs/helper/logger"
	"github.com/uzzeet/uzzeet-gateway/libs/helper/serror"
	"github.com/uzzeet/uzzeet-gateway/service"
)

type chiForwarder struct {
	mutex              *sync.Mutex
	authService        auth.Service
	strictAuthService  auth.Service
	privateAuthService auth.Service
	composite          map[string]*service.Composite
	rd                 *redisreg.Connection
}

func NewChiForwarder(authService, strictAuthService, privateAuthService auth.Service, r chi.Router, rd *redisreg.Connection) service.Forwarder {
	handler := &chiForwarder{&sync.Mutex{}, authService, strictAuthService, privateAuthService, make(map[string]*service.Composite), rd}

	r.Use(handler.agentIdentification)
	r.Get("/", handler.hello)
	r.Route("/{service}", func(r chi.Router) {
		r.With(
			handler.serviceIdentification,
			handler.authorization,
		).HandleFunc("/*", handler.forward)
	})

	return handler
}

func (fwd *chiForwarder) Mount(composite *service.Composite) {
	fwd.mutex.Lock()
	if old, ok := fwd.composite[composite.Keys()]; ok {
		err := old.Stop()
		if err != nil {
			L.Err(serror.NewFromErrorc(err, "while mount service"))
		}
	}

	fwd.composite[composite.Keys()] = composite
	fwd.mutex.Unlock()
}
