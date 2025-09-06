package formatter

import (
	"reflect"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	rerr "github.com/slipros/roamer/err"
)

func createTimeTestTag(value string) reflect.StructTag {
	return reflect.StructTag(`time:"` + value + `"`)
}

func TestNewTime(t *testing.T) {
	t.Parallel()

	t.Run("DefaultFormatters", func(t *testing.T) {
		t.Parallel()

		s := NewTime()
		require.NotNil(t, s)
		assert.Equal(t, TagTime, s.Tag())

		for name := range defaultTimeFormatters {
			_, ok := s.formatters[name]
			assert.True(t, ok, "default formatter %s not found", name)
		}
	})

	t.Run("WithCustomFormatter", func(t *testing.T) {
		t.Parallel()

		customFormatter := func(t *time.Time, arg string) error {
			return nil
		}

		f := NewTime(WithTimeFormatter("custom", customFormatter))
		require.NotNil(t, f)

		_, ok := f.formatters["custom"]
		assert.True(t, ok)
	})

	t.Run("WithCustomFormatters", func(t *testing.T) {
		t.Parallel()

		customFormatters := TimeFormatters{
			"custom1": func(t *time.Time, arg string) error { return nil },
			"custom2": func(t *time.Time, arg string) error { return nil },
		}

		f := NewTime(WithTimeFormatters(customFormatters))
		require.NotNil(t, f)

		_, ok := f.formatters["custom1"]
		assert.True(t, ok)

		_, ok = f.formatters["custom2"]
		assert.True(t, ok)
	})
}

func TestTime_Format_Successfully(t *testing.T) {
	t.Parallel()

	customFormatter := func(t *time.Time, arg string) error {
		*t = time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)
		return nil
	}

	f := NewTime(WithTimeFormatter("custom", customFormatter))

	// Define locations for tests
	pst, err := time.LoadLocation("America/Los_Angeles")
	require.NoError(t, err)
	est, err := time.LoadLocation("America/New_York")
	require.NoError(t, err)
	utcLoc, err := time.LoadLocation("UTC")
	require.NoError(t, err)

	// Original time in a specific location (e.g., PST)
	originalTime := time.Date(2024, 8, 28, 10, 30, 45, 0, pst)

	tests := []struct {
		name         string
		tag          reflect.StructTag
		input        time.Time
		expectedZone string
		expectedTime time.Time
	}{
		{
			name:         "Convert to UTC",
			tag:          createTimeTestTag("timezone=UTC"),
			input:        originalTime,
			expectedZone: "UTC",
			expectedTime: originalTime.In(utcLoc),
		},
		{
			name:         "Convert to EST",
			tag:          createTimeTestTag("timezone=America/New_York"),
			input:        originalTime,
			expectedZone: "EDT", // Note: It's EDT during summer
			expectedTime: originalTime.In(est),
		},
		{
			name:         "Truncate to hour",
			tag:          createTimeTestTag("truncate=hour"),
			input:        originalTime,
			expectedZone: "PDT", // Note: It's PDT during summer
			expectedTime: time.Date(2024, 8, 28, 10, 0, 0, 0, pst),
		},
		{
			name:         "Truncate to minute",
			tag:          createTimeTestTag("truncate=minute"),
			input:        originalTime,
			expectedZone: "PDT",
			expectedTime: time.Date(2024, 8, 28, 10, 30, 0, 0, pst),
		},
		{
			name:         "Timezone and Truncate",
			tag:          createTimeTestTag("timezone=UTC,truncate=hour"),
			input:        originalTime, // 10:30:45 PDT
			expectedZone: "UTC",
			expectedTime: time.Date(2024, 8, 28, 17, 0, 0, 0, utcLoc), // 10:30 PDT is 17:30 UTC, truncated to 17:00
		},
		{
			name:         "Start of day",
			tag:          createTimeTestTag("start_of_day"),
			input:        originalTime,
			expectedZone: "PDT",
			expectedTime: time.Date(2024, 8, 28, 0, 0, 0, 0, pst),
		},
		{
			name:         "End of day",
			tag:          createTimeTestTag("end_of_day"),
			input:        originalTime,
			expectedZone: "PDT",
			expectedTime: time.Date(2024, 8, 28, 23, 59, 59, int(time.Second-time.Nanosecond), pst),
		},
		{
			name:         "Custom formatter",
			tag:          createTimeTestTag("custom"),
			input:        originalTime,
			expectedZone: "UTC",
			expectedTime: time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC),
		},
		{
			name:         "No time tag",
			tag:          reflect.StructTag(`other:"tag"`),
			input:        originalTime,
			expectedZone: "PDT",
			expectedTime: originalTime,
		},
		{
			name:         "Truncate to second",
			tag:          createTimeTestTag("truncate=second"),
			input:        originalTime,
			expectedZone: "PDT",
			expectedTime: time.Date(2024, 8, 28, 10, 30, 45, 0, pst),
		},
		{
			name:         "Truncate with duration string",
			tag:          createTimeTestTag("truncate=1h30m"),
			input:        time.Date(2024, 8, 28, 10, 30, 45, 0, pst),
			expectedZone: "PDT",
			expectedTime: time.Date(2024, 8, 28, 9, 30, 0, 0, pst), // 10:30:45 truncated to 10:30
		},
		{
			name:         "Truncate with 1h duration string",
			tag:          createTimeTestTag("truncate=1h"),
			input:        time.Date(2024, 8, 28, 10, 30, 45, 0, pst),
			expectedZone: "PDT",
			expectedTime: time.Date(2024, 8, 28, 10, 0, 0, 0, pst),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			val := tc.input
			ptr := &val

			err := f.Format(tc.tag, ptr)
			require.NoError(t, err)

			zone, _ := val.Zone()
			assert.Equal(t, tc.expectedZone, zone)
			assert.True(t, tc.expectedTime.Equal(val), "Expected %v, got %v", tc.expectedTime, val)
		})
	}
}

func TestTime_Format_Failure(t *testing.T) {
	t.Parallel()

	f := NewTime()

	tests := []struct {
		name  string
		tag   reflect.StructTag
		input any
		errAs error
		errIs error
	}{
		{name: "Unsupported type", tag: createTimeTestTag("timezone=UTC"), input: new(int), errIs: rerr.NotSupported},
		{name: "Invalid timezone", tag: createTimeTestTag("timezone=Invalid/Timezone"), input: new(time.Time)},
		{name: "Invalid truncate duration", tag: createTimeTestTag("truncate=invalid"), input: new(time.Time)},
		{name: "Formatter not found", tag: createTimeTestTag("non_existent"), input: new(time.Time), errAs: &rerr.FormatterNotFound{}},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := f.Format(tc.tag, tc.input)
			if tc.errIs != nil {
				require.ErrorIs(t, err, tc.errIs)
			} else if tc.errAs != nil {
				require.ErrorAs(t, err, &tc.errAs)
			} else {
				require.Error(t, err)
			}
		})
	}
}
