package v1alpha1

type LogSourceTypeSet map[LogSourceType]string

func NewLogSourceTypeSet() LogSourceTypeSet {
	return LogSourceTypeSet(map[LogSourceType]string{})
}

func (s *LogSourceTypeSet) Insert(sourceType LogSourceType) {
	(*s)[sourceType] = ""
}

func (s *LogSourceTypeSet) List() []LogSourceType {
	list := []LogSourceType{}
	for k := range *s {
		list = append(list, k)
	}
	return list
}
