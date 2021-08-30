package handler

import (
	"net"
	"net/http"
	"strings"

	"github.com/asaskevich/govalidator"

	"github.com/uzzeet/uzzeet-gateway/libs/helper"
	"github.com/uzzeet/uzzeet-gateway/libs/helper/serror"
)

var cidrs []*net.IPNet

func init() {
	maxCidrBlocks := []string{
		"::1/128",
		"fc00::/7",
		"fe80::/10",
	}

	cidrs = make([]*net.IPNet, len(maxCidrBlocks))
	for i, maxCidrBlock := range maxCidrBlocks {
		_, cidr, _ := net.ParseCIDR(maxCidrBlock)
		cidrs[i] = cidr
	}
}

func isPrivateAddress(address string) (bool, serror.SError) {
	ipAddress := net.ParseIP(address)
	if ipAddress == nil {
		return false, serror.New("Address is not valid")
	}

	for i := range cidrs {
		if cidrs[i].Contains(ipAddress) {
			return true, nil
		}
	}

	return false, nil
}

func RecoverRealIP(r *http.Request, skip int) (ip string, ok bool) {
	const localhost = "127.0.0.1"
	var ipTmp string

	ip = localhost

	{
		for _, v := range r.Header["X-Forwarded-For"] {
			if ok {
				break
			}

			xForwardFors := helper.CleanSpit(v, ",")
			for i := len(xForwardFors) - 1; i >= 0; i-- {
				v2 := xForwardFors[i]

				if isPrivate, err := isPrivateAddress(v2); !isPrivate && err == nil {
					ip, ok = v2, true
					break
				}

				if ip == localhost && govalidator.IsIPv4(v2) {
					ipTmp = helper.Chains(ipTmp, v2)

					skip--
					if skip < 0 {
						ip = v2
					}
				}
			}
		}
	}

	if !ok {
		remoteAddr := r.RemoteAddr
		if strings.Contains(remoteAddr, ":") {
			remoteAddr, _, _ = net.SplitHostPort(remoteAddr)
		}

		if isPrivate, err := isPrivateAddress(remoteAddr); !isPrivate && err == nil {
			ip, ok = remoteAddr, true
		}
	}

	if !ok {
		xRealIp := r.Header.Get("X-Real-Ip")
		if isPrivate, err := isPrivateAddress(xRealIp); !isPrivate && err == nil {
			ip = xRealIp
		}
	}

	if ip == localhost {
		ip = helper.Chains(ipTmp, ip)
	}

	return ip, ok
}
