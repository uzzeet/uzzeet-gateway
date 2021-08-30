package manual

import (
	"fmt"
	"github.com/uzzeet/uzzeet-gateway/controller/resolver"
	"github.com/uzzeet/uzzeet-gateway/libs/helper/serror"
)

type manualResolver struct{}

func NewManualResolver() (res resolver.Resolver, errx serror.SError) {
	res = manualResolver{}
	return res, errx
}

func (ox manualResolver) GenerateURL(service string, port string) (url string) {
	url = fmt.Sprintf("%s:%s", service, port)
	return url
}

func (ox manualResolver) Register() (errx serror.SError) {
	return errx
}
