package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/go-chi/chi"
	"github.com/uzzeet/uzzeet-gateway/controller/auth"
	"github.com/uzzeet/uzzeet-gateway/models"
	"github.com/uzzeet/uzzeet-gateway/service"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
)

func (fwd chiForwarder) agentIdentification(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var agent string

		agentToken := strings.Split(r.Header.Get("User-Agent"), "/")
		if len(agentToken) >= 1 {
			agent = agentToken[0]
		}

		clientInfo := &models.ClientInfo{
			Agent: agent,
		}

		r = r.WithContext(context.WithValue(r.Context(), models.ClientInfoContextValueKey, clientInfo))

		next.ServeHTTP(w, r)
	})
}

func (fwd chiForwarder) serviceIdentification(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var composite *service.Composite

		serviceName := chi.URLParam(r, "service")
		for _, each := range fwd.composite {
			if each.Endpoints() == fmt.Sprintf("/%s", serviceName) {
				composite = each

				break
			}
		}

		if composite != nil {
			basePath := composite.Endpoints()
			path := r.URL.Path[strings.Index(r.URL.Path, basePath)+len(basePath):]
			if path == "" {
				path = "/"
			}

			r = r.WithContext(context.WithValue(r.Context(), models.ServiceContextValueKey, composite))
			r = r.WithContext(context.WithValue(r.Context(), models.PathContextValueKey, path))

			next.ServeHTTP(w, r)

			return
		}

		fwd.notFound(serviceName, w, r)
	})
}

func (fwd chiForwarder) authorization(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		composite := r.Context().Value(models.ServiceContextValueKey).(*service.Composite)
		path := r.Context().Value(models.PathContextValueKey).(string)
		if needProtection, isStrict, isPrivate := composite.IsNeedProtection(r.Method, path); needProtection {
			authService := fwd.authService

			switch {
			case isStrict:
				authService = fwd.strictAuthService

			case isPrivate:
				authService = fwd.privateAuthService
			}

			body, err := ioutil.ReadAll(r.Body)
			if err != nil {
				err := fmt.Errorf("while reading body: %v", err)
				logger(json.NewEncoder(w).Encode(models.Response{
					Response:   http.StatusInternalServerError,
					Error:      err.Error(),
					Controller: r.RequestURI,
					Action:     r.Method,
				}))

				return
			}

			timestamp, _ := time.Parse(models.TimestampFormat, r.Header.Get(models.TimestampHeaderKey))
			authInfo, err := authService.Authorize(models.Request{
				Method:      r.Method,
				CompositeID: models.CompositeID(r.Header.Get(models.ClientIDHeaderKey)),
				URL:         r.URL,
				Token:       r.Header.Get(models.AuthorizationHeaderKey),
				Body:        body,
				Timestamp:   timestamp,
				Signature:   r.Header.Get(models.SignatureHeaderKey),
			})
			if err != nil {
				switch erx := err.(type) {
				default:
					w.Header().Set(models.ContentTypeHeaderKey, models.ContentTypeValueJSON)
					logger(json.NewEncoder(w).Encode(models.Response{
						Response:   http.StatusInternalServerError,
						Error:      err.Error(),
						Controller: r.RequestURI,
						Action:     r.Method,
					}))

				case auth.AuthorizationError:
					authErr := erx
					fwd.unauthorized(w, authErr.Message(), []models.Error{
						{
							InternalMessage: authErr.Error(),
						},
					})
				}

				return
			}

			r = r.WithContext(context.WithValue(r.Context(), models.AuthorizationInfoContextValueKey, authInfo))
			r.Body = ioutil.NopCloser(bytes.NewBuffer(body))
		}

		next.ServeHTTP(w, r)
	})
}
