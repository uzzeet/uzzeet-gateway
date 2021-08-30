package serror

import (
	"fmt"
	"github.com/uzzeet/uzzeet-gateway/libs"
	"github.com/uzzeet/uzzeet-gateway/libs/helper"
	"os"
	"strings"
	//"syscall"
)

func isLocal() bool {
	return strings.ToLower(helper.Env(libs.AppEnv, libs.EnvLocal)) == libs.EnvLocal
}

func printErr(m string) {
	fmt.Fprintln(os.Stderr, m)
}

func exit() {
	/*err := syscall.Kill(syscall.Getpid(), syscall.SIGINT)
	if err != nil {
		os.Exit(1)
	}*/
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
