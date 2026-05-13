package arma

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewTime(t *testing.T) {
	d := NewTime(5_000_000)
	assert.Equal(t, uint64(5_000_000), d.Nanoseconds())
}

func TestParseTime(t *testing.T) {
	tests := []struct {
		input  string
		wantNs uint64
	}{
		{"5ns", 5},
		{"5us", 5_000},
		{"5ms", 5_000_000},
		{"5s", 5_000_000_000},
		{"5m", 300_000_000_000},
		{"5h", 18_000_000_000_000},
		{"5d", 432_000_000_000_000},
		// fps: 5fps = 1s/5 = 200ms = 200_000_000 ns
		{"5fps", 200_000_000},
		// 10fps = 1s/10 = 100ms = 100_000_000 ns
		{"10fps", 100_000_000},
		// decimal coefficients — stored as rounded nanoseconds, output always integer
		{"2.5ms", 2_500_000},
		{"0.5s", 500_000_000},
		{"0.5m", 30_000_000_000},     // 0.5 × 60s = 30s
		{"1.5h", 5_400_000_000_000},  // 1.5 × 3600s = 5400s
		{"0.5d", 43_200_000_000_000}, // 0.5 × 86400s = 43200s = 12h
		{"0.5fps", 2_000_000_000},    // 1s / 0.5 = 2s per frame
		{"2.5fps", 400_000_000},      // 1s / 2.5 = 400ms per frame
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			d, err := ParseTime(tt.input)
			require.NoError(t, err)
			assert.Equal(t, tt.wantNs, d.Nanoseconds())
		})
	}
}

func TestParseTimeErrors(t *testing.T) {
	tests := []string{
		"",      // empty
		"abc",   // unknown unit
		"5xyz",  // unknown unit
		"ms",    // missing number
		"xms",   // non-numeric coefficient
		"5 ms",  // space between number and unit
		"5S",    // wrong case
		"5MS",   // wrong case
		"0fps",  // zero fps = infinite duration, invalid
		"-1fps", // negative fps, invalid
	}

	for _, input := range tests {
		t.Run(input, func(t *testing.T) {
			_, err := ParseTime(input)
			assert.Error(t, err)
		})
	}
}

func TestTimeNumericAccessors(t *testing.T) {
	// 1 day = 86_400_000_000_000 ns — divides cleanly into all units
	d := NewTime(86_400_000_000_000)

	assert.Equal(t, uint64(86_400_000_000_000), d.Nanoseconds())
	assert.Equal(t, 86_400_000_000.0, d.Microseconds())
	assert.Equal(t, 86_400_000.0, d.Milliseconds())
	assert.Equal(t, 86_400.0, d.Seconds())
	assert.Equal(t, 1440.0, d.Minutes())
	assert.Equal(t, 24.0, d.Hours())
	assert.Equal(t, 1.0, d.Days())
}

func TestTimeFpsAccessor(t *testing.T) {
	assert.Equal(t, 5.0, NewTime(200_000_000).Fps())  // 200ms = 5fps
	assert.Equal(t, 10.0, NewTime(100_000_000).Fps()) // 100ms = 10fps
	assert.Equal(t, 0.0, NewTime(0).Fps())            // zero guard
}

func TestTimeStringMethods(t *testing.T) {
	// Parse then re-format — coefficient is 5 so output is "5<unit>" cleanly.
	tests := []struct {
		input  string
		method func(Time) string
		want   string
	}{
		{"5ns", Time.AsNs, "5ns"},
		{"5us", Time.AsUs, "5us"},
		{"5ms", Time.AsMs, "5ms"},
		{"5s", Time.AsS, "5s"},
		{"5m", Time.AsM, "5m"},
		{"5h", Time.AsH, "5h"},
		{"5d", Time.AsD, "5d"},
		{"5fps", Time.AsFps, "5fps"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			d, err := ParseTime(tt.input)
			require.NoError(t, err)
			assert.Equal(t, tt.want, tt.method(d))
		})
	}
}

func TestTimeString(t *testing.T) {
	tests := []struct {
		ns   uint64
		want string
	}{
		{0, "0ns"},
		{500, "500ns"},
		{5_000, "5us"},
		{5_000_000, "5ms"},
		{5_000_000_000, "5s"},
		{300_000_000_000, "5m"},
		{18_000_000_000_000, "5h"},
		{432_000_000_000_000, "5d"},
		{999, "999ns"},
		{1_000, "1us"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			assert.Equal(t, tt.want, NewTime(tt.ns).String())
		})
	}
}

func TestTimeUnmarshalJSON(t *testing.T) {
	type wrapper struct {
		T Time `json:"t"`
	}

	t.Run("number_as_ns", func(t *testing.T) {
		var w wrapper
		require.NoError(t, json.Unmarshal([]byte(`{"t":5000000}`), &w))
		assert.Equal(t, uint64(5_000_000), w.T.Nanoseconds())
	})

	t.Run("string_format", func(t *testing.T) {
		var w wrapper
		require.NoError(t, json.Unmarshal([]byte(`{"t":"5ms"}`), &w))
		assert.Equal(t, uint64(5_000_000), w.T.Nanoseconds())
	})

	t.Run("string_fps", func(t *testing.T) {
		var w wrapper
		require.NoError(t, json.Unmarshal([]byte(`{"t":"5fps"}`), &w))
		assert.Equal(t, uint64(200_000_000), w.T.Nanoseconds())
	})

	t.Run("float_ns_rounded", func(t *testing.T) {
		var w wrapper
		require.NoError(t, json.Unmarshal([]byte(`{"t":1234.5}`), &w))
		assert.Equal(t, uint64(1235), w.T.Nanoseconds())
	})

	t.Run("error_bool", func(t *testing.T) {
		var w wrapper
		assert.Error(t, json.Unmarshal([]byte(`{"t":true}`), &w))
	})

	t.Run("error_array", func(t *testing.T) {
		var w wrapper
		assert.Error(t, json.Unmarshal([]byte(`{"t":[]}`), &w))
	})

	t.Run("error_unknown_unit", func(t *testing.T) {
		var w wrapper
		assert.Error(t, json.Unmarshal([]byte(`{"t":"5xyz"}`), &w))
	})

	t.Run("null_as_zero", func(t *testing.T) {
		var w wrapper
		require.NoError(t, json.Unmarshal([]byte(`{"t":null}`), &w))
		assert.Equal(t, uint64(0), w.T.Nanoseconds())
	})
}

func TestTimeMarshalJSON(t *testing.T) {
	d := NewTime(5_000_000)
	data, err := json.Marshal(d)
	require.NoError(t, err)

	var back Time
	require.NoError(t, json.Unmarshal(data, &back))
	assert.Equal(t, d.Nanoseconds(), back.Nanoseconds())
}
