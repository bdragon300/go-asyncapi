package main

type (
	toolConfigSelection struct {
		ObjectKindRe string `yaml:"objectKindRe"`
		ModuleURLRe  string `yaml:"moduleURLRe"`
		PathRe       string `yaml:"pathRe"`
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