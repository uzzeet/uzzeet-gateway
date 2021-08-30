package uttime

import (
	"fmt"
	"github.com/uzzeet/uzzeet-gateway/libs/helper"
	"reflect"
	"strings"
	"time"

	"github.com/gearintellix/u2"

	"github.com/uzzeet/uzzeet-gateway/libs/helper/serror"
)

const (
	Year4Digits   = "2006"
	Year2Digits   = "06"
	Month2Digits  = "01"
	Month1Digits  = "1"
	Day2Digits    = "02"
	Day1Digits    = "2"
	Hour2Digits   = "15"
	Minute2Digits = "04"
	Second2Digits = "05"
	Timezone      = "MST"
)

type DateFormat = string

const (
	DefaultDateFormat                 DateFormat = "Y-m-d"
	DefaultDateWithTimezoneFormat     DateFormat = "Y-m-d TZ"
	INDateFormat                      DateFormat = "d-m-Y"
	DefaultDateTimeFormat             DateFormat = "Y-m-d H:i:s"
	INDateTimeFormat                  DateFormat = "d-m-Y H:i:s"
	DefaultDateTimeWithTimezoneFormat DateFormat = "Y-m-d H:i:s TZ"
	DefaultTimeFormat                 DateFormat = "H:i:s"
)

var (
	EmptyTime   time.Time
	EmptyTimeFN func() time.Time

	materials = []string{
		"!" + time.RFC3339Nano,
		"!" + time.RFC3339,

		DefaultDateTimeFormat,
		"Y-m-d",
		"YmdHis",
		"Ymd",
		"Y-m-dTH:i:s.999999999",
		"Y-m-dTH:i:s",

		DefaultDateTimeWithTimezoneFormat,
		"Y-m-d H:i:s Z07:00",
		"Y-m-dZ07:00",
		"YmdHisZ07:00",
		"YmdZ07:00",

		"Y-m-d H:i",
		"Y-m-d H",
		"Y-m",
		"Y",
		"!" + time.RFC822Z,
		"!" + time.RFC822,
		"!" + time.RFC1123Z,
		"!" + time.RFC1123,
		"!" + time.RFC850,
	}
)

func getEmptyTime() time.Time {
	if EmptyTimeFN != nil {
		return EmptyTimeFN()
	}

	return EmptyTime
}

func Most(tim time.Time, errx serror.SError) time.Time {
	if errx != nil {
		errx.Panic()
	}

	return tim
}

func GoLayout(format DateFormat) string {
	rl := map[string]string{
		"Y":  Year4Digits,
		"y":  Year2Digits,
		"m":  Month2Digits,
		"M":  Month1Digits,
		"d":  Day2Digits,
		"D":  Day1Digits,
		"H":  Hour2Digits,
		"i":  Minute2Digits,
		"s":  Second2Digits,
		"TZ": Timezone,
	}

	for k, v := range rl {
		format = strings.ReplaceAll(format, k, v)
	}

	return format
}

func GetTimezone(zone string) (res *time.Location, errx serror.SError) {
	res = time.Local

	switch {
	case strings.HasPrefix(zone, "+"), strings.HasPrefix(zone, "-"):
		offset := helper.Sub(zone, 1, 0)
		if !helper.IsInteger(offset) {
			errx = serror.Newf("Invalid timezone offset %s", offset)
			return res, errx
		}

		offx := int(helper.StringToInt(offset, 0) * 60 * 60)
		if string(zone[0]) == "-" {
			offx *= -1
		}

		res = time.FixedZone(fmt.Sprintf("UTC%s", zone), offx)

	case zone == "UTC", zone == "0":
		res = time.UTC

	case zone == "@":
		res = time.Local

	default:
		var err error
		res, err = time.LoadLocation(zone)
		if err != nil {
			errx = serror.NewFromErrorc(err, "Failed to load location")
			return res, errx
		}
	}

	return res, errx
}

func WithTimezone(tim time.Time, zone string) (res time.Time, errx serror.SError) {
	res = tim

	var loc *time.Location
	loc, errx = GetTimezone(zone)
	if errx != nil {
		errx.AddComments("while get timezone")
		return res, errx
	}

	res = res.In(loc)
	return res, errx
}

func ForceTimezone(tim time.Time, zone string) (res time.Time, errx serror.SError) {
	res = tim

	var loc *time.Location
	loc, errx = GetTimezone(zone)
	if errx != nil {
		errx.AddComments("while get timezone")
		return res, errx
	}

	var (
		dfmt = GoLayout(DefaultDateTimeFormat)
		val  = tim.Format(dfmt)

		err error
	)

	res, err = time.ParseInLocation(dfmt, val, loc)
	if err != nil {
		errx = serror.NewFromErrorc(err, "Failed to parse in location")
		return res, errx
	}

	return res, errx
}

func Now() time.Time {
	return time.Now()
}

func NowWithTimezone(zone string) (res time.Time, errx serror.SError) {
	res = Now()
	res, errx = WithTimezone(res, zone)
	if errx != nil {
		errx.AddComments("while with timezone")
		return res, errx
	}

	return res, errx
}

func Compose(year int, month int, day int, hour int, minute int, second int) (time.Time, serror.SError) {
	vv := "__year__-__month__-__day__ __hour__:__minute__:__second__"
	vv = u2.Binding(vv, map[string]string{
		"year":   helper.LeftPad(helper.IntToString(year), 4, "0"),
		"month":  helper.LeftPad(helper.IntToString(month), 2, "0"),
		"day":    helper.LeftPad(helper.IntToString(day), 2, "0"),
		"hour":   helper.LeftPad(helper.IntToString(hour), 2, "0"),
		"minute": helper.LeftPad(helper.IntToString(minute), 2, "0"),
		"second": helper.LeftPad(helper.IntToString(second), 2, "0"),
	})

	res, errx := ParseWithFormat(DefaultDateTimeFormat, vv)
	if errx != nil {
		errx.AddComments("while parse with format")
		return res, errx
	}

	return res, nil
}

func ParseWithFormat(format DateFormat, value string) (res time.Time, errx serror.SError) {
	res = getEmptyTime()

	if len(format) <= 0 {
		format = "!" + time.RFC3339
	}

	switch {
	case strings.HasPrefix(format, "!"):
		format = helper.Sub(format, 1, 0)

	default:
		format = GoLayout(format)
	}

	var err error
	res, err = time.ParseInLocation(format, value, time.Local)
	if err != nil {
		errx = serror.NewFromErrorc(err, "Failed to parse time")
		return res, errx
	}

	return res, errx
}

func ParseUTCWithFormat(format DateFormat, value string) (res time.Time, errx serror.SError) {
	res = getEmptyTime()

	if len(format) <= 0 {
		format = "!" + time.RFC3339
	}

	switch {
	case strings.HasPrefix(format, "!"):
		format = helper.Sub(format, 1, 0)

	default:
		format = GoLayout(format)
	}

	var err error
	res, err = time.Parse(format, value)
	if err != nil {
		errx = serror.NewFromErrorc(err, "Failed to parse time")
		return res, errx
	}

	return res, errx
}

func ParseWithFormatAndTimezone(format DateFormat, value string, zone string) (res time.Time, errx serror.SError) {
	res = getEmptyTime()

	res, errx = ParseWithFormat(format, value)
	if errx != nil {
		errx.AddComments("while parse with format")
		return res, errx
	}

	res, errx = WithTimezone(res, zone)
	if errx != nil {
		errx.AddComments("while with timezone")
		return res, errx
	}

	return res, errx
}

func ParseUTCWithFormatAndTimezone(format DateFormat, value string, zone string) (res time.Time, errx serror.SError) {
	res = getEmptyTime()

	res, errx = ParseUTCWithFormat(format, value)
	if errx != nil {
		errx.AddComments("while parse utc with format")
		return res, errx
	}

	res, errx = WithTimezone(res, zone)
	if errx != nil {
		errx.AddComments("while with timezone")
		return res, errx
	}

	return res, errx
}

func ParseWithFormatAndForceTimezone(format DateFormat, value string, zone string) (res time.Time, errx serror.SError) {
	res = getEmptyTime()

	res, errx = ParseWithFormat(format, value)
	if errx != nil {
		errx.AddComments("while parse with format")
		return res, errx
	}

	res, errx = ForceTimezone(res, zone)
	if errx != nil {
		errx.AddComments("while force timezone")
		return res, errx
	}

	return res, errx
}

func ParseFromInteger(value int64) (res time.Time, errx serror.SError) {
	res = time.Unix(value, 0)
	return res, errx
}

func ParseUTCFromInteger(value int64) (res time.Time, errx serror.SError) {
	res = time.Unix(value, 0)
	res, errx = ForceTimezone(res, "UTC")
	if errx != nil {
		errx.AddComments("while force timezone")
		return res, errx
	}

	return res, errx
}

func ParseFromIntegerWithTimezone(value int64, zone string) (res time.Time, errx serror.SError) {
	res, errx = ParseFromInteger(value)
	if errx != nil {
		errx.AddComments("while parse from integer")
		return res, errx
	}

	res, errx = WithTimezone(res, zone)
	if errx != nil {
		errx.AddComments("while with timezone")
		return res, errx
	}

	return res, errx
}

func ParseUTCFromIntegerWithTimezone(value int64, zone string) (res time.Time, errx serror.SError) {
	res, errx = ParseUTCFromInteger(value)
	if errx != nil {
		errx.AddComments("while parse utc from integer")
		return res, errx
	}

	res, errx = WithTimezone(res, zone)
	if errx != nil {
		errx.AddComments("while with timezone")
		return res, errx
	}

	return res, errx
}

func ParseFromIntegerForceTimezone(value int64, zone string) (res time.Time, errx serror.SError) {
	res, errx = ParseFromInteger(value)
	if errx != nil {
		errx.AddComments("while parse from integer")
		return res, errx
	}

	res, errx = ForceTimezone(res, zone)
	if errx != nil {
		errx.AddComments("while force timezone")
		return res, errx
	}

	return res, errx
}

func Parse(value interface{}) (time.Time, serror.SError) {
	return ParseWithTimezone(value, "@")
}

func ParseWithTimezone(value interface{}, zone string) (res time.Time, errx serror.SError) {
	res = getEmptyTime()

	if helper.IsNil(value) {
		errx = serror.New("Cannot parse from nil")
		return res, errx
	}

	val := reflect.ValueOf(value)
	switch val.Kind() {
	case reflect.String:
		res, errx = ParseFromStringWithTimezone(val.String(), zone)
		if errx != nil {
			errx.AddComments("while parse from string with timezone")
			return res, errx
		}

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		res, errx = ParseFromIntegerWithTimezone(val.Int(), zone)
		if errx != nil {
			errx.AddComments("while parse from integer with timezone")
			return res, errx
		}

	case reflect.Float32, reflect.Float64:
		res, errx = ParseFromIntegerWithTimezone(int64(val.Float()), zone)
		if errx != nil {
			errx.AddComments("while parse from integer with timezone")
			return res, errx
		}

	default:
		switch val.Type().String() {
		case "time.Time":
			res, errx = WithTimezone(value.(time.Time), zone)
			if errx != nil {
				errx.AddComments("while with timezone")
				return res, errx
			}
		}
	}

	return res, errx
}

func ParseUTCWithTimezone(value interface{}, zone string) (res time.Time, errx serror.SError) {
	res = getEmptyTime()

	if helper.IsNil(value) {
		errx = serror.New("Cannot parse from nil")
		return res, errx
	}

	val := reflect.ValueOf(value)
	switch val.Kind() {
	case reflect.String:
		res, errx = ParseUTCFromStringWithTimezone(val.String(), zone)
		if errx != nil {
			errx.AddComments("while parse utc from string with timezone")
			return res, errx
		}

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		res, errx = ParseUTCFromIntegerWithTimezone(val.Int(), zone)
		if errx != nil {
			errx.AddComments("while parse utc from integer with timezone")
			return res, errx
		}

	case reflect.Float32, reflect.Float64:
		res, errx = ParseUTCFromIntegerWithTimezone(int64(val.Float()), zone)
		if errx != nil {
			errx.AddComments("while parse utc from integer with timezone")
			return res, errx
		}

	default:
		switch val.Type().String() {
		case "time.Time":
			res, errx = WithTimezone(value.(time.Time), zone)
			if errx != nil {
				errx.AddComments("while with timezone")
				return res, errx
			}
		}
	}

	return res, errx
}

func ParseForceTimezone(value interface{}, zone string) (res time.Time, errx serror.SError) {
	res = getEmptyTime()

	if helper.IsNil(value) {
		errx = serror.New("Cannot parse from nil")
		return res, errx
	}

	val := reflect.ValueOf(value)
	switch val.Kind() {
	case reflect.String:
		res, errx = ParseFromStringForceTimezone(val.String(), zone)
		if errx != nil {
			errx.AddComments("while parse from string force timezone")
			return res, errx
		}

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		res, errx = ParseFromIntegerForceTimezone(val.Int(), zone)
		if errx != nil {
			errx.AddComments("while parse from integer force timezone")
			return res, errx
		}

	case reflect.Float32, reflect.Float64:
		res, errx = ParseFromIntegerForceTimezone(int64(val.Float()), zone)
		if errx != nil {
			errx.AddComments("while parse from integer force timezone")
			return res, errx
		}

	default:
		switch val.Type().String() {
		case "time.Time":
			res, errx = ForceTimezone(value.(time.Time), zone)
			if errx != nil {
				errx.AddComments("while force timezone")
				return res, errx
			}
		}
	}

	return res, errx
}

func ParseFromString(value string) (res time.Time, errx serror.SError) {
	res = getEmptyTime()

	for _, f := range materials {
		res, errx = ParseWithFormat(f, value)
		if errx == nil {
			return res, errx
		}
	}

	if helper.IsInteger(value) {
		res, errx = ParseFromInteger(helper.StringToInt(value, 0))
		if errx == nil {
			return res, errx
		}
	}

	errx = serror.Newf("Cannot parse time from %s", value)
	return res, errx
}

func ParseUTCFromString(value string) (res time.Time, errx serror.SError) {
	res = getEmptyTime()

	for _, f := range materials {
		res, errx = ParseUTCWithFormat(f, value)
		if errx == nil {
			return res, errx
		}
	}

	if helper.IsInteger(value) {
		res, errx = ParseUTCFromInteger(helper.StringToInt(value, 0))
		if errx == nil {
			return res, errx
		}
	}

	errx = serror.Newf("Cannot parse time from %s", value)
	return res, errx
}

func ParseFromStringWithTimezone(value string, zone string) (res time.Time, errx serror.SError) {
	res, errx = ParseFromString(value)
	if errx != nil {
		errx.AddComments("while parse from string")
		return res, errx
	}

	res, errx = WithTimezone(res, zone)
	if errx != nil {
		errx.AddComments("while with timezone")
		return res, errx
	}

	return res, errx
}

func ParseUTCFromStringWithTimezone(value string, zone string) (res time.Time, errx serror.SError) {
	res, errx = ParseUTCFromString(value)
	if errx != nil {
		errx.AddComments("while parse utc from string")
		return res, errx
	}

	res, errx = WithTimezone(res, zone)
	if errx != nil {
		errx.AddComments("while with timezone")
		return res, errx
	}

	return res, errx
}

func ParseFromStringForceTimezone(value string, zone string) (res time.Time, errx serror.SError) {
	res, errx = ParseFromString(value)
	if errx != nil {
		errx.AddComments("while parse from string")
		return res, errx
	}

	res, errx = ForceTimezone(res, zone)
	if errx != nil {
		errx.AddComments("while force timezone")
		return res, errx
	}

	return res, errx
}

func ToString(format DateFormat, value time.Time) string {
	return Format(format, value)
}

func Format(format DateFormat, value time.Time) string {
	return value.Format(GoLayout(format))
}
