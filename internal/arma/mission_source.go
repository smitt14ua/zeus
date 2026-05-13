package arma

// MissionSource describes where mission .pbo files are fetched from.
// Driver "path" is the only supported driver in v1.
type MissionSource struct {
	Driver string `json:"driver"         yaml:"driver"         toml:"driver"`
	Path   string `json:"path,omitempty" yaml:"path,omitempty" toml:"path,omitempty"`
	// Mode: "copy" (default) or "symlink". symlink replaces the mpmissions dir with a link to Path.
	Mode string `json:"mode,omitempty" yaml:"mode,omitempty" toml:"mode,omitempty"`
	// S3 driver fields
	Bucket          string `json:"bucket,omitempty"            yaml:"bucket,omitempty"            toml:"bucket,omitempty"`
	Prefix          string `json:"prefix,omitempty"            yaml:"prefix,omitempty"            toml:"prefix,omitempty"`
	Endpoint        string `json:"endpoint,omitempty"          yaml:"endpoint,omitempty"          toml:"endpoint,omitempty"`
	Region          string `json:"region,omitempty"            yaml:"region,omitempty"            toml:"region,omitempty"`
	AccessKeyID     string `json:"accessKeyId,omitempty"       yaml:"access_key_id,omitempty"     toml:"access_key_id,omitempty"`
	SecretAccessKey string `json:"secretAccessKey,omitempty"   yaml:"secret_access_key,omitempty" toml:"secret_access_key,omitempty"`
}
