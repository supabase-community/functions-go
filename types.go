package functions_go

import "io"

const (
	json        = "json"
	text        = "text"
	arrayBuffer = "arrayBuffer"
	blob        = "blob"
)

type FunctionInvokeOptions struct {
	Body         io.Reader `json:"body,omitempty"`
	ResponseType string    `json:"responseType,omitempty"`
}

type FunctionResponse struct {
	Data  interface{} `json:"data,omitempty"`
	Error interface{} `json:"error,omitempty"`
	//Not specifying int as Status type because it can be null
	Status interface{} `json:"responseType,omitempty"`
}
