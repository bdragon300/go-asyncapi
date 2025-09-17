package asyncapi

type xGoTypeHint struct {
	Kind    string `json:"kind" yaml:"kind"`
	Pointer bool   `json:"pointer" yaml:"pointer"`
}

type xGoTypeImportPackage struct {
	Package string `json:"package" yaml:"package"`
}

type xGoType struct {
	Type     string               `json:"type" yaml:"type"`
	Import   xGoTypeImportPackage `json:"import" yaml:"import"`
	Embedded bool                 `json:"embedded" yaml:"embedded"`
	Hint     xGoTypeHint          `json:"hint" yaml:"hint"`
}
