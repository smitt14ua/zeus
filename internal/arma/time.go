package arma

import (
	"encoding/json"
	"fmt"
	"math"
	"strconv"
	"strings"

	"go.yaml.in/yaml/v4"
)

// Time stores duration internally as nanoseconds.
// Accepts number (nanoseconds) or strings like "5ms", "2h", "30fps".
// "Nfps" means duration of one frame at N frames per second: 1s / N (e.g. 5fps = 200ms).
type Time struct {
	ns uint64
}

const (
	nsPerUs float64 = 1e3
	nsPerMs float64 = 1e6
	nsPerS  float64 = 1e9
	nsPerM  float64 = 60e9
	nsPerH  float64 = 3600e9
	nsPerD  float64 = 86400e9
)

// timeMultipliers ordered longest-suffix-first to avoid prefix collisions (e.g. "s" vs "ms").
// "fps" is handled separately due to its inverse (1/N) relationship.
var timeMultipliers = []struct {
	suffix     string
	multiplier float64
}{
	{"ns", 1},
	{"us", nsPerUs},
	{"ms", nsPerMs},
	{"s", nsPerS},
	{"m", nsPerM},
	{"h", nsPerH},
	{"d", nsPerD},
}

func toNs(val float64) uint64 {
	return uint64(math.Round(val))
}

// NewTime creates Time from nanoseconds.
func NewTime(ns uint64) Time {
	return Time{ns: ns}
}

// ParseTime parses strings like "5ms", "2h", "30fps".
func ParseTime(s string) (Time, error) {
	if strings.HasSuffix(s, "fps") {
		numStr := s[:len(s)-3]
		val, err := strconv.ParseFloat(numStr, 64)
		if err != nil {
			return Time{}, fmt.Errorf("invalid Time %q: %w", s, err)
		}
		if val <= 0 {
			return Time{}, fmt.Errorf("invalid Time %q: fps must be positive", s)
		}
		return Time{ns: toNs(nsPerS / val)}, nil
	}
	for _, u := range timeMultipliers {
		if strings.HasSuffix(s, u.suffix) {
			numStr := s[:len(s)-len(u.suffix)]
			val, err := strconv.ParseFloat(numStr, 64)
			if err != nil {
				return Time{}, fmt.Errorf("invalid Time %q: %w", s, err)
			}
			return Time{ns: toNs(val * u.multiplier)}, nil
		}
	}
	return Time{}, fmt.Errorf("unknown Time unit in %q", s)
}

// UnmarshalJSON handles both number (nanoseconds) and string formats.
func (d *Time) UnmarshalJSON(data []byte) error {
	var n float64
	if err := json.Unmarshal(data, &n); err == nil {
		d.ns = toNs(n)
		return nil
	}
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return fmt.Errorf("Time must be number or string, got %s", data)
	}
	r, err := ParseTime(s)
	if err != nil {
		return err
	}
	*d = r
	return nil
}

// MarshalJSON encodes as nanoseconds.
func (d Time) MarshalJSON() ([]byte, error) {
	return json.Marshal(d.ns)
}

// MarshalYAML encodes as nanoseconds.
func (d Time) MarshalYAML() (interface{}, error) {
	return d.ns, nil
}

// UnmarshalTOML handles both integer/float (nanoseconds) and string formats.
func (d *Time) UnmarshalTOML(v interface{}) error {
	switch val := v.(type) {
	case int64:
		d.ns = uint64(val)
		return nil
	case float64:
		d.ns = toNs(val)
		return nil
	case string:
		r, err := ParseTime(val)
		if err != nil {
			return err
		}
		*d = r
		return nil
	default:
		return fmt.Errorf("Time must be number or string, got %T", v)
	}
}

// UnmarshalYAML handles both integer/float (nanoseconds) and string formats.
func (d *Time) UnmarshalYAML(node *yaml.Node) error {
	switch node.Tag {
	case "!!int", "!!float":
		val, err := strconv.ParseFloat(node.Value, 64)
		if err != nil {
			return fmt.Errorf("invalid Time %q: %w", node.Value, err)
		}
		d.ns = toNs(val)
		return nil
	case "!!str":
		r, err := ParseTime(node.Value)
		if err != nil {
			return err
		}
		*d = r
		return nil
	default:
		return fmt.Errorf("Time must be number or string, got tag %s", node.Tag)
	}
}

// Numeric accessors.
func (d Time) Nanoseconds() uint64   { return d.ns }
func (d Time) Microseconds() float64 { return float64(d.ns) / nsPerUs }
func (d Time) Milliseconds() float64 { return float64(d.ns) / nsPerMs }
func (d Time) Seconds() float64      { return float64(d.ns) / nsPerS }
func (d Time) Minutes() float64      { return float64(d.ns) / nsPerM }
func (d Time) Hours() float64        { return float64(d.ns) / nsPerH }
func (d Time) Days() float64         { return float64(d.ns) / nsPerD }

// Fps returns the frame rate whose frame duration equals this Time (1s / ns).
// Returns 0 if ns is 0.
func (d Time) Fps() float64 {
	if d.ns == 0 {
		return 0
	}
	return nsPerS / float64(d.ns)
}

func fmtTime(val float64, unit string) string {
	return fmt.Sprintf("%d%s", int64(math.Round(val)), unit)
}

// String representations — each matches a supported input format. Values are rounded to integers.
func (d Time) AsNs() string  { return fmt.Sprintf("%dns", d.Nanoseconds()) }
func (d Time) AsUs() string  { return fmtTime(d.Microseconds(), "us") }
func (d Time) AsMs() string  { return fmtTime(d.Milliseconds(), "ms") }
func (d Time) AsS() string   { return fmtTime(d.Seconds(), "s") }
func (d Time) AsM() string   { return fmtTime(d.Minutes(), "m") }
func (d Time) AsH() string   { return fmtTime(d.Hours(), "h") }
func (d Time) AsD() string   { return fmtTime(d.Days(), "d") }
func (d Time) AsFps() string { return fmtTime(d.Fps(), "fps") }

// String returns human-readable form in the largest applicable unit.
func (d Time) String() string {
	switch {
	case d.ns >= uint64(nsPerD):
		return d.AsD()
	case d.ns >= uint64(nsPerH):
		return d.AsH()
	case d.ns >= uint64(nsPerM):
		return d.AsM()
	case d.ns >= uint64(nsPerS):
		return d.AsS()
	case d.ns >= uint64(nsPerMs):
		return d.AsMs()
	case d.ns >= uint64(nsPerUs):
		return d.AsUs()
	default:
		return d.AsNs()
	}
}
