package model

type Tool struct {
	Id       string   `json:"id,omitempty"`
	Type     string   `json:"type,omitempty"` // when splicing claude tools stream messages, it is empty
	Function Function `json:"function"`
	Index    *int     `json:"index,omitempty"` // Index identifies which function call the delta is for in streaming responses
}

type Function struct {
	Description string         `json:"description,omitempty"`
	Name        string         `json:"name,omitempty"`       // when splicing claude tools stream messages, it is empty
	Parameters  map[string]any `json:"parameters,omitempty"` // request
	Arguments   any            `json:"arguments,omitempty"`  // response
	Required    []string       `json:"required,omitempty"`   // request
	Strict      *bool          `json:"strict,omitempty"`     // request
}
