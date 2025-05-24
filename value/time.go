package value

import (
	"reflect"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/pkg/errors"
)

// typeTime is a reflect.Type for the time.Time type.
// It's used for type comparison when handling time values.
var typeTime = reflect.TypeOf(time.Time{})

// timeLayouts contains common time formats used for parsing time strings.
// The layouts are ordered by frequency of use for optimal parsing performance.
var timeLayouts = []string{
	time.RFC3339,
	time.RFC3339Nano,
	time.RFC1123,
	time.RFC1123Z,
	time.RFC822,
	time.RFC822Z,
	time.RFC850,
	"2006-01-02",
	"2006-01-02 15:04:05",
	"2006-01-02T15:04:05",
	"2006-01-02 15:04:05.999999999 -0700 MST",
	"01/02/2006",
	"01/02/2006 15:04:05",
}

var globalTimeCache = &timeFormatCache{
	formats: make(map[string]string),
	maxSize: len(timeLayouts), // Limit cache size
}

var (
	// Pre-compiled regex patterns for faster format detection
	rfc3339Pattern  = regexp.MustCompile(`^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}`)
	dateOnlyPattern = regexp.MustCompile(`^\d{4}-\d{2}-\d{2}$`)
	dateTimePattern = regexp.MustCompile(`^\d{4}-\d{2}-\d{2}\s\d{2}:\d{2}:\d{2}$`)
	usDatePattern   = regexp.MustCompile(`^\d{2}/\d{2}/\d{4}`)
)

// timeFormatCache provides thread-safe caching of successful time format matches
// to improve performance when parsing time strings with known formats.
type timeFormatCache struct {
	mu      sync.RWMutex
	formats map[string]string
	maxSize int
}

func (c *timeFormatCache) getFormat(timeStr string) string {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return c.formats[timeStr]
}

func (c *timeFormatCache) setFormat(timeStr, format string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if len(c.formats) >= c.maxSize {
		c.formats = make(map[string]string, c.maxSize)
	}

	c.formats[timeStr] = format
}

func parseTime(str string) (time.Time, error) {
	if format := globalTimeCache.getFormat(str); format != "" {
		if t, err := time.Parse(format, str); err == nil {
			return t, nil
		}
	}

	var likelyFormats []string

	switch {
	case dateOnlyPattern.MatchString(str):
		likelyFormats = []string{time.DateOnly}
	case rfc3339Pattern.MatchString(str):
		if strings.HasSuffix(str, "Z") {
			likelyFormats = []string{time.RFC3339}
		} else if strings.Contains(str, "+") || strings.Contains(str, "-") {
			likelyFormats = []string{time.RFC3339}
		} else {
			likelyFormats = []string{"2006-01-02T15:04:05"}
		}
	case dateTimePattern.MatchString(str):
		likelyFormats = []string{time.DateTime}
	case usDatePattern.MatchString(str):
		likelyFormats = []string{"01/02/2006", "01/02/2006 15:04:05"}
	default:
		// Fallback to length-based detection
		likelyFormats = getTimeFormatsByLength(len(str))
	}

	for _, format := range likelyFormats {
		if t, err := time.Parse(format, str); err == nil {
			globalTimeCache.setFormat(str, format)

			return t, nil
		}
	}

	for _, format := range timeLayouts {
		skip := false
		for _, checked := range likelyFormats {
			if format == checked {
				skip = true
				break
			}
		}
		if skip {
			continue
		}

		if t, err := time.Parse(format, str); err == nil {
			globalTimeCache.setFormat(str, format)

			return t, nil
		}
	}

	return time.Time{}, errors.Errorf("cannot parse '%s' as time.Time with any known layout", str)
}

func getTimeFormatsByLength(length int) []string {
	switch length {
	case 10:
		return []string{time.DateOnly}
	case 19:
		return []string{time.DateTime, "2006-01-02T15:04:05"}
	case 20:
		return []string{time.RFC3339}
	case 25:
		return []string{time.RFC3339}
	default:
		return []string{time.RFC3339, time.RFC3339Nano}
	}
}
