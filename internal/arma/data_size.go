package arma

import (
	"encoding/json"
	"fmt"
	"math"
	"strconv"
	"strings"

	"go.yaml.in/yaml/v4"
)

// DataSize stores size internally as bytes.
// Accepts number (bytes) or strings like "5MB", "10KiB".
type DataSize struct {
	bytes uint64
}

// sizeMultipliers ordered longest-suffix-first to avoid prefix collisions (e.g. "B" vs "KB").
var sizeMultipliers = []struct {
	suffix     string
	multiplier float64
}{
	{"TiB", tebi},
	{"GiB", gibi},
	{"MiB", mebi},
	{"KiB", kibi},
	{"TB", tera},
	{"GB", giga},
	{"MB", mega},
	{"KB", kilo},
	{"B", 1},
}

func toBytes(val float64) uint64 {
	return uint64(math.Round(val))
}

// NewDataSize creates DataSize from bytes.
func NewDataSize(bytes uint64) DataSize {
	return DataSize{bytes: bytes}
}

// ParseDataSize parses strings like "5MB", "10KiB", "2TB".
func ParseDataSize(s string) (DataSize, error) {
	for _, u := range sizeMultipliers {
		if strings.HasSuffix(s, u.suffix) {
			numStr := s[:len(s)-len(u.suffix)]
			val, err := strconv.ParseFloat(numStr, 64)
			if err != nil {
				return DataSize{}, fmt.Errorf("invalid DataSize %q: %w", s, err)
			}
			return DataSize{bytes: toBytes(val * u.multiplier)}, nil
		}
	}
	return DataSize{}, fmt.Errorf("unknown DataSize unit in %q", s)
}

// UnmarshalJSON handles both number (bytes) and string formats.
func (d *DataSize) UnmarshalJSON(data []byte) error {
	var n float64
	if err := json.Unmarshal(data, &n); err == nil {
		d.bytes = toBytes(n)
		return nil
	}
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return fmt.Errorf("DataSize must be number or string, got %s", data)
	}
	r, err := ParseDataSize(s)
	if err != nil {
		return err
	}
	*d = r
	return nil
}

// MarshalJSON encodes as bytes.
func (d DataSize) MarshalJSON() ([]byte, error) {
	return json.Marshal(d.bytes)
}

// MarshalYAML encodes as bytes.
func (d DataSize) MarshalYAML() (interface{}, error) {
	return d.bytes, nil
}

// MarshalTOML encodes as bytes.
func (d DataSize) MarshalTOML() ([]byte, error) {
	return []byte(strconv.FormatUint(d.bytes, 10)), nil
}

// UnmarshalTOML handles both integer/float (bytes) and string formats.
func (d *DataSize) UnmarshalTOML(v interface{}) error {
	switch val := v.(type) {
	case int64:
		d.bytes = uint64(val)
		return nil
	case float64:
		d.bytes = toBytes(val)
		return nil
	case string:
		r, err := ParseDataSize(val)
		if err != nil {
			return err
		}
		*d = r
		return nil
	default:
		return fmt.Errorf("DataSize must be number or string, got %T", v)
	}
}

// UnmarshalYAML handles both integer/float (bytes) and string formats.
func (d *DataSize) UnmarshalYAML(node *yaml.Node) error {
	switch node.Tag {
	case "!!int", "!!float":
		val, err := strconv.ParseFloat(node.Value, 64)
		if err != nil {
			return fmt.Errorf("invalid DataSize %q: %w", node.Value, err)
		}
		d.bytes = toBytes(val)
		return nil
	case "!!str":
		r, err := ParseDataSize(node.Value)
		if err != nil {
			return err
		}
		*d = r
		return nil
	default:
		return fmt.Errorf("DataSize must be number or string, got tag %s", node.Tag)
	}
}

// Numeric accessors — SI byte units.
func (d DataSize) B() uint64   { return d.bytes }
func (d DataSize) KB() float64 { return float64(d.bytes) / kilo }
func (d DataSize) MB() float64 { return float64(d.bytes) / mega }
func (d DataSize) GB() float64 { return float64(d.bytes) / giga }
func (d DataSize) TB() float64 { return float64(d.bytes) / tera }

// Numeric accessors — IEC byte units.
func (d DataSize) KiB() float64 { return float64(d.bytes) / kibi }
func (d DataSize) MiB() float64 { return float64(d.bytes) / mebi }
func (d DataSize) GiB() float64 { return float64(d.bytes) / gibi }
func (d DataSize) TiB() float64 { return float64(d.bytes) / tebi }

func fmtSize(val float64, unit string) string {
	return fmt.Sprintf("%d%s", int64(math.Round(val)), unit)
}

// String representations — each matches a supported input format. Values are rounded to integers.
func (d DataSize) AsB() string   { return fmt.Sprintf("%dB", d.B()) }
func (d DataSize) AsKB() string  { return fmtSize(d.KB(), "KB") }
func (d DataSize) AsMB() string  { return fmtSize(d.MB(), "MB") }
func (d DataSize) AsGB() string  { return fmtSize(d.GB(), "GB") }
func (d DataSize) AsTB() string  { return fmtSize(d.TB(), "TB") }
func (d DataSize) AsKiB() string { return fmtSize(d.KiB(), "KiB") }
func (d DataSize) AsMiB() string { return fmtSize(d.MiB(), "MiB") }
func (d DataSize) AsGiB() string { return fmtSize(d.GiB(), "GiB") }
func (d DataSize) AsTiB() string { return fmtSize(d.TiB(), "TiB") }

// String returns human-readable form in the largest applicable SI unit.
func (d DataSize) String() string {
	switch {
	case d.bytes >= uint64(tera):
		return d.AsTB()
	case d.bytes >= uint64(giga):
		return d.AsGB()
	case d.bytes >= uint64(mega):
		return d.AsMB()
	case d.bytes >= uint64(kilo):
		return d.AsKB()
	default:
		return d.AsB()
	}
}
