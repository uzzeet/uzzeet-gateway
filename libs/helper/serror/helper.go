package serror

import (
	"fmt"
	"github.com/uzzeet/uzzeet-gateway/libs"
	"github.com/uzzeet/uzzeet-gateway/libs/helper"
	"os"
	"strings"
	"syscall"
	//"syscall"
)

func isLocal() bool {
	return strings.ToLower(helper.Env(libs.AppEnv, libs.EnvLocal)) == libs.EnvLocal
}

func printErr(m string) {
	fmt.Fprintln(os.Stderr, m)
}

func exit() {
	p, _ := os.FindProcess(os.Getpid())
	_ = p.Signal(syscall.SIGTERM)
}

func getPath(val string) string {
	for _, v := range rootPaths {
		if strings.HasPrefix(val, v) {
			val = helper.Sub(val, len(v), 0)
			return val
		}
	}

	return val
}
