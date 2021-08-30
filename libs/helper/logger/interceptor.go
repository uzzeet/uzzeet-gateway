package logger

import (
	"fmt"
	"os"
	"runtime"
	"time"

	"github.com/uzzeet/uzzeet-gateway/libs/helper"
	"github.com/uzzeet/uzzeet-gateway/libs/helper/serror"
	"github.com/uzzeet/uzzeet-gateway/libs/utils/uttime"
)

type (
	LogInterceptor interface {
		Translate(lvl ErrorLevel, obj interface{}) string
		Process(lvl ErrorLevel, msg string)
	}

	defaultInterceptorObj struct{}
)

func (defaultInterceptorObj) Translate(lvl ErrorLevel, obj interface{}) string {
	return DefaultTranslate(lvl, obj)
}

func (defaultInterceptorObj) Process(lvl ErrorLevel, msg string) {
	DefaultProcess(lvl, msg)
}

func DefaultInterceptor() LogInterceptor {
	return defaultInterceptorObj{}
}

func DefaultTranslate(lvl ErrorLevel, obj interface{}) string {
	m, m2 := DefaultTransform(lvl, obj)

	if !isLocal() {
		m2 = m
	}

	cur := time.Now()

	type lvll struct {
		Label string
		Color helper.Color
	}

	lbl := "?"
	ls := map[ErrorLevel]lvll{
		ErrorLevelInfo:     lvll{"INFO", helper.LIGHT_BLUE},
		ErrorLevelLog:      lvll{"LOG", helper.LIGHT_GRAY},
		ErrorLevelWarning:  lvll{"WARN", helper.LIGHT_YELLOW},
		ErrorLevelCritical: lvll{"ERR", helper.RED},
	}
	if cur, ok := ls[lvl]; ok {
		lbl = cur.Label

		if isLocal() {
			lbl = helper.ApplyForeColor(lbl, cur.Color)
		}
	}

	return fmt.Sprintf("[%s] %s: %s", uttime.Format(uttime.DefaultDateTimeFormat, cur), lbl, m2)
}

func DefaultTransform(lvl ErrorLevel, obj interface{}) (plainMsg string, colorMsg string) {
	plainMsg = fmt.Sprintf("%v", obj)
	colorMsg = plainMsg

	switch lvl {
	case ErrorLevelCritical, ErrorLevelWarning:
		switch vx := obj.(type) {
		case serror.SError:
			plainMsg = vx.String()
			colorMsg = vx.ColoredString()

		case error:
			pc, fn, line, _ := runtime.Caller(4)
			plainMsg = fmt.Sprintf(serror.StandardFormat(), runtime.FuncForPC(pc).Name(), fn, line, plainMsg)
			colorMsg = fmt.Sprintf(serror.StandardColorFormat(), runtime.FuncForPC(pc).Name(), fn, line, colorMsg)
		}
	}

	return plainMsg, colorMsg
}

func DefaultProcess(lvl ErrorLevel, msg string) {
	if msg == "" {
		return
	}

	switch lvl {
	case ErrorLevelCritical, ErrorLevelWarning:
		DefaultStderr(msg)

	default:
		DefaultStdout(msg)
	}
}

func DefaultStdout(msg string) {
	fmt.Fprintln(os.Stdout, msg)
}

func DefaultStderr(msg string) {
	fmt.Fprintln(os.Stderr, msg)
}
