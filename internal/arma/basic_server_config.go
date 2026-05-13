package arma

// BasicServerConfig holds all Arma 3 basic.cfg parameters.
// All fields are optional (pointer); nil means the parameter is omitted from the cfg file.
type BasicServerConfig struct {
	Language             *string           `json:"language,omitempty"             yaml:"language,omitempty"               toml:"language,omitempty"`
	MaxMsgSend           *uint16           `json:"maxMsgSend,omitempty"           yaml:"max_msg_send,omitempty"           toml:"max_msg_send,omitempty"`
	MaxSizeGuaranteed    *uint16           `json:"maxSizeGuaranteed,omitempty"    yaml:"max_size_guaranteed,omitempty"    toml:"max_size_guaranteed,omitempty"`
	MaxSizeNonguaranteed *uint16           `json:"maxSizeNonguaranteed,omitempty" yaml:"max_size_nonguaranteed,omitempty" toml:"max_size_nonguaranteed,omitempty"`
	MinBandwidth         *DataTransferRate `json:"minBandwidth,omitempty"         yaml:"min_bandwidth,omitempty"          toml:"min_bandwidth,omitempty"`
	MaxBandwidth         *DataTransferRate `json:"maxBandwidth,omitempty"         yaml:"max_bandwidth,omitempty"          toml:"max_bandwidth,omitempty"`
	MinErrorToSend       *float32          `json:"minErrorToSend,omitempty"       yaml:"min_error_to_send,omitempty"      toml:"min_error_to_send,omitempty"`
	MinErrorToSendNear   *float32          `json:"minErrorToSendNear,omitempty"   yaml:"min_error_to_send_near,omitempty" toml:"min_error_to_send_near,omitempty"`
	MaxCustomFileSize    *uint16           `json:"maxCustomFileSize,omitempty"    yaml:"max_custom_file_size,omitempty"   toml:"max_custom_file_size,omitempty"`
	Sockets              *Sockets          `json:"sockets,omitempty"              yaml:"sockets,omitempty"                toml:"sockets,omitempty"`
}

type Sockets struct {
	MaxPacketSize *uint16 `json:"maxPacketSize,omitempty" yaml:"max_packet_size,omitempty" toml:"max_packet_size,omitempty"`
}

// NewDefaultBasicServerConfig returns BasicServerConfig populated with documented Arma 3 defaults.
func NewDefaultBasicServerConfig() BasicServerConfig {
	return BasicServerConfig{
		Language:             ptr("English"),
		MaxMsgSend:           ptr(uint16(128)),
		MaxSizeGuaranteed:    ptr(uint16(512)),
		MaxSizeNonguaranteed: ptr(uint16(256)),
		MaxBandwidth:         ptr(NewDataTransferRate(100_000_000)),
		MinErrorToSend:       ptr(float32(0.001)),
		MinErrorToSendNear:   ptr(float32(0.01)),
		MaxCustomFileSize:    ptr(uint16(1024)),
		Sockets: &Sockets{
			MaxPacketSize: ptr(uint16(1400)),
		},
	}
}
