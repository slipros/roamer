package formatter

import (
	"reflect"
	"strings"
	"time"

	"github.com/pkg/errors"
	rerr "github.com/slipros/roamer/err"
)

const (
	// TagTime is the struct tag name used for time formatting.
	TagTime = "time"
)

// Time is a formatter for time.Time values.
// It applies transformations to time fields based on the "time" struct tag.
type Time struct{}

// NewTime creates a Time formatter.
func NewTime() *Time {
	return &Time{}
}

// Tag returns the name of the struct tag that this formatter handles.
func (t *Time) Tag() string {
	return TagTime
}

// Format applies time formatters to a field value based on the struct tag.
func (t *Time) Format(tag reflect.StructTag, ptr any) error {
	tagValue, ok := tag.Lookup(TagTime)
	if !ok {
		return nil
	}

	v, ok := ptr.(*time.Time)
	if !ok {
		return errors.Wrapf(rerr.NotSupported, "time formatter for %T", ptr)
	}

	rules := strings.Split(tagValue, SplitSymbol)
	for _, rule := range rules {
		name, arg := parseRule(rule)
		switch name {
		case "timezone":
			loc, err := time.LoadLocation(arg)
			if err != nil {
				return errors.Wrapf(err, "invalid timezone: %s", arg)
			}
			*v = v.In(loc)
		case "truncate":
			d, err := parseDuration(arg)
			if err != nil {
				return err
			}
			*v = v.Truncate(d)
		case "start_of_day":
			*v = time.Date(v.Year(), v.Month(), v.Day(), 0, 0, 0, 0, v.Location())
		case "end_of_day":
			y, m, d := v.Date()
			*v = time.Date(y, m, d, 23, 59, 59, int(time.Second-time.Nanosecond), v.Location())
		}
	}

	return nil
}

func parseDuration(arg string) (time.Duration, error) {
	switch arg {
	case "hour":
		return time.Hour, nil
	case "minute":
		return time.Minute, nil
	case "second":
		return time.Second, nil
	default:
		// Attempt to parse as a standard duration string (e.g., "1h30m")
		d, err := time.ParseDuration(arg)
		if err != nil {
			return 0, errors.Wrapf(err, "invalid duration: %s", arg)
		}

		return d, nil
	}
}
