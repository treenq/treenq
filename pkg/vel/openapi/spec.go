package openapi

type Spec struct {
	Description string
	Errors      []ErrorSpec
}

type ErrorSpec struct {
	Code        string
	Description string
	Meta        []ErrorMetaSpec
}

type ErrorMetaSpec struct {
	Key          string
	ValueType    string
	ValueExample string
	Description  string
}
