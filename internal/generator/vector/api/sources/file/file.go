package file


type sourceType string

const SourceTypeFile sourceType = "file"

type File struct {
	// Type is required to be 'file'
	Type sourceType `json:"type" yaml:"type" toml:"type"`

	// Include is file paths to include for this source
	Include []string `json:"include" yaml:"include" toml:"include"`

	HostKey                           string `json:"host_key,omitempty" yaml:"host_key,omitempty" toml:"host_key,omitempty"`
	GlobalMinimumCooldownMilliSeconds int64  `json:"glob_minimum_cooldown_ms,omitempty" yaml:"glob_minimum_cooldown_ms" toml:"glob_minimum_cooldown_ms,omitempty"`
	IgnoreOlderSecs                   int64  `json:"ignore_older_secs,omitempty" yaml:"ignore_older_secs,omitempty" toml:"ignore_older_secs,omitempty"`
	MaxLineBytes                      int64  `json:"max_line_bytes,omitempty" yaml:"max_line_bytes,omitempty" toml:"max_line_bytes,omitempty"`
	MaxReadBytes                      int64  `json:"max_read_bytes,omitempty" yaml:"max_read_bytes,omitempty" toml:"max_read_bytes,omitempty"`
	RotateWaitSecs                    int64  `json:"rotate_wait_secs,omitempty" yaml:"rotate_wait_secs,omitempty" toml:"rotate_wait_secs,omitempty"`
}

func New(include ...string) *File {
	return &File{
		Type:    SourceTypeFile,
		Include: include,
	}
}
