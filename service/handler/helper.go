package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/uzzeet/uzzeet-gateway/models"
	"io/ioutil"
	"net/http"
	"strings"

	"google.golang.org/grpc/metadata"

	"github.com/uzzeet/uzzeet-gateway/libs/helper"
	"github.com/uzzeet/uzzeet-gateway/packets"
)

func transformRequestFromHttp(r *http.Request) (context.Context, *packets.Request, error) {
	path := r.Context().Value(models.PathContextValueKey).(string)
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return nil, nil, fmt.Errorf("while reading request body: %v", err)
	}

	ctx := r.Context()
	if ctx.Value(models.ClientInfoContextValueKey) != nil {
		if clientInfo, ok := r.Context().Value(models.ClientInfoContextValueKey).(*models.ClientInfo); ok {
			b, err := json.Marshal(clientInfo)
			if err != nil {
				return nil, nil, fmt.Errorf("while marshaling json: %v", err)
			}

			ctx = metadata.AppendToOutgoingContext(ctx, http.CanonicalHeaderKey(models.ClientInfoContextValueKey), string(b))
		}
	}

	if ctx.Value(models.AuthorizationInfoContextValueKey) != nil {
		if authInfo, ok := r.Context().Value(models.AuthorizationInfoContextValueKey).(*models.AuthorizationInfo); ok {
			b, err := json.Marshal(authInfo)
			if err != nil {
				return nil, nil, fmt.Errorf("while marshaling json: %v", err)
			}

			ctx = metadata.AppendToOutgoingContext(ctx, http.CanonicalHeaderKey(models.AuthorizationInfoContextValueKey), string(b))
		}
	}

	{
		ctx = metadata.AppendToOutgoingContext(ctx, http.CanonicalHeaderKey(models.BvXRemoteAddrTypeHeaderKey), r.RemoteAddr)

		{
			realIP, ok := RecoverRealIP(r, int(helper.StringToInt(helper.Env("DEFAULT_SKIP_FORWARDED_FOR", "0"), 0)))
			ctx = metadata.AppendToOutgoingContext(ctx, http.CanonicalHeaderKey(models.BvRealIPTypeHeaderKey), realIP)
			ctx = metadata.AppendToOutgoingContext(ctx, http.CanonicalHeaderKey(models.BvRealIPProofTypeHeaderKey), helper.BoolToString(ok))
		}
	}

	for key, vals := range r.Header {
		switch strings.ToLower(key) {
		case "x-remote-addr", "real-ip", "real-ip-proof":
			continue

		case "authorization":
			ctx = metadata.AppendToOutgoingContext(ctx, http.CanonicalHeaderKey(key), vals[0])
		}

		ctx = metadata.AppendToOutgoingContext(ctx, http.CanonicalHeaderKey(fmt.Sprintf("bv-%s", key)), vals[0])
	}

	if r.URL.RawQuery != "" {
		path = fmt.Sprintf("%s?%s", path, r.URL.RawQuery)
	}

	return ctx, &packets.Request{
		Method: r.Method,
		Path:   path,
		Body:   body,
	}, nil
}

func transformResponseToHTTP(resp *packets.Response, header metadata.MD, w http.ResponseWriter) error {
	for key, vals := range header {
		if http.CanonicalHeaderKey(key) == http.CanonicalHeaderKey(models.ContentTypeHeaderKey) {
			continue
		}

		if http.CanonicalHeaderKey(key) == http.CanonicalHeaderKey(models.BvContentTypeHeaderKey) {
			key = models.ContentTypeHeaderKey
		}

		w.Header().Set(key, vals[0])
	}

	w.WriteHeader(int(resp.Status))
	_, err := w.Write(resp.Body)
	if err != nil {
		return fmt.Errorf("while writing to response body: %v", err)
	}

	return nil
}
