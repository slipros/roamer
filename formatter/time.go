package formatter

import (
	"reflect"
	"strings"
	"time"

	"github.com/pkg/errors"
	rerr "github.com/slipros/roamer/err"
)

// defaultTimeFormatters defines the built-in time formatting functions.
var defaultTimeFormatters = TimeFormatters{
	"timezone":     applyTimezone,
	"truncate":     applyTruncate,
	"start_of_day": wrapTimeFunc(applyStartOfDay),
	"end_of_day":   wrapTimeFunc(applyEndOfDay),
}

// TimeFormatterFunc is a function type for time transformations.
type TimeFormatterFunc = func(t *time.Time, arg string) error

// TimeFormatters is a map of named time formatting functions.
type TimeFormatters map[string]TimeFormatterFunc

const (
	// TagTime is the struct tag name used for time formatting.
	TagTime = "time"
)

// Time is a formatter for time.Time values.
// It applies transformations to time fields based on the "time" struct tag.
type Time struct {
	formatters TimeFormatters
}

// NewTime creates a Time formatter.
func NewTime(opts ...TimeOptionsFunc) *Time {
	t := &Time{
		formatters: make(TimeFormatters),
	}

	for name, fn := range defaultTimeFormatters {
		t.formatters[name] = fn
	}

	for _, opt := range opts {
		opt(t)
	}

	return t
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

	for _, f := range strings.Split(tagValue, SplitSymbol) {
		name, arg := ParseFormatter(f)

		formatter, ok := t.formatters[name]
		if !ok {
			return errors.WithStack(rerr.FormatterNotFound{Tag: TagTime, Formatter: name})
		}

		if err := formatter(v, arg); err != nil {
			return err
		}
	}

	return nil
}

// wrapTimeFunc wraps a simple time function to match TimeFormatterFunc signature
func wrapTimeFunc(fn func(t *time.Time) error) TimeFormatterFunc {
	return func(t *time.Time, _ string) error {
		return fn(t)
	}
}

func applyTimezone(t *time.Time, arg string) error {
	loc, err := time.LoadLocation(arg)
	if err != nil {
		return errors.Wrapf(err, "invalid timezone: %s", arg)
	}
	*t = t.In(loc)
	return nil
}

func applyTruncate(t *time.Time, arg string) error {
	d, err := parseDuration(arg)
	if err != nil {
		return err
	}
	*t = t.Truncate(d)
	return nil
}

func applyStartOfDay(t *time.Time) error {
	*t = time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
	return nil
}

func applyEndOfDay(t *time.Time) error {
	y, m, d := t.Date()
	*t = time.Date(y, m, d, 23, 59, 59, int(time.Second-time.Nanosecond), t.Location())
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
