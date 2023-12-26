package asyncapi

type xGoTypeHint struct {
	Kind string `json:"kind" yaml:"kind"`
}

type xGoType struct {
	Type     string      `json:"type" yaml:"type"`
	Embedded *bool       `json:"embedded" yaml:"embedded"`
	Hint     xGoTypeHint `json:"hint" yaml:"hint"`
}

