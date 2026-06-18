package common2

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

type D2DiagramEngine string

const (
	D2DiagramEngineELK   D2DiagramEngine = "elk"
	D2DiagramEngineDagre D2DiagramEngine = "dagre"
)

// Structures, that represent the tool's configuration file
type (
	ToolConfig struct {
		ConfigVersion int    `yaml:"configVersion"`
		ProjectModule string `yaml:"projectModule"`
		RuntimeModule string `yaml:"runtimeModule"`
		TemplatesDir  string `yaml:"templatesDir"`
		Quiet         bool   `yaml:"quiet"`

		Locator ToolConfigLocator `yaml:"locator"`

		Code    ToolConfigCode    `yaml:"code"`
		Client  ToolConfigClient  `yaml:"client"`
		Infra   ToolConfigInfra   `yaml:"infra"`
		Diagram ToolConfigDiagram `yaml:"diagram"`
		UI      ToolConfigUI      `yaml:"ui"`
		Doc     ToolConfigDoc     `yaml:"doc"`
	}

	ToolConfigLocator struct {
		AllowRemoteReferences bool          `yaml:"allowRemoteReferences"`
		RootDirectory         string        `yaml:"rootDirectory"`
		Timeout               time.Duration `yaml:"timeout"`
		Command               string        `yaml:"command"`
	}

	ToolConfigCode struct {
		OnlyPublish       bool   `yaml:"onlyPublish"`
		OnlySubscribe     bool   `yaml:"onlySubscribe"`
		DisableFormatting bool   `yaml:"disableFormatting"`
		TargetDir         string `yaml:"targetDir"`

		Layout []ToolConfigCodeLayout `yaml:"layout"`

		PreambleTemplate string `yaml:"preambleTemplate"`

		Util           ToolConfigCodeUtil           `yaml:"util"`
		Implementation ToolConfigCodeImplementation `yaml:"implementation"`
	}

	ToolConfigCodeLayout struct {
		NameRe        string                 `yaml:"nameRe"`
		ArtifactKinds []string               `yaml:"artifactKinds"`
		ModuleURLRe   string                 `yaml:"moduleURLRe"` // TODO: rename to locationRe or smth like that
		PathRe        string                 `yaml:"pathRe"`      // TODO: remove? almost duplicate of moduleURLRe
		Protocols     []string               `yaml:"protocols"`
		Not           bool                   `yaml:"not"` // Inverts the match, i.e. NOT operation
		Render        ToolConfigLayoutRender `yaml:"render"`
	}

	ToolConfigLayoutRender struct {
		Protocols []string `yaml:"protocols"`
		Template  string   `yaml:"template"`
		File      string   `yaml:"file"`
		Package   string   `yaml:"package"` // TODO: make it inline template
	}

	ToolConfigCodeUtil struct {
		Directory string                       `yaml:"directory"` // Template expression, relative to the target directory
		Custom    []ToolConfigCodeUtilProtocol `yaml:"custom"`
	}

	ToolConfigCodeUtilProtocol struct {
		Protocol          string `yaml:"protocol"`
		TemplateDirectory string `yaml:"templateDirectory"`
	}

	ToolConfigCodeImplementation struct {
		Directory string                             `yaml:"directory"` // Template expression, relative to the target directory
		Disable   bool                               `yaml:"disable"`
		Custom    []ToolConfigImplementationProtocol `yaml:"custom"`
	}

	ToolConfigImplementationProtocol struct {
		Protocol          string `yaml:"protocol"`
		Name              string `yaml:"name"`
		Disable           bool   `yaml:"disable"`
		TemplateDirectory string `yaml:"templateDirectory"`
		Package           string `yaml:"package"`
	}

	ToolConfigClient struct {
		OutputFile       string `yaml:"outputFile"`
		OutputSourceFile string `yaml:"outputSourceFile"`
		KeepSource       bool   `yaml:"keepSource"`
		GoModTemplate    string `yaml:"goModTemplate"`
		TempDir          string `yaml:"tempDir"`
	}

	ToolConfigInfra struct {
		ServerOpts []ToolConfigInfraServerOpt `yaml:"serverOpts"`
		Engine     string                     `yaml:"engine"`
		OutputFile string                     `yaml:"outputFile"`
	}

	ToolConfigInfraServerOpt struct {
		ServerName string                                                                             `yaml:"serverName"` // TODO: make required
		Variables  types.Union2[types.OrderedMap[string, string], []types.OrderedMap[string, string]] `yaml:"variables"`
	}

	ToolConfigDiagram struct {
		Format common.DiagramOutputFormat `yaml:"format"`

		OutputFile        string `yaml:"outputFile"`
		TargetDir         string `yaml:"targetDir"`
		MultipleFiles     bool   `yaml:"multipleFiles"`
		DisableFormatting bool   `yaml:"disableFormatting"`

		ChannelsCentric bool `yaml:"channelsCentric"`
		ServersCentric  bool `yaml:"serversCentric"`
		DocumentBorders bool `yaml:"documentBorders"`

		D2 ToolConfigDiagramD2Opts `yaml:"d2"`
	}

	ToolConfigDiagramD2Opts struct {
		Engine      D2DiagramEngine              `yaml:"engine"`
		Direction   common.D2DiagramDirection    `yaml:"direction"`
		ThemeID     *int64                       `yaml:"themeId"`
		DarkThemeID *int64                       `yaml:"darkThemeId"`
		Pad         *int64                       `yaml:"pad"`
		Sketch      *bool                        `yaml:"sketch"`
		Center      *bool                        `yaml:"center"`
		Scale       *float64                     `yaml:"scale"`
		ELK         ToolConfigDiagramD2ELKOpts   `yaml:"elk"`
		Dagre       ToolConfigDiagramD2DagreOpts `yaml:"dagre"`
	}

	ToolConfigDiagramD2ELKOpts struct {
		Algorithm       string `yaml:"algorithm"`
		NodeSpacing     int64  `yaml:"nodeSpacing"`
		Padding         string `yaml:"padding"`
		EdgeSpacing     int64  `yaml:"edgeSpacing"`
		SelfLoopSpacing int64  `yaml:"selfLoopSpacing"`
	}

	ToolConfigDiagramD2DagreOpts struct {
		NodeSep int64 `yaml:"nodeSep"`
		EdgeSep int64 `yaml:"edgeSep"`
	}

	ToolConfigUI struct {
		OutputFile string `yaml:"outputFile"`

		Listen        *bool  `yaml:"listen"`
		ListenAddress string `yaml:"listenAddress"`
		ListenPath    string `yaml:"listenPath"`
		Bundle        *bool  `yaml:"bundle"`
		BundleDir     string `yaml:"bundleDir"`
	}

	ToolConfigDoc struct {
		Cp       ToolConfigDocCp       `yaml:"cp"`
		Validate ToolConfigDocValidate `yaml:"validate"`
		Flatten  ToolConfigDocFlatten  `yaml:"flatten"`
		Indent   int                   `yaml:"indent"`
		Format   string                `yaml:"format"`
	}

	ToolConfigDocValidate struct {
		Schema string `yaml:"schema"`
	}

	ToolConfigDocFlatten struct {
		OutputFile   string `yaml:"outputFile"`
		WithExternal bool   `yaml:"withExternal"`
		WithRemote   bool   `yaml:"withRemote"`
	}

	ToolConfigDocCp struct {
		Recursive        bool `yaml:"recursive"`
		Shallow          bool `yaml:"shallow"`
		Headless         bool `yaml:"headless"`
		Force            bool `yaml:"force"`
		Interactive      bool `yaml:"interactive"`
		DisableRewriting bool `yaml:"disableRewriting"`
	}
)

// ToD2PluginOpts converts the config options to the JSON options of the d2 plugin.
// For json tags see d2.d2layouts.d2elklayout.DefaultOpts.
func (t ToolConfigDiagramD2ELKOpts) ToD2PluginOpts() ([]byte, error) {
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
func (t ToolConfigDiagramD2DagreOpts) ToD2PluginOpts() ([]byte, error) {
	out := map[string]any{
		"nodesep": t.NodeSep,
		"edgesep": t.EdgeSep,
	}
	return json.Marshal(out)
}

// LoadConfig loads and parses the configuration file with the given baseName from the given file system.
func LoadConfig(fileFS fs.FS, baseName string) (res ToolConfig, err error) {
	f, err := fileFS.Open(baseName)
	if err != nil {
		return ToolConfig{}, fmt.Errorf("open: %w", err)
	}
	defer f.Close()

	buf, err := io.ReadAll(f)
	if err != nil {
		return ToolConfig{}, fmt.Errorf("read: %w", err)
	}

	if err = yaml.Unmarshal(buf, &res); err != nil {
		return ToolConfig{}, fmt.Errorf("parse YAML: %w", err)
	}
	return
}

// MergeConfig merges the default configuration with the user-provided one.
func MergeConfig(defaultConf, userConf ToolConfig) ToolConfig {
	var res ToolConfig

	res.ConfigVersion = Coalesce(userConf.ConfigVersion, defaultConf.ConfigVersion)
	res.ProjectModule = Coalesce(userConf.ProjectModule, defaultConf.ProjectModule)
	res.RuntimeModule = Coalesce(userConf.RuntimeModule, defaultConf.RuntimeModule)
	res.TemplatesDir = Coalesce(userConf.TemplatesDir, defaultConf.TemplatesDir)
	res.Quiet = Coalesce(userConf.Quiet, defaultConf.Quiet)

	// *Replace* layout
	res.Code.Layout = defaultConf.Code.Layout
	if len(userConf.Code.Layout) > 0 {
		res.Code.Layout = userConf.Code.Layout
	}
	res.Code.OnlyPublish = Coalesce(userConf.Code.OnlyPublish, defaultConf.Code.OnlyPublish)
	res.Code.OnlySubscribe = Coalesce(userConf.Code.OnlySubscribe, defaultConf.Code.OnlySubscribe)
	res.Code.DisableFormatting = Coalesce(userConf.Code.DisableFormatting, defaultConf.Code.DisableFormatting)
	res.Code.TargetDir = Coalesce(userConf.Code.TargetDir, defaultConf.Code.TargetDir)
	res.Code.PreambleTemplate = Coalesce(userConf.Code.PreambleTemplate, defaultConf.Code.PreambleTemplate)

	// *Replace* the whole list
	res.Code.Implementation.Custom = defaultConf.Code.Implementation.Custom
	if len(userConf.Code.Implementation.Custom) > 0 {
		res.Code.Implementation.Custom = userConf.Code.Implementation.Custom
	}
	res.Code.Implementation.Directory = Coalesce(userConf.Code.Implementation.Directory, defaultConf.Code.Implementation.Directory)
	res.Code.Implementation.Disable = Coalesce(userConf.Code.Implementation.Disable, defaultConf.Code.Implementation.Disable)

	res.Code.Util.Directory = Coalesce(userConf.Code.Util.Directory, defaultConf.Code.Util.Directory)
	// *Replace* the whole list
	res.Code.Util.Custom = defaultConf.Code.Util.Custom
	if len(userConf.Code.Util.Custom) > 0 {
		res.Code.Util.Custom = userConf.Code.Util.Custom
	}

	res.Locator.AllowRemoteReferences = Coalesce(userConf.Locator.AllowRemoteReferences, defaultConf.Locator.AllowRemoteReferences)
	res.Locator.RootDirectory = Coalesce(userConf.Locator.RootDirectory, defaultConf.Locator.RootDirectory)
	res.Locator.Timeout = Coalesce(userConf.Locator.Timeout, defaultConf.Locator.Timeout)
	res.Locator.Command = Coalesce(userConf.Locator.Command, defaultConf.Locator.Command)

	res.Client.GoModTemplate = Coalesce(userConf.Client.GoModTemplate, defaultConf.Client.GoModTemplate)
	res.Client.OutputFile = Coalesce(userConf.Client.OutputFile, defaultConf.Client.OutputFile)
	res.Client.OutputSourceFile = Coalesce(userConf.Client.OutputSourceFile, defaultConf.Client.OutputSourceFile)
	res.Client.KeepSource = Coalesce(userConf.Client.KeepSource, defaultConf.Client.KeepSource)

	res.Infra.Engine = Coalesce(userConf.Infra.Engine, defaultConf.Infra.Engine)
	res.Infra.OutputFile = Coalesce(userConf.Infra.OutputFile, defaultConf.Infra.OutputFile)
	res.Infra.ServerOpts = defaultConf.Infra.ServerOpts
	// *Replace* infra.servers
	if len(userConf.Infra.ServerOpts) > 0 {
		res.Infra.ServerOpts = userConf.Infra.ServerOpts
	}

	res.Diagram.Format = Coalesce(userConf.Diagram.Format, defaultConf.Diagram.Format)
	res.Diagram.OutputFile = Coalesce(userConf.Diagram.OutputFile, defaultConf.Diagram.OutputFile)
	res.Diagram.TargetDir = Coalesce(userConf.Diagram.TargetDir, defaultConf.Diagram.TargetDir)
	res.Diagram.MultipleFiles = Coalesce(userConf.Diagram.MultipleFiles, defaultConf.Diagram.MultipleFiles)
	res.Diagram.DisableFormatting = Coalesce(userConf.Diagram.DisableFormatting, defaultConf.Diagram.DisableFormatting)
	res.Diagram.ServersCentric = Coalesce(userConf.Diagram.ServersCentric, defaultConf.Diagram.ServersCentric)
	res.Diagram.ChannelsCentric = Coalesce(userConf.Diagram.ChannelsCentric, defaultConf.Diagram.ChannelsCentric)
	res.Diagram.DocumentBorders = Coalesce(userConf.Diagram.DocumentBorders, defaultConf.Diagram.DocumentBorders)
	// Diagram engine-specific options
	res.Diagram.D2.Engine = Coalesce(userConf.Diagram.D2.Engine, defaultConf.Diagram.D2.Engine)
	res.Diagram.D2.Direction = Coalesce(userConf.Diagram.D2.Direction, defaultConf.Diagram.D2.Direction)
	res.Diagram.D2.ThemeID = Coalesce(userConf.Diagram.D2.ThemeID, defaultConf.Diagram.D2.ThemeID)
	res.Diagram.D2.DarkThemeID = Coalesce(userConf.Diagram.D2.DarkThemeID, defaultConf.Diagram.D2.DarkThemeID)
	res.Diagram.D2.Pad = Coalesce(userConf.Diagram.D2.Pad, defaultConf.Diagram.D2.Pad)
	res.Diagram.D2.Sketch = Coalesce(userConf.Diagram.D2.Sketch, defaultConf.Diagram.D2.Sketch)
	res.Diagram.D2.Center = Coalesce(userConf.Diagram.D2.Center, defaultConf.Diagram.D2.Center)
	res.Diagram.D2.Scale = Coalesce(userConf.Diagram.D2.Scale, defaultConf.Diagram.D2.Scale)

	res.Diagram.D2.ELK.Algorithm = Coalesce(userConf.Diagram.D2.ELK.Algorithm, defaultConf.Diagram.D2.ELK.Algorithm)
	res.Diagram.D2.ELK.NodeSpacing = Coalesce(userConf.Diagram.D2.ELK.NodeSpacing, defaultConf.Diagram.D2.ELK.NodeSpacing)
	res.Diagram.D2.ELK.Padding = Coalesce(userConf.Diagram.D2.ELK.Padding, defaultConf.Diagram.D2.ELK.Padding)
	res.Diagram.D2.ELK.EdgeSpacing = Coalesce(userConf.Diagram.D2.ELK.EdgeSpacing, defaultConf.Diagram.D2.ELK.EdgeSpacing)
	res.Diagram.D2.ELK.SelfLoopSpacing = Coalesce(userConf.Diagram.D2.ELK.SelfLoopSpacing, defaultConf.Diagram.D2.ELK.SelfLoopSpacing)

	res.Diagram.D2.Dagre.NodeSep = Coalesce(userConf.Diagram.D2.Dagre.NodeSep, defaultConf.Diagram.D2.Dagre.NodeSep)
	res.Diagram.D2.Dagre.EdgeSep = Coalesce(userConf.Diagram.D2.Dagre.EdgeSep, defaultConf.Diagram.D2.Dagre.EdgeSep)

	res.UI.OutputFile = Coalesce(userConf.UI.OutputFile, defaultConf.UI.OutputFile)
	res.UI.Listen = Coalesce(userConf.UI.Listen, defaultConf.UI.Listen)
	res.UI.ListenAddress = Coalesce(userConf.UI.ListenAddress, defaultConf.UI.ListenAddress)
	res.UI.ListenPath = Coalesce(userConf.UI.ListenPath, defaultConf.UI.ListenPath)
	res.UI.Bundle = Coalesce(userConf.UI.Bundle, defaultConf.UI.Bundle)
	res.UI.BundleDir = Coalesce(userConf.UI.BundleDir, defaultConf.UI.BundleDir)

	res.Doc.Indent = Coalesce(userConf.Doc.Indent, defaultConf.Doc.Indent)
	res.Doc.Format = Coalesce(userConf.Doc.Format, defaultConf.Doc.Format)
	res.Doc.Validate.Schema = Coalesce(userConf.Doc.Validate.Schema, defaultConf.Doc.Validate.Schema)
	res.Doc.Flatten.OutputFile = Coalesce(userConf.Doc.Flatten.OutputFile, defaultConf.Doc.Flatten.OutputFile)
	res.Doc.Flatten.WithExternal = Coalesce(userConf.Doc.Flatten.WithExternal, defaultConf.Doc.Flatten.WithExternal)
	res.Doc.Flatten.WithRemote = Coalesce(userConf.Doc.Flatten.WithRemote, defaultConf.Doc.Flatten.WithRemote)
	res.Locator.AllowRemoteReferences = Coalesce(res.Locator.AllowRemoteReferences, res.Doc.Flatten.WithRemote)
	res.Doc.Cp.Shallow = Coalesce(userConf.Doc.Cp.Shallow, defaultConf.Doc.Cp.Shallow)
	res.Doc.Cp.Recursive = Coalesce(userConf.Doc.Cp.Recursive, defaultConf.Doc.Cp.Recursive)
	res.Doc.Cp.Headless = Coalesce(userConf.Doc.Cp.Headless, defaultConf.Doc.Cp.Headless)
	res.Doc.Cp.Force = Coalesce(userConf.Doc.Cp.Force, defaultConf.Doc.Cp.Force)
	res.Doc.Cp.Interactive = Coalesce(userConf.Doc.Cp.Interactive, defaultConf.Doc.Cp.Interactive)
	res.Doc.Cp.DisableRewriting = Coalesce(userConf.Doc.Cp.DisableRewriting, defaultConf.Doc.Cp.DisableRewriting)

	return res
}
