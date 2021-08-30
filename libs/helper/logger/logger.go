package logger

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/uzzeet/uzzeet-gateway/libs/helper"
	"github.com/uzzeet/uzzeet-gateway/libs/helper/serror"
	"github.com/uzzeet/uzzeet-gateway/libs/utils/uttime"
)

type (
	ErrorLevel string
)

type Mode int

const (
	ModeDaily Mode = 1 + iota
	ModeMonthly
	ModeYearly
	ModePermanent
)

const (
	ErrorLevelDebug    ErrorLevel = "debug"
	ErrorLevelLog      ErrorLevel = "log"
	ErrorLevelInfo     ErrorLevel = "info"
	ErrorLevelCritical ErrorLevel = "critical"
	ErrorLevelWarning  ErrorLevel = "warn"
)

type (
	Options struct {
		Mode        Mode
		Path        string
		Writing     bool
		FileFormat  string
		Interceptor LogInterceptor
	}

	logger struct {
		sync.Mutex
		Path       string
		Writing    bool
		Mode       Mode
		FileFormat string

		_ready       bool
		_file        string
		_name        string
		_queues      []string
		_interceptor LogInterceptor
		_stream      *os.File
	}

	Logger interface {
		Startup() error
		Info(msg interface{})
		Infof(msg string, args ...interface{})
		Log(msg interface{})
		Logf(msg string, args ...interface{})
		Warn(msg interface{})
		Warnf(msg string, args ...interface{})
		Err(msg interface{})
		Errf(msg string, args ...interface{})
		Panic(msg interface{})
		IsWriting() bool
		StartWriting()
		StopWriting()
	}
)

func Construct(opt Options) Logger {
	log := &logger{
		Mode:       opt.Mode,
		Path:       opt.Path,
		Writing:    opt.Writing,
		FileFormat: helper.Chains(opt.FileFormat, "log-%v.log"),

		_ready:       false,
		_interceptor: opt.Interceptor,
	}
	return log
}

func (ox *logger) Startup() error {
	var err error
	if ox.Writing {
		cur := time.Now()
		fv := map[string]string{
			"d": cur.Format("02"),
			"m": cur.Format("01"),
			"y": cur.Format("2006"),
			"h": cur.Format("15"),
			"i": cur.Format("04"),
			"s": cur.Format("05"),
			"v": "",
		}

		formt := ox.FileFormat

		switch ox.Mode {
		case ModeDaily:
			fv["v"] = cur.Format("20060102")

		case ModeMonthly:
			fv["v"] = cur.Format("200601")

		case ModeYearly:
			fv["v"] = cur.Format("2006")

		case ModePermanent:
			fv["v"] = ""
		}

		for k, v := range fv {
			formt = strings.ReplaceAll(formt, "%"+k, v)
		}

		if ox._name != formt {
			ox._name = formt
			ox._file = filepath.Join(ox.Path, ox._name)

			if !helper.IsExists(ox.Path) {
				err = os.MkdirAll(ox.Path, os.ModePerm)
				if err != nil {
					return err
				}
			}

			err = ox.open()
			if err != nil {
				return err
			}
		}
	}

	if !ox._ready {
		go func() {
			for {
				time.Sleep(3 * time.Second)
				ox.flush()
			}
		}()
	}

	ox._ready = true
	return err
}

func (ox *logger) Info(msg interface{}) {
	lvl := ErrorLevelInfo

	m := ox._interceptor.Translate(lvl, msg)
	ox._interceptor.Process(lvl, m)
	_ = ox.write(m)
}

func (ox *logger) Infof(msg string, args ...interface{}) {
	ox.Info(fmt.Sprintf(msg, args...))
}

func (ox *logger) Log(msg interface{}) {
	lvl := ErrorLevelLog

	m := ox._interceptor.Translate(lvl, msg)
	ox._interceptor.Process(lvl, m)
	_ = ox.write(m)
}

func (ox *logger) Logf(msg string, args ...interface{}) {
	ox.Log(fmt.Sprintf(msg, args...))
}

func (ox *logger) Warn(msg interface{}) {
	lvl := ErrorLevelWarning

	m := ox._interceptor.Translate(lvl, msg)
	ox._interceptor.Process(lvl, m)
	_ = ox.write(m)
}

func (ox *logger) Warnf(msg string, args ...interface{}) {
	ox.Warn(fmt.Sprintf(msg, args...))
}

func (ox *logger) Err(msg interface{}) {
	lvl := ErrorLevelCritical

	m := ox._interceptor.Translate(lvl, msg)
	ox._interceptor.Process(lvl, m)
	_ = ox.write(m)
}

func (ox *logger) Errf(msg string, args ...interface{}) {
	ox.Err(serror.Newsf(1, msg, args...))
}

func (ox *logger) Panic(msg interface{}) {
	ox.Err(castToSError(msg, 1))
	exit()
}

func (ox *logger) IsReady() bool {
	return ox._ready
}

func (ox *logger) IsWriting() bool {
	return ox.Writing
}

func (ox *logger) StopWriting() {
	ox.Writing = false
}

func (ox *logger) StartWriting() {
	ox.Writing = true
}

func (ox *logger) open() error {
	if !ox.Writing {
		return nil
	}

	var err error

	ox.Lock()
	defer ox.Unlock()

	if ox._stream != nil {
		_ = ox._stream.Close()
	}

	ox._stream, err = os.OpenFile(ox._file, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}

	return err
}

func (ox *logger) write(m string) error {
	if !ox.Writing {
		return nil
	}

	if !ox._ready {
		return errors.New("Logger not yet ready")
	}

	if m == "" {
		return nil
	}

	ox.Lock()
	ox._queues = append(ox._queues, m)
	ox.Unlock()

	return nil
}

func (ox *logger) flush() error {
	if !ox.Writing {
		return nil
	}

	err := ox.Startup()
	if err != nil {
		return err
	}

	ox.Lock()
	lists := ox._queues
	ox._queues = []string{}
	ox.Unlock()

	defer func() {
		if err != nil {
			ox.Lock()
			ox._queues = append(lists, ox._queues...)
			ox.Unlock()
		}
	}()

	if len(lists) > 0 {
		for _, v := range lists {
			_, err = ox._stream.WriteString(fmt.Sprintf("%s\n", v))
			if err != nil {
				ox.printf("Failed to writing, details: %+v", err)

				errs := ox.open()
				if errs != nil {
					ox.printf("Failed to re-open file %s, details: %+v", ox._file, errs)
				}
				return err
			}
		}

		err = ox._stream.Sync()
		if err != nil {
			ox.printf("Failed to flushing stream, details: %+v", err)
			return err
		}
	}

	return err
}

func (ox *logger) printf(msg string, opts ...interface{}) {
	fmt.Printf("[%s] ERR: %s\n", uttime.Format(uttime.DefaultDateTimeFormat, time.Now()), fmt.Sprintf(msg, opts...))
}
