package operators

// RouterOperator routes logs to different operators based on expressions
type RouterOperator struct {
	Type    OperatorType `yaml:"type"`              // Must be OperatorTypeRouter
	ID      string       `yaml:"id,omitempty"`      // Optional unique ID
	Routes  []RouteEntry `yaml:"routes,omitempty"`  // Route definitions
	Default string       `yaml:"default,omitempty"` // Default output if no route matches
}

// NewRouterOperator creates a router operator
func NewRouterOperator(id string, routes []RouteEntry) *RouterOperator {
	return &RouterOperator{
		Type:   OperatorTypeRouter,
		ID:     id,
		Routes: routes,
	}
}

// FilterOperator includes or excludes log entries based on expressions
type FilterOperator struct {
	Type   OperatorType `yaml:"type"`             // Must be OperatorTypeFilter
	ID     string       `yaml:"id,omitempty"`     // Optional unique ID
	Output string       `yaml:"output,omitempty"` // Next operator ID
	Expr   string       `yaml:"expr"`             // Boolean expression for filtering
}
