package logger

import (
	"fmt"
	"github.com/uzzeet/uzzeet-gateway/libs"
	"github.com/uzzeet/uzzeet-gateway/libs/helper"
	"github.com/uzzeet/uzzeet-gateway/libs/helper/serror"
	"strings"
)

var interceptor LogInterceptor = DefaultInterceptor()

func Info(msg interface{}) {
	lvl := ErrorLevelInfo

	m := interceptor.Translate(lvl, msg)
	interceptor.Process(lvl, m)
}

func Infof(msg string, args ...interface{}) {
	Info(fmt.Sprintf(msg, args...))
}

func Log(msg interface{}) {
	lvl := ErrorLevelLog

	m := interceptor.Translate(lvl, msg)
	interceptor.Process(lvl, m)
}

func Logf(msg string, args ...interface{}) {
	Log(fmt.Sprintf(msg, args...))
}

func Warn(msg interface{}) {
	lvl := ErrorLevelWarning

	m := interceptor.Translate(lvl, msg)
	interceptor.Process(lvl, m)
}

func Warnf(msg string, args ...interface{}) {
	Warn(fmt.Sprintf(msg, args...))
}

func Err(msg interface{}) {
	lvl := ErrorLevelCritical

	m := interceptor.Translate(lvl, msg)
	interceptor.Process(lvl, m)
}

func Panic(msg interface{}) {
	Err(castToSError(msg, 1))
	exit()
}

func isLocal() bool {
	return strings.ToLower(helper.Env(libs.AppEnv, libs.EnvLocal)) == libs.EnvLocal
}

func exit() {
	/*err := syscall.Kill(syscall.Getpid(), syscall.SIGINT)
	if err != nil {
		os.Exit(1)
	}*/
}

func castToSError(obj interface{}, skip int) serror.SError {
	var errx serror.SError

	if cur, ok := obj.(serror.SError); ok {
		errx = cur

	} else if cur, ok := obj.(error); ok {
		errx = serror.NewFromErrors(skip+1, cur)

	} else {
		errx = serror.Newsf(skip+1, "%+v", obj)
	}

	return errx
}
