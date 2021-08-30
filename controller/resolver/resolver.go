package resolver

import (
	"github.com/uzzeet/uzzeet-gateway/libs/helper/serror"
)

type OptionKey string

type GenerateOptions map[OptionKey]interface{}

type Resolver interface {
	Register() (errx serror.SError)
	GenerateURL(name string, port string) (url string)
}
