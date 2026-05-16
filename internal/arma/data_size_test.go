package arma

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.yaml.in/yaml/v4"
)

func TestNewDataSize(t *testing.T) {
	d := NewDataSize(1024)
	assert.Equal(t, uint64(1024), d.B())
}

func TestParseDataSize(t *testing.T) {
	tests := []struct {
		input     string
		wantBytes uint64
	}{
		// SI units
		{"5B", 5},
		{"5KB", 5_000},
		{"5MB", 5_000_000},
		{"5GB", 5_000_000_000},
		{"5TB", 5_000_000_000_000},
		// IEC units
		{"5KiB", 5 * 1024},
		{"5MiB", 5 * 1024 * 1024},
		{"5GiB", 5 * 1024 * 1024 * 1024},
		{"5TiB", 5 * 1024 * 1024 * 1024 * 1024},
		// decimal coefficient
		{"2.5MB", 2_500_000},
		{"0.5KB", 500},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			d, err := ParseDataSize(tt.input)
			require.NoError(t, err)
			assert.Equal(t, tt.wantBytes, d.B())
		})
	}
}

func TestParseDataSizeErrors(t *testing.T) {
	tests := []string{
		"",     // empty
		"abc",  // unknown unit
		"5xyz", // unknown unit
		"MB",   // missing number
		"xMB",  // non-numeric coefficient
		"5 MB", // space between number and unit
		"5mb",  // wrong case
		"5gb",  // wrong case
		"5kib", // wrong case
	}

	for _, input := range tests {
		t.Run(input, func(t *testing.T) {
			_, err := ParseDataSize(input)
			assert.Error(t, err)
		})
	}
}

func TestDataSizeNumericAccessors(t *testing.T) {
	// 8 TB — divides cleanly into all SI units
	d := NewDataSize(8_000_000_000_000)

	assert.Equal(t, uint64(8_000_000_000_000), d.B())
	assert.Equal(t, 8e9, d.KB())
	assert.Equal(t, 8e6, d.MB())
	assert.Equal(t, 8e3, d.GB())
	assert.Equal(t, 8.0, d.TB())
	assert.InDelta(t, 8e12/1024, d.KiB(), 1e-3)
	assert.InDelta(t, 8e12/(1024*1024), d.MiB(), 1e-6)
	assert.InDelta(t, 8e12/(1024*1024*1024), d.GiB(), 1e-9)
	assert.InDelta(t, 8e12/(1024*1024*1024*1024), d.TiB(), 1e-12)
}

func TestDataSizeStringMethods(t *testing.T) {
	// Parse then re-format — coefficient is 5 so output is "5<unit>" cleanly.
	tests := []struct {
		input  string
		method func(DataSize) string
		want   string
	}{
		{"5B", DataSize.AsB, "5B"},
		{"5KB", DataSize.AsKB, "5KB"},
		{"5MB", DataSize.AsMB, "5MB"},
		{"5GB", DataSize.AsGB, "5GB"},
		{"5TB", DataSize.AsTB, "5TB"},
		{"5KiB", DataSize.AsKiB, "5KiB"},
		{"5MiB", DataSize.AsMiB, "5MiB"},
		{"5GiB", DataSize.AsGiB, "5GiB"},
		{"5TiB", DataSize.AsTiB, "5TiB"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			d, err := ParseDataSize(tt.input)
			require.NoError(t, err)
			assert.Equal(t, tt.want, tt.method(d))
		})
	}
}

func TestDataSizeString(t *testing.T) {
	tests := []struct {
		bytes uint64
		want  string
	}{
		{500, "500B"},
		{5_000, "5KB"},
		{5_000_000, "5MB"},
		{5_000_000_000, "5GB"},
		{5_000_000_000_000, "5TB"},
		{0, "0B"},
		{999, "999B"},
		{1_000, "1KB"},
		{999_999, "1000KB"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			assert.Equal(t, tt.want, NewDataSize(tt.bytes).String())
		})
	}
}

func TestDataSizeUnmarshalJSON(t *testing.T) {
	type wrapper struct {
		Size DataSize `json:"size"`
	}

	t.Run("number_as_bytes", func(t *testing.T) {
		var w wrapper
		require.NoError(t, json.Unmarshal([]byte(`{"size":5000000}`), &w))
		assert.Equal(t, uint64(5_000_000), w.Size.B())
	})

	t.Run("string_format", func(t *testing.T) {
		var w wrapper
		require.NoError(t, json.Unmarshal([]byte(`{"size":"5MB"}`), &w))
		assert.Equal(t, uint64(5_000_000), w.Size.B())
	})

	t.Run("float_bytes_rounded", func(t *testing.T) {
		var w wrapper
		require.NoError(t, json.Unmarshal([]byte(`{"size":1234.5}`), &w))
		assert.Equal(t, uint64(1235), w.Size.B())
	})

	t.Run("error_bool", func(t *testing.T) {
		var w wrapper
		assert.Error(t, json.Unmarshal([]byte(`{"size":true}`), &w))
	})

	t.Run("error_array", func(t *testing.T) {
		var w wrapper
		assert.Error(t, json.Unmarshal([]byte(`{"size":[]}`), &w))
	})

	t.Run("error_unknown_unit", func(t *testing.T) {
		var w wrapper
		assert.Error(t, json.Unmarshal([]byte(`{"size":"5xyz"}`), &w))
	})

	t.Run("null_as_zero", func(t *testing.T) {
		var w wrapper
		require.NoError(t, json.Unmarshal([]byte(`{"size":null}`), &w))
		assert.Equal(t, uint64(0), w.Size.B())
	})
}

func TestDataSizeMarshalJSON(t *testing.T) {
	d := NewDataSize(5_000_000)
	data, err := json.Marshal(d)
	require.NoError(t, err)

	var back DataSize
	require.NoError(t, json.Unmarshal(data, &back))
	assert.Equal(t, d.B(), back.B())
}

// ── YAML ──────────────────────────────────────────────────────────────────────

func TestDataSizeMarshalYAML(t *testing.T) {
	type wrapper struct {
		Size DataSize `yaml:"size"`
	}
	w := wrapper{Size: NewDataSize(5_000_000)}
	data, err := yaml.Marshal(w)
	require.NoError(t, err)

	var back wrapper
	require.NoError(t, yaml.Unmarshal(data, &back))
	assert.Equal(t, uint64(5_000_000), back.Size.B())
}

func TestDataSizeUnmarshalYAML_Int(t *testing.T) {
	type wrapper struct{ Size DataSize `yaml:"size"` }
	var w wrapper
	require.NoError(t, yaml.Unmarshal([]byte("size: 5000000"), &w))
	assert.Equal(t, uint64(5_000_000), w.Size.B())
}

func TestDataSizeUnmarshalYAML_String(t *testing.T) {
	type wrapper struct{ Size DataSize `yaml:"size"` }
	var w wrapper
	require.NoError(t, yaml.Unmarshal([]byte(`size: "5MB"`), &w))
	assert.Equal(t, uint64(5_000_000), w.Size.B())
}

func TestDataSizeUnmarshalYAML_Error(t *testing.T) {
	type wrapper struct{ Size DataSize `yaml:"size"` }
	var w wrapper
	assert.Error(t, yaml.Unmarshal([]byte("size: [1,2,3]"), &w))
}

func TestDataSizeUnmarshalYAML_InvalidString(t *testing.T) {
	type wrapper struct{ Size DataSize `yaml:"size"` }
	var w wrapper
	assert.Error(t, yaml.Unmarshal([]byte(`size: "badvalue"`), &w))
}

// ── TOML ──────────────────────────────────────────────────────────────────────

func TestDataSizeMarshalTOML(t *testing.T) {
	d := NewDataSize(5_000_000)
	data, err := d.MarshalTOML()
	require.NoError(t, err)
	assert.Equal(t, "5000000", string(data))
}

func TestDataSizeUnmarshalTOML_Int64(t *testing.T) {
	var d DataSize
	require.NoError(t, d.UnmarshalTOML(int64(5_000_000)))
	assert.Equal(t, uint64(5_000_000), d.B())
}

func TestDataSizeUnmarshalTOML_Float64(t *testing.T) {
	var d DataSize
	require.NoError(t, d.UnmarshalTOML(float64(5_000_000)))
	assert.Equal(t, uint64(5_000_000), d.B())
}

func TestDataSizeUnmarshalTOML_String(t *testing.T) {
	var d DataSize
	require.NoError(t, d.UnmarshalTOML("5MB"))
	assert.Equal(t, uint64(5_000_000), d.B())
}

func TestDataSizeUnmarshalTOML_InvalidString(t *testing.T) {
	var d DataSize
	assert.Error(t, d.UnmarshalTOML("bad"))
}

func TestDataSizeUnmarshalTOML_UnsupportedType(t *testing.T) {
	var d DataSize
	assert.Error(t, d.UnmarshalTOML(true))
}
