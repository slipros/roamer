package formatter

import (
	"reflect"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func createTimeTestTag(value string) reflect.StructTag {
	return reflect.StructTag(`time:"` + value + `"`)
}

func TestTime_Tag(t *testing.T) {
	f := NewTime()
	assert.Equal(t, "time", f.Tag())
}

func TestTime_Format_Successfully(t *testing.T) {
	f := NewTime()

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
	f := NewTime()

	tests := []struct {
		name    string
		tag     reflect.StructTag
		input   any
		wantErr string
	}{
		{name: "Unsupported type", tag: createTimeTestTag("timezone=UTC"), input: new(int), wantErr: "not supported"},
		{name: "Invalid timezone", tag: createTimeTestTag("timezone=Invalid/Timezone"), input: new(time.Time), wantErr: "invalid timezone"},
		{name: "Invalid truncate duration", tag: createTimeTestTag("truncate=invalid"), input: new(time.Time), wantErr: "invalid duration"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := f.Format(tc.tag, tc.input)
			require.Error(t, err)
			assert.Contains(t, err.Error(), tc.wantErr)
		})
	}
}
