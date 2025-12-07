package asyncapi

type xGoTypeHint struct {
	Kind    string `json:"kind,omitzero" yaml:"kind"`
	Pointer bool   `json:"pointer,omitzero" yaml:"pointer"`
}

type xGoTypeImportPackage struct {
	Package string `json:"package,omitzero" yaml:"package"`
}

type xGoType struct {
	Type     string               `json:"type,omitzero" yaml:"type"`
	Import   xGoTypeImportPackage `json:"import,omitzero" yaml:"import"`
	Embedded bool                 `json:"embedded,omitzero" yaml:"embedded"`
	Hint     xGoTypeHint          `json:"hint,omitzero" yaml:"hint"`
}
