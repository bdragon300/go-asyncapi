package schema

type ComponentsItem struct {
	Schemas map[string]Object `json:"schemas" yaml:"schemas" cgen:"noinline"`
}
