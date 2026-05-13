package arma

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.yaml.in/yaml/v4"
)

func boolPtr(v bool) *bool       { return &v }
func uint16Ptr(v uint16) *uint16 { return &v }

func TestStartupParamsAllNilByDefault(t *testing.T) {
	var p StartupParams
	assert.Nil(t, p.World)
	assert.Nil(t, p.WorldCfg)
	assert.Nil(t, p.Profiles)
	assert.Nil(t, p.DoNothing)
	assert.Nil(t, p.Mod)
	assert.Nil(t, p.MpMissions)
	assert.Nil(t, p.Server)
	assert.Nil(t, p.Port)
	assert.Nil(t, p.Pid)
	assert.Nil(t, p.Ranking)
	assert.Nil(t, p.Netlog)
	assert.Nil(t, p.Cfg)
	assert.Nil(t, p.Config)
	assert.Nil(t, p.BePath)
	assert.Nil(t, p.Ip)
	assert.Nil(t, p.Par)
	assert.Nil(t, p.LoadMissionToMemory)
	assert.Nil(t, p.AutoInit)
	assert.Nil(t, p.ServerMod)
	assert.Nil(t, p.DisableServerThread)
	assert.Nil(t, p.BandwidthAlg)
	assert.Nil(t, p.LimitFPS)
	assert.Nil(t, p.MaxMem)
	assert.Nil(t, p.MaxFileCacheSize)
	assert.Nil(t, p.CpuCount)
	assert.Nil(t, p.CpuAffinity)
	assert.Nil(t, p.CpuMainThreadAffinity)
	assert.Nil(t, p.EnableHT)
	assert.Nil(t, p.ExThreads)
	assert.Nil(t, p.Malloc)
	assert.Nil(t, p.HugePages)
	assert.Nil(t, p.Debug)
	assert.Nil(t, p.NoFreezeCheck)
	assert.Nil(t, p.NoLogs)
	assert.Nil(t, p.NoFilePatching)
	assert.Nil(t, p.FilePatching)
	assert.Nil(t, p.Init)
	assert.Nil(t, p.Autotest)
	assert.Nil(t, p.Beta)
	assert.Nil(t, p.CfgDependenciesDebugPrint)
	assert.Nil(t, p.CheckSignatures)
	assert.Nil(t, p.CheckSignaturesFull)
	assert.Nil(t, p.DebugCallExtension)
	assert.Nil(t, p.KeysFolder)
	assert.Empty(t, p.PreprocDefine)
	assert.Nil(t, p.DumpAddonDependencyGraph)
	assert.Nil(t, p.NetworkDiagInterval)
}

func TestStartupParamsUnmarshalYAML(t *testing.T) {
	t.Run("bool_flags", func(t *testing.T) {
		content := `
server: true
netlog: true
no_logs: false
enable_ht: true
huge_pages: true
file_patching: true
`
		var p StartupParams
		require.NoError(t, yaml.Load([]byte(content), &p))

		require.NotNil(t, p.Server)
		assert.True(t, *p.Server)
		require.NotNil(t, p.Netlog)
		assert.True(t, *p.Netlog)
		require.NotNil(t, p.NoLogs)
		assert.False(t, *p.NoLogs)
		require.NotNil(t, p.EnableHT)
		assert.True(t, *p.EnableHT)
		require.NotNil(t, p.HugePages)
		assert.True(t, *p.HugePages)
		require.NotNil(t, p.FilePatching)
		assert.True(t, *p.FilePatching)
		// unset fields remain nil
		assert.Nil(t, p.Debug)
		assert.Nil(t, p.AutoInit)
	})

	t.Run("string_params", func(t *testing.T) {
		content := `
world: "Altis"
profiles: "C:\\arma3\\profiles"
mod:
  - "mod1"
  - "mod2"
  - "mod3"
server_mod:
  - "servermod1"
config: "C:\\arma3\\server.cfg"
cfg: "C:\\arma3\\basic.cfg"
be_path: "C:\\BattlEye"
ip: "192.168.1.100"
malloc: "tbbmalloc"
keys_folder:
  - "@mod1/keys"
  - "@mod2/keys"
`
		var p StartupParams
		require.NoError(t, yaml.Load([]byte(content), &p))

		require.NotNil(t, p.World)
		assert.Equal(t, "Altis", *p.World)
		require.NotNil(t, p.Profiles)
		assert.Equal(t, "C:\\arma3\\profiles", *p.Profiles)
		assert.Equal(t, []string{"mod1", "mod2", "mod3"}, p.Mod)
		assert.Equal(t, []string{"servermod1"}, p.ServerMod)
		require.NotNil(t, p.Config)
		assert.Equal(t, "C:\\arma3\\server.cfg", *p.Config)
		require.NotNil(t, p.Cfg)
		assert.Equal(t, "C:\\arma3\\basic.cfg", *p.Cfg)
		require.NotNil(t, p.BePath)
		assert.Equal(t, "C:\\BattlEye", *p.BePath)
		require.NotNil(t, p.Ip)
		assert.Equal(t, "192.168.1.100", *p.Ip)
		require.NotNil(t, p.Malloc)
		assert.Equal(t, "tbbmalloc", *p.Malloc)
		assert.Equal(t, []string{"@mod1/keys", "@mod2/keys"}, p.KeysFolder)
	})

	t.Run("numeric_params", func(t *testing.T) {
		content := `
port: 2302
limit_fps: 100
max_mem: 8192
max_file_cache_size: 2048
cpu_count: 8
ex_threads: 7
bandwidth_alg: 2
network_diag_interval: 5
`
		var p StartupParams
		require.NoError(t, yaml.Load([]byte(content), &p))

		require.NotNil(t, p.Port)
		assert.Equal(t, uint16(2302), *p.Port)
		require.NotNil(t, p.LimitFPS)
		assert.Equal(t, uint32(100), *p.LimitFPS)
		require.NotNil(t, p.MaxMem)
		assert.Equal(t, uint32(8192), *p.MaxMem)
		require.NotNil(t, p.MaxFileCacheSize)
		assert.Equal(t, uint32(2048), *p.MaxFileCacheSize)
		require.NotNil(t, p.CpuCount)
		assert.Equal(t, uint32(8), *p.CpuCount)
		require.NotNil(t, p.ExThreads)
		assert.Equal(t, uint8(7), *p.ExThreads)
		require.NotNil(t, p.BandwidthAlg)
		assert.Equal(t, uint8(2), *p.BandwidthAlg)
		require.NotNil(t, p.NetworkDiagInterval)
		assert.Equal(t, uint32(5), *p.NetworkDiagInterval)
	})

	t.Run("cpu_affinity_hex_string", func(t *testing.T) {
		content := `
cpu_affinity: "0xFF"
cpu_main_thread_affinity: "0x01"
`
		var p StartupParams
		require.NoError(t, yaml.Load([]byte(content), &p))

		require.NotNil(t, p.CpuAffinity)
		assert.Equal(t, "0xFF", *p.CpuAffinity)
		require.NotNil(t, p.CpuMainThreadAffinity)
		assert.Equal(t, "0x01", *p.CpuMainThreadAffinity)
	})

	t.Run("preproc_define_list", func(t *testing.T) {
		content := `
preproc_define:
  - "CMD__MACRO1=VALUE1"
  - "CMD__MACRO2"
`
		var p StartupParams
		require.NoError(t, yaml.Load([]byte(content), &p))

		require.Len(t, p.PreprocDefine, 2)
		assert.Equal(t, "CMD__MACRO1=VALUE1", p.PreprocDefine[0])
		assert.Equal(t, "CMD__MACRO2", p.PreprocDefine[1])
	})

	t.Run("empty_yaml_leaves_all_nil", func(t *testing.T) {
		var p StartupParams
		require.NoError(t, yaml.Load([]byte(`{}`), &p))
		assert.Nil(t, p.Server)
		assert.Nil(t, p.Port)
		assert.Nil(t, p.MaxMem)
		assert.Empty(t, p.PreprocDefine)
	})
}

func TestStartupParamsUnmarshalJSON(t *testing.T) {
	t.Run("bool_flags", func(t *testing.T) {
		data := `{"server":true,"netlog":true,"noLogs":false,"enableHT":true}`
		var p StartupParams
		require.NoError(t, json.Unmarshal([]byte(data), &p))

		require.NotNil(t, p.Server)
		assert.True(t, *p.Server)
		require.NotNil(t, p.Netlog)
		assert.True(t, *p.Netlog)
		require.NotNil(t, p.NoLogs)
		assert.False(t, *p.NoLogs)
		require.NotNil(t, p.EnableHT)
		assert.True(t, *p.EnableHT)
		assert.Nil(t, p.Debug)
	})

	t.Run("string_and_numeric", func(t *testing.T) {
		data := `{"world":"Stratis","port":2302,"maxMem":8192,"mod":["mod1","mod2"]}`
		var p StartupParams
		require.NoError(t, json.Unmarshal([]byte(data), &p))

		require.NotNil(t, p.World)
		assert.Equal(t, "Stratis", *p.World)
		require.NotNil(t, p.Port)
		assert.Equal(t, uint16(2302), *p.Port)
		require.NotNil(t, p.MaxMem)
		assert.Equal(t, uint32(8192), *p.MaxMem)
		assert.Equal(t, []string{"mod1", "mod2"}, p.Mod)
	})

	t.Run("preproc_define", func(t *testing.T) {
		data := `{"preprocDefine":["MACRO1=VAL","MACRO2"]}`
		var p StartupParams
		require.NoError(t, json.Unmarshal([]byte(data), &p))

		require.Len(t, p.PreprocDefine, 2)
		assert.Equal(t, "MACRO1=VAL", p.PreprocDefine[0])
		assert.Equal(t, "MACRO2", p.PreprocDefine[1])
	})

	t.Run("empty_object_leaves_all_nil", func(t *testing.T) {
		var p StartupParams
		require.NoError(t, json.Unmarshal([]byte(`{}`), &p))
		assert.Nil(t, p.Server)
		assert.Nil(t, p.Port)
		assert.Empty(t, p.PreprocDefine)
	})
}

func TestStartupParamsMarshalOmitsNil(t *testing.T) {
	t.Run("json_omits_nil", func(t *testing.T) {
		p := StartupParams{
			Server: boolPtr(true),
			Port:   uint16Ptr(2302),
		}
		data, err := json.Marshal(p)
		require.NoError(t, err)

		var m map[string]interface{}
		require.NoError(t, json.Unmarshal(data, &m))

		assert.Contains(t, m, "server")
		assert.Contains(t, m, "port")
		assert.NotContains(t, m, "maxMem")
		assert.NotContains(t, m, "mod")
		assert.NotContains(t, m, "command")
	})

	t.Run("yaml_omits_nil", func(t *testing.T) {
		p := StartupParams{
			Server: boolPtr(true),
			Port:   uint16Ptr(2302),
		}
		data, err := yaml.Dump(&p)
		require.NoError(t, err)

		var m map[string]interface{}
		require.NoError(t, yaml.Load(data, &m))

		assert.Contains(t, m, "server")
		assert.Contains(t, m, "port")
		assert.NotContains(t, m, "max_mem")
		assert.NotContains(t, m, "mod")
	})

	t.Run("false_bool_included_json", func(t *testing.T) {
		// false *bool must serialize (it is set), not be omitted
		p := StartupParams{NoLogs: boolPtr(false)}
		data, err := json.Marshal(p)
		require.NoError(t, err)

		var m map[string]interface{}
		require.NoError(t, json.Unmarshal(data, &m))
		assert.Contains(t, m, "noLogs")
		assert.Equal(t, false, m["noLogs"])
	})

	t.Run("false_bool_included_yaml", func(t *testing.T) {
		p := StartupParams{NoLogs: boolPtr(false)}
		data, err := yaml.Dump(&p)
		require.NoError(t, err)

		var m map[string]interface{}
		require.NoError(t, yaml.Load(data, &m))
		assert.Contains(t, m, "no_logs")
		assert.Equal(t, false, m["no_logs"])
	})
}
