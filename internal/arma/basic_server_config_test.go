package arma

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.yaml.in/yaml/v4"
)

func TestUnmarshal(t *testing.T) {
	t.Run("default_config", func(t *testing.T) {
		config := BasicServerConfig{}
		content := `
language: English
max_msg_send: 128
max_size_guaranteed: 512
max_size_nonguaranteed: 256
min_bandwidth: 131072
max_bandwidth: 0
min_error_to_send: 0.001
min_error_to_send_near: 0.01
max_custom_file_size: 1024
sockets:
  max_packet_size: 1400
`

		err := yaml.Load([]byte(content), &config)
		require.NoError(t, err)

		require.NotNil(t, config.Language)
		assert.Equal(t, "English", *config.Language)
		require.NotNil(t, config.MaxMsgSend)
		assert.Equal(t, uint16(128), *config.MaxMsgSend)
		require.NotNil(t, config.MaxSizeGuaranteed)
		assert.Equal(t, uint16(512), *config.MaxSizeGuaranteed)
		require.NotNil(t, config.MaxSizeNonguaranteed)
		assert.Equal(t, uint16(256), *config.MaxSizeNonguaranteed)
		require.NotNil(t, config.MinBandwidth)
		assert.Equal(t, uint64(131072), config.MinBandwidth.Bps())
		require.NotNil(t, config.MaxBandwidth)
		assert.Equal(t, uint64(0), config.MaxBandwidth.Bps())
		require.NotNil(t, config.MinErrorToSend)
		assert.Equal(t, float32(0.001), *config.MinErrorToSend)
		require.NotNil(t, config.MinErrorToSendNear)
		assert.Equal(t, float32(0.01), *config.MinErrorToSendNear)
		require.NotNil(t, config.MaxCustomFileSize)
		assert.Equal(t, uint16(1024), *config.MaxCustomFileSize)
		require.NotNil(t, config.Sockets)
		require.NotNil(t, config.Sockets.MaxPacketSize)
		assert.Equal(t, uint16(1400), *config.Sockets.MaxPacketSize)
	})

	t.Run("bandwidth_as_string", func(t *testing.T) {
		config := BasicServerConfig{}
		content := `
min_bandwidth: "5bps"
max_bandwidth: "5kbps"
`

		err := yaml.Load([]byte(content), &config)
		require.NoError(t, err)

		require.NotNil(t, config.MinBandwidth)
		assert.Equal(t, uint64(5), config.MinBandwidth.Bps())
		require.NotNil(t, config.MaxBandwidth)
		assert.Equal(t, uint64(0x1388), config.MaxBandwidth.Bps())
	})
}
