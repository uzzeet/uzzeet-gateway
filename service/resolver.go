package service

import (
	"github.com/uzzeet/uzzeet-gateway/controller/resolver"
	"github.com/uzzeet/uzzeet-gateway/controller/resolver/manual"
	"github.com/uzzeet/uzzeet-gateway/libs"
	"github.com/uzzeet/uzzeet-gateway/libs/helper"
	"github.com/uzzeet/uzzeet-gateway/libs/helper/serror"
)

func NewResolver() (resolv resolver.Resolver, errx serror.SError) {
	switch helper.Env(libs.AppCluster, libs.ClusterLocal) {

	default:
		resolv, errx = manual.NewManualResolver()
	}

	if errx != nil {
		return resolv, errx
	}

	errx = resolv.Register()
	if errx != nil {
		return resolv, errx
	}

	return resolv, errx
}
