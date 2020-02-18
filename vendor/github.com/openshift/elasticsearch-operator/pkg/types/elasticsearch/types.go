package elasticsearch

func NewIndexTemplate(pattern string, aliases []string, shards, replicas int32) *IndexTemplate {
	template := IndexTemplate{
		Template: pattern,
		Settings: IndexSettings{
			NumberOfShards:   shards,
			NumberOfReplicas: replicas,
		},
		Aliases: map[string]IndexAlias{},
	}
	for _, alias := range aliases {
		template.Aliases[alias] = IndexAlias{}
	}
	return &template
}

func NewIndex(name string, shards, replicas int32) *Index {
	index := Index{
		Name: name,
		Settings: IndexSettings{
			NumberOfShards:   shards,
			NumberOfReplicas: replicas,
		},
		Aliases: map[string]IndexAlias{},
	}
	return &index
}

func (index *Index) AddAlias(name string, isWriteIndex bool) *Index {
	alias := IndexAlias{}
	if isWriteIndex {
		alias.IsWriteIndex = true
	}
	index.Aliases[name] = alias
	return index
}

type Index struct {
	//Name  intentionally not serialized
	Name     string                `json:"-"`
	Settings IndexSettings         `json:"settings,omitempty"`
	Aliases  map[string]IndexAlias `json:"aliases,omitempty"`
}

type IndexTemplate struct {
	Template string                `json:"template,omitempty"`
	Settings IndexSettings         `json:"settings,omitempty"`
	Aliases  map[string]IndexAlias `json:"aliases,omitempty"`
}

type Aliases struct {
}

type IndexAlias struct {
	IsWriteIndex bool `json:"is_write_index,omitempty"`
}

type IndexSettings struct {
	NumberOfShards   int32 `json:"number_of_shards"`
	NumberOfReplicas int32 `json:"number_of_replicas"`
}
