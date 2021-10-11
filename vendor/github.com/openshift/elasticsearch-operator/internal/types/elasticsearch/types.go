package elasticsearch

func NewIndexTemplate(pattern string, aliases []string, shards, replicas int32) *IndexTemplate {
	template := IndexTemplate{
		Template: pattern,
		Settings: IndexSettings{
			Index: &IndexingSettings{
				NumberOfShards:   shards,
				NumberOfReplicas: replicas,
			},
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
		Settings: &IndexSettings{
			Index: &IndexingSettings{
				NumberOfShards:   shards,
				NumberOfReplicas: replicas,
			},
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
	// Name  intentionally not serialized
	Name     string                 `json:"-"`
	Settings *IndexSettings         `json:"settings,omitempty"`
	Aliases  map[string]IndexAlias  `json:"aliases,omitempty"`
	Mappings map[string]interface{} `json:"mappings,omitempty"`
}

type IndexTemplate struct {
	Template string                `json:"template,omitempty"`
	Settings IndexSettings         `json:"settings,omitempty"`
	Aliases  map[string]IndexAlias `json:"aliases,omitempty"`
}

type GetIndexTemplate struct {
	Order         int32                           `json:"order,omitempty"`
	IndexPatterns []string                        `json:"index_patterns,omitempty"`
	Settings      GetIndexTemplateSettings        `json:"settings,omitempty"`
	Aliases       map[string]IndexAlias           `json:"aliases,omitempty"`
	Mappings      map[string]IndexMappingSettings `json:"mappings,omitempty"`
}

type GetIndexTemplateSettings struct {
	Index IndexTemplateSettings `json:"index,omitempty"`
}

type IndexTemplateSettings struct {
	Unassigned       UnassignedIndexSetting `json:"unassigned,omitempty"`
	Translog         TranslogIndexSetting   `json:"translog,omitempty"`
	RefreshInterval  string                 `json:"refresh_interval,omitempty"`
	NumberOfShards   string                 `json:"number_of_shards,omitempty"`
	NumberOfReplicas string                 `json:"number_of_replicas,omitempty"`
}

type UnassignedIndexSetting struct {
	NodeLeft NodeLeftSetting `json:"node_left,omitempty"`
}

type NodeLeftSetting struct {
	DelayedTimeout string `json:"delayed_timeout,omitempty"`
}

type TranslogIndexSetting struct {
	FlushThresholdSize string `json:"flush_threshold_size,omitempty"`
}

type Aliases struct {
}

type IndexAlias struct {
	IsWriteIndex bool `json:"is_write_index,omitempty"`
}

type IndexSettings struct {
	Index *IndexingSettings `json:"index,omitempty"`
}

type IndexingSettings struct {
	NumberOfShards   int32                 `json:"number_of_shards,string,omitempty"`
	NumberOfReplicas int32                 `json:"number_of_replicas,string,omitempty"`
	Format           int32                 `json:"format,omitempty"`
	Blocks           *IndexBlocksSettings  `json:"blocks,omitempty"`
	Mapper           *IndexMapperSettings  `json:"mapper,omitempty"`
	Mapping          *IndexMappingSettings `json:"mapping,omitempty"`
}

type IndexBlocksSettings struct {
	Write               bool    `json:"write,omitempty"`
	ReadOnlyAllowDelete *string `json:"read_only_allow_delete"`
}

type IndexMapperSettings struct {
	Dynamic bool `json:"dynamic"`
}

type IndexMappingSettings struct {
	SingleType bool `json:"single_type"`
}

type ReIndex struct {
	Source IndexRef      `json:"source"`
	Dest   IndexRef      `json:"dest"`
	Script ReIndexScript `json:"script"`
}

type ReIndexScript struct {
	Inline string `json:"inline"`
	Lang   string `json:"lang"`
}

type IndexRef struct {
	Index string `json:"index"`
}

type AliasActions struct {
	Actions []AliasAction `json:"actions"`
}

type AliasAction struct {
	Add         *AddAliasAction    `json:"add,omitempty"`
	RemoveIndex *RemoveAliasAction `json:"remove_index,omitempty"`
}

type AddAliasAction struct {
	Index string `json:"index"`
	Alias string `json:"alias"`
}

type RemoveAliasAction struct {
	Index string `json:"index"`
}

type CatIndicesResponses []CatIndicesResponse

type CatIndicesResponse struct {
	Health           string `json:"health,omitempty"`
	Status           string `json:"status,omitempty"`
	Index            string `json:"index,omitempty"`
	UUID             string `json:"uuis,omitempty"`
	Primaries        string `json:"pri,omitempty"`
	Replicas         string `json:"rep,omitempty"`
	DocsCount        string `json:"docs.count,omitempty"`
	DocsDeleted      string `json:"docs.deleted,omitempty"`
	StoreSize        string `json:"store.size,omitempty"`
	PrimaryStoreSize string `json:"pri.store.size,omitempty"`
}

type MasterNodeAndNodeStateResponse struct {
	ClusterName string                       `json:"cluster_name,omitempty"`
	MasterNode  string                       `json:"master_node,omitempty"`
	Nodes       map[string]NodeStateResponse `json:"nodes,omitempty"`
}

type NodesStateResponse struct {
	Nodes map[string]NodeStateResponse `json:"nodes,omitempty"`
}

type NodeStateResponse struct {
	Name             string            `json:"name,omitempty"`
	EphemeralID      string            `json:"ephemeral_id,omitempty"`
	TransportAddress string            `json:"transport_address,omitempty"`
	Attributes       map[string]string `json:"attributes,omitempty"`
}

type StatsNodesResponse struct {
	Nodes StatsNode `json:"nodes,omitempty"`
}

type StatsNode struct {
	Versions []string       `json:"versions,omitempty"`
	Count    map[string]int `json:"count,omitempty"`
}
