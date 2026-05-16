package arma

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.yaml.in/yaml/v4"
)

func TestNewDataTransferRate(t *testing.T) {
	d := NewDataTransferRate(1000)
	assert.Equal(t, uint64(1000), d.Bps())
}

func TestParseDataTransferRate(t *testing.T) {
	tests := []struct {
		input   string
		wantBps uint64
	}{
		// SI bit/s shorthand
		{"5bps", 5},
		{"5kbps", 5_000},
		{"5Mbps", 5_000_000},
		{"5Gbps", 5_000_000_000},
		// bit/s slash form
		{"5bit/s", 5},
		{"5kbit/s", 5_000},
		{"5Mbit/s", 5_000_000},
		{"5Gbit/s", 5_000_000_000},
		// IEC bit/s shorthand
		{"5Kibps", 5 * 1024},
		{"5Mibps", 5 * 1024 * 1024},
		{"5Gibps", 5 * 1024 * 1024 * 1024},
		// IEC bit/s slash form
		{"5Kibit/s", 5 * 1024},
		{"5Mibit/s", 5 * 1024 * 1024},
		{"5Gibit/s", 5 * 1024 * 1024 * 1024},
		// SI byte/s
		{"5kB/s", 40_000},
		{"5MB/s", 40_000_000},
		{"5GB/s", 40_000_000_000},
		// IEC byte/s
		{"5KiB/s", 5 * 1024 * 8},
		{"5MiB/s", 5 * 1024 * 1024 * 8},
		{"5GiB/s", 5 * 1024 * 1024 * 1024 * 8},
		// decimal coefficient — rounded to nearest bps
		{"2.5Mbps", 2_500_000},
		{"0.5kbps", 500},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			d, err := ParseDataTransferRate(tt.input)
			require.NoError(t, err)
			assert.Equal(t, tt.wantBps, d.Bps())
		})
	}
}

func TestParseDataTransferRateErrors(t *testing.T) {
	tests := []string{
		"",       // empty
		"abc",    // no number, unknown unit
		"5xyz",   // unknown unit
		"Mbps",   // missing number
		"xMbps",  // non-numeric coefficient
		"5 Mbps", // space between number and unit
		"5mbps",  // wrong case (lowercase m)
		"5MBPS",  // wrong case
	}

	for _, input := range tests {
		t.Run(input, func(t *testing.T) {
			_, err := ParseDataTransferRate(input)
			assert.Error(t, err)
		})
	}
}

func TestDataTransferRateNumericAccessors(t *testing.T) {
	// 8 Gbps — divides cleanly into all units
	d := NewDataTransferRate(8_000_000_000)

	assert.Equal(t, uint64(8_000_000_000), d.Bps())
	assert.Equal(t, 8e6, d.Kbps())
	assert.Equal(t, 8e3, d.Mbps())
	assert.Equal(t, 8.0, d.Gbps())
	assert.InDelta(t, 8e9/1024, d.Kibps(), 1e-3)
	assert.InDelta(t, 8e9/(1024*1024), d.Mibps(), 1e-6)
	assert.InDelta(t, 8e9/(1024*1024*1024), d.Gibps(), 1e-9)
	assert.Equal(t, 1e6, d.KBps())
	assert.Equal(t, 1e3, d.MBps())
	assert.Equal(t, 1.0, d.GBps())
	assert.InDelta(t, 8e9/(1024*8), d.KiBps(), 1e-3)
	assert.InDelta(t, 8e9/(1024*1024*8), d.MiBps(), 1e-6)
	assert.InDelta(t, 8e9/(1024*1024*1024*8), d.GiBps(), 1e-9)
}

func TestDataTransferRateStringMethods(t *testing.T) {
	// Parse then re-format — coefficient is always 5 so output is "5<unit>" cleanly.
	tests := []struct {
		input  string
		method func(DataTransferRate) string
		want   string
	}{
		{"5bps", DataTransferRate.AsBps, "5bps"},
		{"5kbps", DataTransferRate.AsKbps, "5kbps"},
		{"5Mbps", DataTransferRate.AsMbps, "5Mbps"},
		{"5Gbps", DataTransferRate.AsGbps, "5Gbps"},
		{"5bit/s", DataTransferRate.AsBitPerS, "5bit/s"},
		{"5kbit/s", DataTransferRate.AsKbitPerS, "5kbit/s"},
		{"5Mbit/s", DataTransferRate.AsMbitPerS, "5Mbit/s"},
		{"5Gbit/s", DataTransferRate.AsGbitPerS, "5Gbit/s"},
		{"5Kibps", DataTransferRate.AsKibps, "5Kibps"},
		{"5Mibps", DataTransferRate.AsMibps, "5Mibps"},
		{"5Gibps", DataTransferRate.AsGibps, "5Gibps"},
		{"5Kibit/s", DataTransferRate.AsKibitPerS, "5Kibit/s"},
		{"5Mibit/s", DataTransferRate.AsMibitPerS, "5Mibit/s"},
		{"5Gibit/s", DataTransferRate.AsGibitPerS, "5Gibit/s"},
		{"5kB/s", DataTransferRate.AsKBPerS, "5kB/s"},
		{"5MB/s", DataTransferRate.AsMBPerS, "5MB/s"},
		{"5GB/s", DataTransferRate.AsGBPerS, "5GB/s"},
		{"5KiB/s", DataTransferRate.AsKiBPerS, "5KiB/s"},
		{"5MiB/s", DataTransferRate.AsMiBPerS, "5MiB/s"},
		{"5GiB/s", DataTransferRate.AsGiBPerS, "5GiB/s"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			d, err := ParseDataTransferRate(tt.input)
			require.NoError(t, err)
			assert.Equal(t, tt.want, tt.method(d))
		})
	}
}

func TestDataTransferRateString(t *testing.T) {
	tests := []struct {
		bps  uint64
		want string
	}{
		{500, "500bps"},
		{5_000, "5kbps"},
		{5_000_000, "5Mbps"},
		{5_000_000_000, "5Gbps"},
		{0, "0bps"},
		{999, "999bps"},
		{1_000, "1kbps"},
		{999_999, "1000kbps"},
		{1_000_000_000_000, "1000Gbps"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			assert.Equal(t, tt.want, NewDataTransferRate(tt.bps).String())
		})
	}
}

func TestDataTransferRateUnmarshalJSON(t *testing.T) {
	type wrapper struct {
		Rate DataTransferRate `json:"rate"`
	}

	t.Run("number_as_bps", func(t *testing.T) {
		var w wrapper
		require.NoError(t, json.Unmarshal([]byte(`{"rate":5000000}`), &w))
		assert.Equal(t, uint64(5_000_000), w.Rate.Bps())
	})

	t.Run("string_format", func(t *testing.T) {
		var w wrapper
		require.NoError(t, json.Unmarshal([]byte(`{"rate":"5Mbps"}`), &w))
		assert.Equal(t, uint64(5_000_000), w.Rate.Bps())
	})

	t.Run("float_bps_rounded", func(t *testing.T) {
		var w wrapper
		require.NoError(t, json.Unmarshal([]byte(`{"rate":1234.5}`), &w))
		assert.Equal(t, uint64(1235), w.Rate.Bps())
	})

	t.Run("error_bool", func(t *testing.T) {
		var w wrapper
		assert.Error(t, json.Unmarshal([]byte(`{"rate":true}`), &w))
	})

	t.Run("error_array", func(t *testing.T) {
		var w wrapper
		assert.Error(t, json.Unmarshal([]byte(`{"rate":[]}`), &w))
	})

	t.Run("error_unknown_unit", func(t *testing.T) {
		var w wrapper
		assert.Error(t, json.Unmarshal([]byte(`{"rate":"5xyz"}`), &w))
	})

	t.Run("null_as_zero", func(t *testing.T) {
		// JSON null unmarshals as 0 bps (Go stdlib behavior)
		var w wrapper
		require.NoError(t, json.Unmarshal([]byte(`{"rate":null}`), &w))
		assert.Equal(t, uint64(0), w.Rate.Bps())
	})
}

func TestDataTransferRateMarshalJSON(t *testing.T) {
	d := NewDataTransferRate(5_000_000)
	data, err := json.Marshal(d)
	require.NoError(t, err)

	var back DataTransferRate
	require.NoError(t, json.Unmarshal(data, &back))
	assert.Equal(t, d.Bps(), back.Bps())
}

// ── YAML ──────────────────────────────────────────────────────────────────────

func TestDataTransferRateMarshalYAML(t *testing.T) {
	type wrapper struct {
		Rate DataTransferRate `yaml:"rate"`
	}
	w := wrapper{Rate: NewDataTransferRate(750_000_000)}
	data, err := yaml.Marshal(w)
	require.NoError(t, err)

	var back wrapper
	require.NoError(t, yaml.Unmarshal(data, &back))
	assert.Equal(t, uint64(750_000_000), back.Rate.Bps())
}

func TestDataTransferRateUnmarshalYAML_Int(t *testing.T) {
	type wrapper struct{ Rate DataTransferRate `yaml:"rate"` }
	var w wrapper
	require.NoError(t, yaml.Unmarshal([]byte("rate: 750000000"), &w))
	assert.Equal(t, uint64(750_000_000), w.Rate.Bps())
}

func TestDataTransferRateUnmarshalYAML_String(t *testing.T) {
	type wrapper struct{ Rate DataTransferRate `yaml:"rate"` }
	var w wrapper
	require.NoError(t, yaml.Unmarshal([]byte(`rate: "750Mbps"`), &w))
	assert.Equal(t, uint64(750_000_000), w.Rate.Bps())
}

func TestDataTransferRateUnmarshalYAML_Error(t *testing.T) {
	type wrapper struct{ Rate DataTransferRate `yaml:"rate"` }
	var w wrapper
	assert.Error(t, yaml.Unmarshal([]byte("rate: [1,2,3]"), &w))
}

func TestDataTransferRateUnmarshalYAML_InvalidString(t *testing.T) {
	type wrapper struct{ Rate DataTransferRate `yaml:"rate"` }
	var w wrapper
	assert.Error(t, yaml.Unmarshal([]byte(`rate: "badvalue"`), &w))
}

// ── TOML ──────────────────────────────────────────────────────────────────────

func TestDataTransferRateMarshalTOML(t *testing.T) {
	d := NewDataTransferRate(750_000_000)
	data, err := d.MarshalTOML()
	require.NoError(t, err)
	assert.Equal(t, "750000000", string(data))
}

func TestDataTransferRateUnmarshalTOML_Int64(t *testing.T) {
	var d DataTransferRate
	require.NoError(t, d.UnmarshalTOML(int64(750_000_000)))
	assert.Equal(t, uint64(750_000_000), d.Bps())
}

func TestDataTransferRateUnmarshalTOML_Float64(t *testing.T) {
	var d DataTransferRate
	require.NoError(t, d.UnmarshalTOML(float64(750_000_000)))
	assert.Equal(t, uint64(750_000_000), d.Bps())
}

func TestDataTransferRateUnmarshalTOML_String(t *testing.T) {
	var d DataTransferRate
	require.NoError(t, d.UnmarshalTOML("750Mbps"))
	assert.Equal(t, uint64(750_000_000), d.Bps())
}

func TestDataTransferRateUnmarshalTOML_InvalidString(t *testing.T) {
	var d DataTransferRate
	assert.Error(t, d.UnmarshalTOML("bad"))
}

func TestDataTransferRateUnmarshalTOML_UnsupportedType(t *testing.T) {
	var d DataTransferRate
	assert.Error(t, d.UnmarshalTOML(true))
}
