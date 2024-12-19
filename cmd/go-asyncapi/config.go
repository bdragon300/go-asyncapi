package main

type (
	toolConfigSelection struct {
		// TODO: add filtering by name
		ObjectKindRe string `yaml:"objectKindRe"`
		ModuleURLRe  string `yaml:"moduleURLRe"`
		PathRe       string `yaml:"pathRe"`
		Protocols   []string `yaml:"protocols"`
		IgnoreCommon bool `yaml:"ignoreCommon"` // TODO: rename
		Template     string                       `yaml:"template"`
		File         string                       `yaml:"file"`
		Package      string                       `yaml:"package"`
		TemplateArgs map[string]string            `yaml:"templateArgs"`
	}

	toolConfigRender struct {
		Selections []toolConfigSelection `yaml:"selections"`
	}

	toolConfig struct {
		Render toolConfigRender `yaml:"render"`
	}
)