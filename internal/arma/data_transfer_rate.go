package arma

import (
	"encoding/json"
	"fmt"
	"math"
	"strconv"
	"strings"

	"go.yaml.in/yaml/v4"
)

// DataTransferRate stores bit rate internally as bits per second.
// Accepts number (bps) or strings like "5Mbps", "10kB/s", "100Mibit/s".
type DataTransferRate struct {
	bps uint64
}

const (
	kilo float64 = 1e3
	mega float64 = 1e6
	giga float64 = 1e9
	tera float64 = 1e12
	kibi float64 = 1024
	mebi float64 = 1024 * 1024
	gibi float64 = 1024 * 1024 * 1024
	tebi float64 = 1024 * 1024 * 1024 * 1024
)

// unitMultipliers ordered longest-suffix-first to avoid prefix collisions (e.g. "bps" vs "kbps").
var unitMultipliers = []struct {
	suffix     string
	multiplier float64
}{
	{"Gibit/s", gibi},
	{"Mibit/s", mebi},
	{"Kibit/s", kibi},
	{"Gbit/s", giga},
	{"Mbit/s", mega},
	{"kbit/s", kilo},
	{"bit/s", 1},
	{"GiB/s", gibi * 8},
	{"MiB/s", mebi * 8},
	{"KiB/s", kibi * 8},
	{"GB/s", giga * 8},
	{"MB/s", mega * 8},
	{"kB/s", kilo * 8},
	{"Gibps", gibi},
	{"Mibps", mebi},
	{"Kibps", kibi},
	{"Gbps", giga},
	{"Mbps", mega},
	{"kbps", kilo},
	{"bps", 1},
}

func toBps(val float64) uint64 {
	return uint64(math.Round(val))
}

// NewDataTransferRate creates DataTransferRate from bits per second.
func NewDataTransferRate(bps uint64) DataTransferRate {
	return DataTransferRate{bps: bps}
}

// ParseDataTransferRate parses strings like "5Mbps", "10kB/s", "100Mibit/s".
func ParseDataTransferRate(s string) (DataTransferRate, error) {
	for _, u := range unitMultipliers {
		if strings.HasSuffix(s, u.suffix) {
			numStr := s[:len(s)-len(u.suffix)]
			val, err := strconv.ParseFloat(numStr, 64)
			if err != nil {
				return DataTransferRate{}, fmt.Errorf("invalid DataTransferRate %q: %w", s, err)
			}
			return DataTransferRate{bps: toBps(val * u.multiplier)}, nil
		}
	}
	return DataTransferRate{}, fmt.Errorf("unknown DataTransferRate unit in %q", s)
}

// UnmarshalJSON handles both number (bps) and string formats.
func (d *DataTransferRate) UnmarshalJSON(data []byte) error {
	var n float64
	if err := json.Unmarshal(data, &n); err == nil {
		d.bps = toBps(n)
		return nil
	}
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return fmt.Errorf("DataTransferRate must be number or string, got %s", data)
	}
	r, err := ParseDataTransferRate(s)
	if err != nil {
		return err
	}
	*d = r
	return nil
}

// MarshalJSON encodes as bits per second.
func (d DataTransferRate) MarshalJSON() ([]byte, error) {
	return json.Marshal(d.bps)
}

// MarshalYAML encodes as bits per second.
func (d DataTransferRate) MarshalYAML() (interface{}, error) {
	return d.bps, nil
}

// MarshalTOML encodes as bits per second.
func (d DataTransferRate) MarshalTOML() ([]byte, error) {
	return []byte(strconv.FormatUint(d.bps, 10)), nil
}

// UnmarshalTOML handles both integer/float (bps) and string formats.
func (d *DataTransferRate) UnmarshalTOML(v interface{}) error {
	switch val := v.(type) {
	case int64:
		d.bps = uint64(val)
		return nil
	case float64:
		d.bps = toBps(val)
		return nil
	case string:
		r, err := ParseDataTransferRate(val)
		if err != nil {
			return err
		}
		*d = r
		return nil
	default:
		return fmt.Errorf("DataTransferRate must be number or string, got %T", v)
	}
}

// UnmarshalYAML handles both integer/float (bps) and string formats.
func (d *DataTransferRate) UnmarshalYAML(node *yaml.Node) error {
	switch node.Tag {
	case "!!int", "!!float":
		val, err := strconv.ParseFloat(node.Value, 64)
		if err != nil {
			return fmt.Errorf("invalid DataTransferRate %q: %w", node.Value, err)
		}
		d.bps = toBps(val)
		return nil
	case "!!str":
		r, err := ParseDataTransferRate(node.Value)
		if err != nil {
			return err
		}
		*d = r
		return nil
	default:
		return fmt.Errorf("DataTransferRate must be number or string, got tag %s", node.Tag)
	}
}

// Numeric accessors — SI bit/s units.
func (d DataTransferRate) Bps() uint64    { return d.bps }
func (d DataTransferRate) Kbps() float64  { return float64(d.bps) / kilo }
func (d DataTransferRate) Mbps() float64  { return float64(d.bps) / mega }
func (d DataTransferRate) Gbps() float64  { return float64(d.bps) / giga }
func (d DataTransferRate) Kibps() float64 { return float64(d.bps) / kibi }
func (d DataTransferRate) Mibps() float64 { return float64(d.bps) / mebi }
func (d DataTransferRate) Gibps() float64 { return float64(d.bps) / gibi }

// Numeric accessors — byte/s units.
func (d DataTransferRate) KBps() float64  { return float64(d.bps) / (kilo * 8) }
func (d DataTransferRate) MBps() float64  { return float64(d.bps) / (mega * 8) }
func (d DataTransferRate) GBps() float64  { return float64(d.bps) / (giga * 8) }
func (d DataTransferRate) KiBps() float64 { return float64(d.bps) / (kibi * 8) }
func (d DataTransferRate) MiBps() float64 { return float64(d.bps) / (mebi * 8) }
func (d DataTransferRate) GiBps() float64 { return float64(d.bps) / (gibi * 8) }

func fmtRate(val float64, unit string) string {
	return fmt.Sprintf("%d%s", int64(math.Round(val)), unit)
}

// String representations — each matches a supported input format. Values are rounded to integers.
func (d DataTransferRate) AsBps() string       { return fmt.Sprintf("%dbps", d.Bps()) }
func (d DataTransferRate) AsKbps() string      { return fmtRate(d.Kbps(), "kbps") }
func (d DataTransferRate) AsMbps() string      { return fmtRate(d.Mbps(), "Mbps") }
func (d DataTransferRate) AsGbps() string      { return fmtRate(d.Gbps(), "Gbps") }
func (d DataTransferRate) AsBitPerS() string   { return fmt.Sprintf("%dbit/s", d.Bps()) }
func (d DataTransferRate) AsKbitPerS() string  { return fmtRate(d.Kbps(), "kbit/s") }
func (d DataTransferRate) AsMbitPerS() string  { return fmtRate(d.Mbps(), "Mbit/s") }
func (d DataTransferRate) AsGbitPerS() string  { return fmtRate(d.Gbps(), "Gbit/s") }
func (d DataTransferRate) AsKibps() string     { return fmtRate(d.Kibps(), "Kibps") }
func (d DataTransferRate) AsMibps() string     { return fmtRate(d.Mibps(), "Mibps") }
func (d DataTransferRate) AsGibps() string     { return fmtRate(d.Gibps(), "Gibps") }
func (d DataTransferRate) AsKibitPerS() string { return fmtRate(d.Kibps(), "Kibit/s") }
func (d DataTransferRate) AsMibitPerS() string { return fmtRate(d.Mibps(), "Mibit/s") }
func (d DataTransferRate) AsGibitPerS() string { return fmtRate(d.Gibps(), "Gibit/s") }
func (d DataTransferRate) AsKBPerS() string    { return fmtRate(d.KBps(), "kB/s") }
func (d DataTransferRate) AsMBPerS() string    { return fmtRate(d.MBps(), "MB/s") }
func (d DataTransferRate) AsGBPerS() string    { return fmtRate(d.GBps(), "GB/s") }
func (d DataTransferRate) AsKiBPerS() string   { return fmtRate(d.KiBps(), "KiB/s") }
func (d DataTransferRate) AsMiBPerS() string   { return fmtRate(d.MiBps(), "MiB/s") }
func (d DataTransferRate) AsGiBPerS() string   { return fmtRate(d.GiBps(), "GiB/s") }

// String returns human-readable form in the largest applicable SI unit.
func (d DataTransferRate) String() string {
	switch {
	case d.bps >= uint64(giga):
		return d.AsGbps()
	case d.bps >= uint64(mega):
		return d.AsMbps()
	case d.bps >= uint64(kilo):
		return d.AsKbps()
	default:
		return d.AsBps()
	}
}
