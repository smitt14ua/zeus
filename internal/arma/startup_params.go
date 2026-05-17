package arma

// DefaultPort is the Arma 3 dedicated server game port used when no explicit
// port is configured in a profile.
const DefaultPort uint16 = 2302

// StartupParams holds Arma 3 dedicated server startup parameters.
// All fields are optional (pointer/slice); nil means the parameter is omitted.
// Boolean flags (no value) use *bool — nil = omit, true = include, false = omit.
// The `arg` tag holds the exact CLI parameter name as used by the engine.
type StartupParams struct {
	// --- Game loading ---
	// Empty string "empty" disables default world load.
	World    *string `json:"world,omitempty"    yaml:"world,omitempty"    toml:"world,omitempty"    arg:"world"`
	WorldCfg *string `json:"worldCfg,omitempty" yaml:"world_cfg,omitempty" toml:"world_cfg,omitempty" arg:"worldCfg"`

	// --- Profile ---
	Name     *string `json:"name,omitempty"     yaml:"name,omitempty"     toml:"name,omitempty"     arg:"name"`
	Profiles *string `json:"profiles,omitempty" yaml:"profiles,omitempty" toml:"profiles,omitempty" arg:"profiles"`

	// --- Misc ---
	DoNothing *bool    `json:"doNothing,omitempty"  yaml:"do_nothing,omitempty"  toml:"do_nothing,omitempty"  arg:"doNothing"`
	Mod       []string `json:"mod,omitempty"        yaml:"mod,omitempty"         toml:"mod,omitempty"         arg:"mod"`
	// Relative to Arma 3 directory.
	MpMissions *string `json:"mpMissions,omitempty" yaml:"mp_missions,omitempty" toml:"mp_missions,omitempty" arg:"mpmissions"`

	// --- Server ---
	Server *bool   `json:"server,omitempty" yaml:"server,omitempty" toml:"server,omitempty" arg:"server"`
	Port   *uint16 `json:"port,omitempty"   yaml:"port,omitempty"   toml:"port,omitempty"   arg:"port"`
	// Path for the server PID file; removed automatically on exit.
	Pid     *string `json:"pid,omitempty"     yaml:"pid,omitempty"     toml:"pid,omitempty"     arg:"pid"`
	Ranking *string `json:"ranking,omitempty" yaml:"ranking,omitempty" toml:"ranking,omitempty" arg:"ranking"`
	Netlog  *bool   `json:"netlog,omitempty"  yaml:"netlog,omitempty"  toml:"netlog,omitempty"  arg:"netlog"`
	// Path to basic.cfg (network performance tuning).
	Cfg *string `json:"cfg,omitempty" yaml:"cfg,omitempty" toml:"cfg,omitempty" arg:"cfg"`
	// Path to server.cfg (admin password, mission selection, etc.).
	Config *string `json:"config,omitempty" yaml:"config,omitempty" toml:"config,omitempty" arg:"config"`
	// Custom BattlEye folder path; default is BattlEye inside profiles folder.
	BePath *string `json:"bePath,omitempty" yaml:"be_path,omitempty" toml:"be_path,omitempty" arg:"bePath"`
	// Bind IP for multihome servers.
	Ip *string `json:"ip,omitempty" yaml:"ip,omitempty" toml:"ip,omitempty" arg:"ip"`
	// Path to a startup parameters config file.
	Par                 *string `json:"par,omitempty"                yaml:"par,omitempty"                 toml:"par,omitempty"                 arg:"par"`
	LoadMissionToMemory *bool   `json:"loadMissionToMemory,omitempty" yaml:"load_mission_to_memory,omitempty" toml:"load_mission_to_memory,omitempty" arg:"loadMissionToMemory"`
	// Requires Persistent=1 in server.cfg; breaks mission parameters.
	AutoInit            *bool    `json:"autoInit,omitempty"            yaml:"auto_init,omitempty"            toml:"auto_init,omitempty"            arg:"autoInit"`
	ServerMod           []string `json:"serverMod,omitempty"           yaml:"server_mod,omitempty"           toml:"server_mod,omitempty"           arg:"serverMod"`
	DisableServerThread *bool    `json:"disableServerThread,omitempty" yaml:"disable_server_thread,omitempty" toml:"disable_server_thread,omitempty" arg:"disableServerThread"`
	// Set to 2 for experimental networking algorithm.
	BandwidthAlg *uint8 `json:"bandwidthAlg,omitempty" yaml:"bandwidth_alg,omitempty" toml:"bandwidth_alg,omitempty" arg:"bandwidthAlg"`
	// Use !keys to exclude the default base game "keys" folder.
	KeysFolder []string `json:"keysFolder,omitempty" yaml:"keys_folder,omitempty" toml:"keys_folder,omitempty" arg:"keysFolder"`

	// --- Performance ---
	// FPS limit for dedicated server / headless client (range 5–1000; default 50).
	LimitFPS *uint32 `json:"limitFPS,omitempty" yaml:"limit_fps,omitempty" toml:"limit_fps,omitempty" arg:"limitFPS"`
	// Memory allocation limit in MiB; minimum 1024.
	MaxMem *uint32 `json:"maxMem,omitempty" yaml:"max_mem,omitempty" toml:"max_mem,omitempty" arg:"maxMem"`
	// File cache size in MiB; minimum 512, recommended ≥1024.
	MaxFileCacheSize *uint32 `json:"maxFileCacheSize,omitempty" yaml:"max_file_cache_size,omitempty" toml:"max_file_cache_size,omitempty" arg:"maxFileCacheSize"`
	// Override CPU core count (≥2 since 2.20).
	CpuCount *uint32 `json:"cpuCount,omitempty" yaml:"cpu_count,omitempty" toml:"cpu_count,omitempty" arg:"cpuCount"`
	// Process CPU affinity mask (hex string, e.g. "0xFF").
	CpuAffinity *string `json:"cpuAffinity,omitempty" yaml:"cpu_affinity,omitempty" toml:"cpu_affinity,omitempty" arg:"cpuAffinity"`
	// Main thread CPU affinity mask (hex string, e.g. "0x01").
	CpuMainThreadAffinity *string `json:"cpuMainThreadAffinity,omitempty" yaml:"cpu_main_thread_affinity,omitempty" toml:"cpu_main_thread_affinity,omitempty" arg:"cpuMainThreadAffinity"`
	// Enable logical CPU cores (HyperThreading); overridden by cpuCount/cpuAffinity.
	EnableHT *bool `json:"enableHT,omitempty" yaml:"enable_ht,omitempty" toml:"enable_ht,omitempty" arg:"enableHT"`
	// Extra threads: 0=none, 1=fileOps, 3=tex+file, 5=geo+file, 7=all.
	ExThreads *uint8  `json:"exThreads,omitempty" yaml:"ex_threads,omitempty" toml:"ex_threads,omitempty" arg:"exThreads"`
	Malloc    *string `json:"malloc,omitempty"    yaml:"malloc,omitempty"    toml:"malloc,omitempty"    arg:"malloc"`
	HugePages *bool   `json:"hugePages,omitempty" yaml:"huge_pages,omitempty" toml:"huge_pages,omitempty" arg:"hugePages"`

	// --- Developer ---
	Debug         *bool `json:"debug,omitempty"         yaml:"debug,omitempty"          toml:"debug,omitempty"          arg:"debug"`
	NoFreezeCheck *bool `json:"noFreezeCheck,omitempty" yaml:"no_freeze_check,omitempty" toml:"no_freeze_check,omitempty" arg:"noFreezeCheck"`
	NoLogs        *bool `json:"noLogs,omitempty"        yaml:"no_logs,omitempty"         toml:"no_logs,omitempty"         arg:"noLogs"`
	// Deprecated since 1.50; prefer FilePatching.
	NoFilePatching *bool `json:"noFilePatching,omitempty" yaml:"no_file_patching,omitempty" toml:"no_file_patching,omitempty" arg:"noFilePatching"`
	FilePatching   *bool `json:"filePatching,omitempty"   yaml:"file_patching,omitempty"   toml:"file_patching,omitempty"   arg:"filePatching"`
	// SQF expression to run on main menu startup.
	Init     *string `json:"init,omitempty"     yaml:"init,omitempty"     toml:"init,omitempty"     arg:"init"`
	Autotest *string `json:"autotest,omitempty" yaml:"autotest,omitempty" toml:"autotest,omitempty" arg:"autotest"`
	// Separated by semicolons; absolute paths and multiple stacked folders are possible.
	Beta                      []string `json:"beta,omitempty"                      yaml:"beta,omitempty"                      toml:"beta,omitempty"                      arg:"beta"`
	CfgDependenciesDebugPrint *bool    `json:"cfgDependenciesDebugPrint,omitempty" yaml:"cfg_dependencies_debug_print,omitempty" toml:"cfg_dependencies_debug_print,omitempty" arg:"cfgDependenciesDebugPrint"`
	CheckSignatures           *bool    `json:"checkSignatures,omitempty"           yaml:"check_signatures,omitempty"           toml:"check_signatures,omitempty"           arg:"checkSignatures"`
	CheckSignaturesFull       *bool    `json:"checkSignaturesFull,omitempty"       yaml:"check_signatures_full,omitempty"      toml:"check_signatures_full,omitempty"      arg:"checkSignaturesFull"`
	DebugCallExtension        *bool    `json:"debugCallExtension,omitempty"        yaml:"debug_call_extension,omitempty"       toml:"debug_call_extension,omitempty"       arg:"debugCallExtension"`
	// Each entry defines a preprocessor macro, optionally with value ("MACRO=VALUE" or "MACRO").
	PreprocDefine            []string `json:"preprocDefine,omitempty"            yaml:"preproc_define,omitempty"             toml:"preproc_define,omitempty"             arg:"preprocDefine"`
	DumpAddonDependencyGraph *bool    `json:"dumpAddonDependencyGraph,omitempty" yaml:"dump_addon_dependency_graph,omitempty" toml:"dump_addon_dependency_graph,omitempty" arg:"dumpAddonDependencyGraph"`
	// Seconds between network diagnostics polling.
	NetworkDiagInterval *uint32 `json:"networkDiagInterval,omitempty" yaml:"network_diag_interval,omitempty" toml:"network_diag_interval,omitempty" arg:"networkDiagInterval"`
}
