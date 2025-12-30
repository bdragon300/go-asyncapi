package main

import (
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"time"

	"github.com/bdragon300/go-asyncapi/internal/common"
	"github.com/bdragon300/go-asyncapi/internal/types"
	"gopkg.in/yaml.v3"
)

type d2DiagramEngine string

const (
	D2DiagramEngineELK   d2DiagramEngine = "elk"
	D2DiagramEngineDagre d2DiagramEngine = "dagre"
)

// Structures, that represent the tool's configuration file
type (
	toolConfig struct {
		ConfigVersion int    `yaml:"configVersion"`
		ProjectModule string `yaml:"projectModule"`
		RuntimeModule string `yaml:"runtimeModule"`
		TemplatesDir  string `yaml:"templatesDir"`

		Locator toolConfigLocator `yaml:"locator"`

		Code    toolConfigCode    `yaml:"code"`
		Client  toolConfigClient  `yaml:"client"`
		Infra   toolConfigInfra   `yaml:"infra"`
		Diagram toolConfigDiagram `yaml:"diagram"`
		UI      toolConfigUI      `yaml:"ui"`
	}

	toolConfigLocator struct {
		AllowRemoteReferences bool          `yaml:"allowRemoteReferences"`
		RootDirectory         string        `yaml:"rootDirectory"`
		Timeout               time.Duration `yaml:"timeout"`
		Command               string        `yaml:"command"`
	}

	toolConfigCode struct {
		OnlyPublish       bool   `yaml:"onlyPublish"`
		OnlySubscribe     bool   `yaml:"onlySubscribe"`
		DisableFormatting bool   `yaml:"disableFormatting"`
		TargetDir         string `yaml:"targetDir"`

		Layout []toolConfigCodeLayout `yaml:"layout"`

		PreambleTemplate string `yaml:"preambleTemplate"`

		Util           toolConfigCodeUtil           `yaml:"util"`
		Implementation toolConfigCodeImplementation `yaml:"implementation"`
	}

	toolConfigCodeLayout struct {
		NameRe        string                 `yaml:"nameRe"`
		ArtifactKinds []string               `yaml:"artifactKinds"`
		ModuleURLRe   string                 `yaml:"moduleURLRe"` // TODO: rename to locationRe or smth like that
		PathRe        string                 `yaml:"pathRe"`      // TODO: remove? almost duplicate of moduleURLRe
		Protocols     []string               `yaml:"protocols"`
		Not           bool                   `yaml:"not"` // Inverts the match, i.e. NOT operation
		Render        toolConfigLayoutRender `yaml:"render"`
	}

	toolConfigLayoutRender struct {
		Protocols []string `yaml:"protocols"`
		Template  string   `yaml:"template"`
		File      string   `yaml:"file"`
		Package   string   `yaml:"package"` // TODO: make it inline template
	}

	toolConfigCodeUtil struct {
		Directory string                       `yaml:"directory"` // Template expression, relative to the target directory
		Custom    []toolConfigCodeUtilProtocol `yaml:"custom"`
	}

	toolConfigCodeUtilProtocol struct {
		Protocol          string `yaml:"protocol"`
		TemplateDirectory string `yaml:"templateDirectory"`
	}

	toolConfigCodeImplementation struct {
		Directory string                             `yaml:"directory"` // Template expression, relative to the target directory
		Disable   bool                               `yaml:"disable"`
		Custom    []toolConfigImplementationProtocol `yaml:"custom"`
	}

	toolConfigImplementationProtocol struct {
		Protocol          string `yaml:"protocol"`
		Name              string `yaml:"name"`
		Disable           bool   `yaml:"disable"`
		TemplateDirectory string `yaml:"templateDirectory"`
		Package           string `yaml:"package"`
	}

	toolConfigClient struct {
		OutputFile       string `yaml:"outputFile"`
		OutputSourceFile string `yaml:"outputSourceFile"`
		KeepSource       bool   `yaml:"keepSource"`
		GoModTemplate    string `yaml:"goModTemplate"`
		TempDir          string `yaml:"tempDir"`
	}

	toolConfigInfra struct {
		ServerOpts []toolConfigInfraServerOpt `yaml:"serverOpts"`
		Engine     string                     `yaml:"engine"`
		OutputFile string                     `yaml:"outputFile"`
	}

	toolConfigInfraServerOpt struct {
		ServerName string                                                                             `yaml:"serverName"` // TODO: make required
		Variables  types.Union2[types.OrderedMap[string, string], []types.OrderedMap[string, string]] `yaml:"variables"`
	}

	toolConfigDiagram struct {
		Format common.DiagramOutputFormat `yaml:"format"`

		OutputFile        string `yaml:"outputFile"`
		TargetDir         string `yaml:"targetDir"`
		MultipleFiles     bool   `yaml:"multipleFiles"`
		DisableFormatting bool   `yaml:"disableFormatting"`

		ChannelsCentric bool `yaml:"channelsCentric"`
		ServersCentric  bool `yaml:"serversCentric"`
		DocumentBorders bool `yaml:"documentBorders"`

		D2 toolConfigDiagramD2Opts `yaml:"d2"`
	}

	toolConfigDiagramD2Opts struct {
		Engine      d2DiagramEngine              `yaml:"engine"`
		Direction   common.D2DiagramDirection    `yaml:"direction"`
		ThemeID     *int64                       `yaml:"themeId"`
		DarkThemeID *int64                       `yaml:"darkThemeId"`
		Pad         *int64                       `yaml:"pad"`
		Sketch      *bool                        `yaml:"sketch"`
		Center      *bool                        `yaml:"center"`
		Scale       *float64                     `yaml:"scale"`
		ELK         toolConfigDiagramD2ELKOpts   `yaml:"elk"`
		Dagre       toolConfigDiagramD2DagreOpts `yaml:"dagre"`
	}

	toolConfigDiagramD2ELKOpts struct {
		Algorithm       string `yaml:"algorithm"`
		NodeSpacing     int64  `yaml:"nodeSpacing"`
		Padding         string `yaml:"padding"`
		EdgeSpacing     int64  `yaml:"edgeSpacing"`
		SelfLoopSpacing int64  `yaml:"selfLoopSpacing"`
	}

	toolConfigDiagramD2DagreOpts struct {
		NodeSep int64 `yaml:"nodeSep"`
		EdgeSep int64 `yaml:"edgeSep"`
	}

	toolConfigUI struct {
		OutputFile string `yaml:"outputFile"`

		Listen        *bool  `yaml:"listen"`
		ListenAddress string `yaml:"listenAddress"`
		ListenPath    string `yaml:"listenPath"`
		ReferenceOnly *bool  `yaml:"referenceOnly"`
		Bundle        *bool  `yaml:"bundle"`
		BundleDir     string `yaml:"bundleDir"`
	}
)

// ToD2PluginOpts converts the config options to the JSON options of the d2 plugin.
// For json tags see d2.d2layouts.d2elklayout.DefaultOpts.
func (t toolConfigDiagramD2ELKOpts) ToD2PluginOpts() ([]byte, error) {
	out := map[string]any{
		"elk.algorithm":                 t.Algorithm,
		"spacing.nodeNodeBetweenLayers": t.NodeSpacing,
		"elk.padding":                   t.Padding,
		"spacing.edgeNodeBetweenLayers": t.EdgeSpacing,
		"elk.spacing.nodeSelfLoop":      t.SelfLoopSpacing,
	}
	return json.Marshal(out)
}

// ToD2PluginOpts converts the config options to the JSON options of the d2 plugin.
// For json tags see d2.d2layouts.d2dagrelayout.DefaultOpts
func (t toolConfigDiagramD2DagreOpts) ToD2PluginOpts() ([]byte, error) {
	out := map[string]any{
		"nodesep": t.NodeSep,
		"edgesep": t.EdgeSep,
	}
	return json.Marshal(out)
}

// loadConfig loads and parses the configuration file with the given baseName from the given file system.
func loadConfig(fileFS fs.FS, baseName string) (res toolConfig, err error) {
	f, err := fileFS.Open(baseName)
	if err != nil {
		return toolConfig{}, fmt.Errorf("open: %w", err)
	}
	defer f.Close()

	buf, err := io.ReadAll(f)
	if err != nil {
		return toolConfig{}, fmt.Errorf("read: %w", err)
	}

	if err = yaml.Unmarshal(buf, &res); err != nil {
		return toolConfig{}, fmt.Errorf("parse YAML: %w", err)
	}
	return
}

// mergeConfig merges the default configuration with the user-provided one.
func mergeConfig(defaultConf, userConf toolConfig) toolConfig {
	var res toolConfig

	res.ConfigVersion = coalesce(userConf.ConfigVersion, defaultConf.ConfigVersion)
	res.ProjectModule = coalesce(userConf.ProjectModule, defaultConf.ProjectModule)
	res.RuntimeModule = coalesce(userConf.RuntimeModule, defaultConf.RuntimeModule)
	res.TemplatesDir = coalesce(userConf.TemplatesDir, defaultConf.TemplatesDir)

	// *Replace* layout
	res.Code.Layout = defaultConf.Code.Layout
	if len(userConf.Code.Layout) > 0 {
		res.Code.Layout = userConf.Code.Layout
	}
	res.Code.TargetDir = coalesce(userConf.Code.TargetDir, defaultConf.Code.TargetDir)
	res.Code.PreambleTemplate = coalesce(userConf.Code.PreambleTemplate, defaultConf.Code.PreambleTemplate)
	res.Code.DisableFormatting = coalesce(userConf.Code.DisableFormatting, defaultConf.Code.DisableFormatting)

	// *Replace* the whole list
	res.Code.Implementation.Custom = defaultConf.Code.Implementation.Custom
	if len(userConf.Code.Implementation.Custom) > 0 {
		res.Code.Implementation.Custom = userConf.Code.Implementation.Custom
	}
	res.Code.Implementation.Directory = coalesce(userConf.Code.Implementation.Directory, defaultConf.Code.Implementation.Directory)
	res.Code.Implementation.Disable = coalesce(userConf.Code.Implementation.Disable, defaultConf.Code.Implementation.Disable)

	res.Code.Util.Directory = coalesce(userConf.Code.Util.Directory, defaultConf.Code.Util.Directory)
	// *Replace* the whole list
	res.Code.Util.Custom = defaultConf.Code.Util.Custom
	if len(userConf.Code.Util.Custom) > 0 {
		res.Code.Util.Custom = userConf.Code.Util.Custom
	}

	res.Locator.AllowRemoteReferences = coalesce(userConf.Locator.AllowRemoteReferences, defaultConf.Locator.AllowRemoteReferences)
	res.Locator.RootDirectory = coalesce(userConf.Locator.RootDirectory, defaultConf.Locator.RootDirectory)
	res.Locator.Timeout = coalesce(userConf.Locator.Timeout, defaultConf.Locator.Timeout)
	res.Locator.Command = coalesce(userConf.Locator.Command, defaultConf.Locator.Command)

	res.Client.GoModTemplate = coalesce(userConf.Client.GoModTemplate, defaultConf.Client.GoModTemplate)
	res.Client.OutputFile = coalesce(userConf.Client.OutputFile, defaultConf.Client.OutputFile)
	res.Client.OutputSourceFile = coalesce(userConf.Client.OutputSourceFile, defaultConf.Client.OutputSourceFile)
	res.Client.KeepSource = coalesce(userConf.Client.KeepSource, defaultConf.Client.KeepSource)

	res.Infra.Engine = coalesce(userConf.Infra.Engine, defaultConf.Infra.Engine)
	res.Infra.OutputFile = coalesce(userConf.Infra.OutputFile, defaultConf.Infra.OutputFile)
	res.Infra.ServerOpts = defaultConf.Infra.ServerOpts
	// *Replace* infra.servers
	if len(userConf.Infra.ServerOpts) > 0 {
		res.Infra.ServerOpts = userConf.Infra.ServerOpts
	}

	res.Diagram.Format = coalesce(userConf.Diagram.Format, defaultConf.Diagram.Format)
	res.Diagram.OutputFile = coalesce(userConf.Diagram.OutputFile, defaultConf.Diagram.OutputFile)
	res.Diagram.TargetDir = coalesce(userConf.Diagram.TargetDir, defaultConf.Diagram.TargetDir)
	res.Diagram.MultipleFiles = coalesce(userConf.Diagram.MultipleFiles, defaultConf.Diagram.MultipleFiles)
	res.Diagram.DisableFormatting = coalesce(userConf.Diagram.DisableFormatting, defaultConf.Diagram.DisableFormatting)
	res.Diagram.ServersCentric = coalesce(userConf.Diagram.ServersCentric, defaultConf.Diagram.ServersCentric)
	res.Diagram.ChannelsCentric = coalesce(userConf.Diagram.ChannelsCentric, defaultConf.Diagram.ChannelsCentric)
	res.Diagram.DocumentBorders = coalesce(userConf.Diagram.DocumentBorders, defaultConf.Diagram.DocumentBorders)
	// Diagram engine-specific options
	res.Diagram.D2.Engine = coalesce(userConf.Diagram.D2.Engine, defaultConf.Diagram.D2.Engine)
	res.Diagram.D2.Direction = coalesce(userConf.Diagram.D2.Direction, defaultConf.Diagram.D2.Direction)
	res.Diagram.D2.ThemeID = coalesce(userConf.Diagram.D2.ThemeID, defaultConf.Diagram.D2.ThemeID)
	res.Diagram.D2.DarkThemeID = coalesce(userConf.Diagram.D2.DarkThemeID, defaultConf.Diagram.D2.DarkThemeID)
	res.Diagram.D2.Pad = coalesce(userConf.Diagram.D2.Pad, defaultConf.Diagram.D2.Pad)
	res.Diagram.D2.Sketch = coalesce(userConf.Diagram.D2.Sketch, defaultConf.Diagram.D2.Sketch)
	res.Diagram.D2.Center = coalesce(userConf.Diagram.D2.Center, defaultConf.Diagram.D2.Center)
	res.Diagram.D2.Scale = coalesce(userConf.Diagram.D2.Scale, defaultConf.Diagram.D2.Scale)

	res.Diagram.D2.ELK.Algorithm = coalesce(userConf.Diagram.D2.ELK.Algorithm, defaultConf.Diagram.D2.ELK.Algorithm)
	res.Diagram.D2.ELK.NodeSpacing = coalesce(userConf.Diagram.D2.ELK.NodeSpacing, defaultConf.Diagram.D2.ELK.NodeSpacing)
	res.Diagram.D2.ELK.Padding = coalesce(userConf.Diagram.D2.ELK.Padding, defaultConf.Diagram.D2.ELK.Padding)
	res.Diagram.D2.ELK.EdgeSpacing = coalesce(userConf.Diagram.D2.ELK.EdgeSpacing, defaultConf.Diagram.D2.ELK.EdgeSpacing)
	res.Diagram.D2.ELK.SelfLoopSpacing = coalesce(userConf.Diagram.D2.ELK.SelfLoopSpacing, defaultConf.Diagram.D2.ELK.SelfLoopSpacing)

	res.Diagram.D2.Dagre.NodeSep = coalesce(userConf.Diagram.D2.Dagre.NodeSep, defaultConf.Diagram.D2.Dagre.NodeSep)
	res.Diagram.D2.Dagre.EdgeSep = coalesce(userConf.Diagram.D2.Dagre.EdgeSep, defaultConf.Diagram.D2.Dagre.EdgeSep)

	res.UI.OutputFile = coalesce(userConf.UI.OutputFile, defaultConf.UI.OutputFile)
	res.UI.Listen = coalesce(userConf.UI.Listen, defaultConf.UI.Listen)
	res.UI.ListenAddress = coalesce(userConf.UI.ListenAddress, defaultConf.UI.ListenAddress)
	res.UI.ListenPath = coalesce(userConf.UI.ListenPath, defaultConf.UI.ListenPath)
	res.UI.ReferenceOnly = coalesce(userConf.UI.ReferenceOnly, defaultConf.UI.ReferenceOnly)
	res.UI.Bundle = coalesce(userConf.UI.Bundle, defaultConf.UI.Bundle)
	res.UI.BundleDir = coalesce(userConf.UI.BundleDir, defaultConf.UI.BundleDir)

	return res
}
