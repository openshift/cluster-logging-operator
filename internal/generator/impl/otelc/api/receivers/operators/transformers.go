package operators

// MoveOperator moves a field from one location to another
type MoveOperator struct {
	Type   OperatorType `yaml:"type"`             // Must be OperatorTypeMove
	ID     string       `yaml:"id,omitempty"`     // Optional unique ID
	Output string       `yaml:"output,omitempty"` // Next operator ID
	From   string       `yaml:"from"`             // Source field path
	To     string       `yaml:"to"`               // Destination field path
}

// NewMoveOperator creates a move operator
func NewMoveOperator(id, from, to string) *MoveOperator {
	return &MoveOperator{
		Type: OperatorTypeMove,
		ID:   id,
		From: from,
		To:   to,
	}
}

// AddOperator adds a new field with a specified value
type AddOperator struct {
	Type      OperatorType `yaml:"type"`                 // Must be OperatorTypeAdd
	ID        string       `yaml:"id,omitempty"`         // Optional unique ID
	Output    string       `yaml:"output,omitempty"`     // Next operator ID
	Field     string       `yaml:"field"`                // Field path to add
	Value     interface{}  `yaml:"value"`                // Value to set
	ValueExpr string       `yaml:"value_expr,omitempty"` // Expression to evaluate for value
	IfExpr    string       `yaml:"if,omitempty"`         // Condition for adding field
}

// NewAddOperator creates an add operator
func NewAddOperator(id, field string, value interface{}) *AddOperator {
	return &AddOperator{
		Type:  OperatorTypeAdd,
		ID:    id,
		Field: field,
		Value: value,
	}
}

// RemoveOperator removes specified fields
type RemoveOperator struct {
	Type   OperatorType `yaml:"type"`             // Must be OperatorTypeRemove
	ID     string       `yaml:"id,omitempty"`     // Optional unique ID
	Output string       `yaml:"output,omitempty"` // Next operator ID
	Field  string       `yaml:"field"`            // Field path to remove
}

// NewRemoveOperator creates a remove operator
func NewRemoveOperator(id, field string) *RemoveOperator {
	return &RemoveOperator{
		Type:  OperatorTypeRemove,
		ID:    id,
		Field: field,
	}
}

// RecombineOperator combines multiple log lines into a single entry
type RecombineOperator struct {
	Type             OperatorType `yaml:"type"`                        // Must be OperatorTypeRecombine
	ID               string       `yaml:"id,omitempty"`                // Optional unique ID
	Output           string       `yaml:"output,omitempty"`            // Next operator ID
	CombineField     string       `yaml:"combine_field,omitempty"`     // Field to combine (default: "body")
	IsFirstEntry     string       `yaml:"is_first_entry,omitempty"`    // Expression to detect first entry
	IsLastEntry      string       `yaml:"is_last_entry,omitempty"`     // Expression to detect last entry
	MaxLogSize       int          `yaml:"max_log_size,omitempty"`      // Maximum combined size
	OverwriteWith    string       `yaml:"overwrite_with,omitempty"`    // "oldest" or "newest"
	SourceIdentifier string       `yaml:"source_identifier,omitempty"` // Expression for grouping entries
}
