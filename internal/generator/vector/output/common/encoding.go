package common

const (
	CodecJSON              = "json"
	TimeStampFormatRFC3339 = "rfc3339"
)

type Encoding struct {
	ID             string   `toml:"-"`
	Codec          string  `toml:"codec,omitempty"`
	//ExceptFields is a VRL acceptable List
	ExceptFields   []string `toml:"except_fields,omitempty"`
	TimeStampFormat string `toml:"timestamp_format,omitempty"`
}

func NewEncoding(id string, codec string, inits ...func(*Encoding)) Encoding {
	e := &Encoding{
		ID:           id,
		ExceptFields: []string{"_internal"},
	}

	if codec != "" {
		e.Codec = codec
	}

	for _, init := range inits {
		init(e)
	}

	return *e
}

func (e Encoding) Name() string {
	return "encoding"
}

func (e Encoding) Template() string {
	return ""
}

func (e Encoding) Config() any {
	return map[string]interface{}{
		"sinks." + e.ID + ".encoding": e,
	}
}
