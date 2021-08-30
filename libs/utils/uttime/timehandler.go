package uttime

import (
	"time"

	"github.com/uzzeet/uzzeet-gateway/libs/helper/serror"
)

func Construct(zone string) (res *TimeHandler, errx serror.SError) {
	res = &TimeHandler{}

	var loc *time.Location
	loc, errx = GetTimezone(zone)
	if errx != nil {
		return nil, errx
	}
	res._loc = loc
	return
}

func (ox *TimeHandler) ChangeTimezone(zone string) serror.SError {
	loc, errx := GetTimezone(zone)
	if errx != nil {
		return errx
	}

	ox._loc = loc
	return nil
}

func (ox TimeHandler) Timezone() string {
	return ox._loc.String()
}

func (ox TimeHandler) Now() time.Time {
	res, _ := NowWithTimezone(ox._loc.String())
	return res
}

func (ox TimeHandler) FNow() time.Time {
	res, _ := ParseForceTimezone(Format(DefaultDateTimeFormat, time.Now()), ox._loc.String())
	return res
}

func (ox TimeHandler) Parse(value interface{}) (time.Time, serror.SError) {
	return ParseWithTimezone(value, ox._loc.String())
}

func (ox TimeHandler) MostParse(value interface{}) time.Time {
	res, errx := ox.Parse(value)
	if errx != nil {
		panic(errx.Error)
	}
	return res
}

func (ox TimeHandler) FParse(value interface{}) (time.Time, serror.SError) {
	return ParseForceTimezone(value, ox._loc.String())
}

func (ox TimeHandler) FMostParse(value interface{}) time.Time {
	res, errx := ox.FParse(value)
	if errx != nil {
		panic(errx.Error)
	}
	return res
}

func (ox TimeHandler) ToString(value time.Time) string {
	return ox.ToStringWithFormat(DefaultDateTimeFormat, value)
}

func (ox TimeHandler) FToString(value time.Time) string {
	return ox.FToStringWithFormat(DefaultDateTimeFormat, value)
}

func (ox TimeHandler) ToStringWithFormat(format DateFormat, value time.Time) string {
	current, _ := ox.Parse(value)
	return Format(format, current)
}

func (ox TimeHandler) FToStringWithFormat(format DateFormat, value time.Time) string {
	current, _ := ox.FParse(value)
	return Format(format, current)
}
