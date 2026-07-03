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

type DocNodesPathStyle string

const (
	DocNodesPathStyleJSONPointer  = "json-pointer"
	DocNodesPathStyleYq           = "yq"
	DocNodesPathStyleHuman        = "human"
	DocNodesPathStyleHumanNoColor = "human-no-color"
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
		Cp          ToolConfigDocCp          `yaml:"cp"`
		Mv          ToolConfigDocMv          `yaml:"mv"`
		Validate    ToolConfigDocValidate    `yaml:"validate"`
		Flatten     ToolConfigDocFlatten     `yaml:"flatten"`
		GenExamples ToolConfigDocGenExamples `yaml:"genExamples"`
		Nodes       ToolConfigDocNodes       `yaml:"nodes"`
		Deps        ToolConfigDocDeps        `yaml:"deps"`
	}

	ToolConfigDocCp struct {
		FollowRefs       bool   `yaml:"followRefs"`
		ShallowRefs      bool   `yaml:"shallowRefs"`
		Headless         bool   `yaml:"headless"`
		Force            bool   `yaml:"force"`
		Interactive      bool   `yaml:"interactive"`
		DisableRewriting bool   `yaml:"disableRewriting"`
		Indent           int    `yaml:"indent"`
		Format           string `yaml:"format"`
	}

	ToolConfigDocMv struct {
		FollowRefs       bool   `yaml:"followRefs"`
		ShallowRefs      bool   `yaml:"shallowRefs"`
		Headless         bool   `yaml:"headless"`
		Force            bool   `yaml:"force"`
		Interactive      bool   `yaml:"interactive"`
		DisableRewriting bool   `yaml:"disableRewriting"`
		Indent           int    `yaml:"indent"`
		Format           string `yaml:"format"`
	}

	ToolConfigDocValidate struct {
		Schema string `yaml:"schema"`
	}

	ToolConfigDocFlatten struct {
		OutputFile   string `yaml:"outputFile"`
		ExternalRefs bool   `yaml:"externalRefs"`
		RemoteRefs   bool   `yaml:"remoteRefs"`
		Indent       int    `yaml:"indent"`
		Format       string `yaml:"format"`
	}

	ToolConfigDocGenExamples struct {
		OutputFile            string `yaml:"outputFile"`
		OnlyMessages          bool   `yaml:"onlyMessages"`
		OnlySchemas           bool   `yaml:"onlySchemas"`
		Append                bool   `yaml:"append"`
		Count                 int    `yaml:"count"`
		AllowRemoteReferences bool   `yaml:"allowRemoteReferences"`
		DateFormat            string `yaml:"dateFormat"`
		TimeFormat            string `yaml:"timeFormat"`
		DateTimeFormat        string `yaml:"dateTimeFormat"`
		Indent                int    `yaml:"indent"`
		Format                string `yaml:"format"`
	}

	ToolConfigDocNodes struct {
		Entities              string            `yaml:"entities"`
		Expand                bool              `yaml:"expand"`
		ExpandAll             bool              `yaml:"expandAll"`
		Components            bool              `yaml:"components"`
		TopLevel              bool              `yaml:"topLevel"`
		FollowExternalRefs    bool              `yaml:"followExternalRefs"`
		AllowRemoteReferences bool              `yaml:"allowRemoteReferences"`
		Tree                  bool              `yaml:"tree"`
		EntryStyle            DocNodesPathStyle `yaml:"entryStyle"`
	}

	ToolConfigDocDeps struct {
		Tree                  bool `yaml:"tree"`
		AllowRemoteReferences bool `yaml:"allowRemoteReferences"`
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

	res.Doc.Validate.Schema = Coalesce(userConf.Doc.Validate.Schema, defaultConf.Doc.Validate.Schema)
	res.Doc.Flatten.OutputFile = Coalesce(userConf.Doc.Flatten.OutputFile, defaultConf.Doc.Flatten.OutputFile)
	res.Doc.Flatten.ExternalRefs = Coalesce(userConf.Doc.Flatten.ExternalRefs, defaultConf.Doc.Flatten.ExternalRefs)
	res.Doc.Flatten.RemoteRefs = Coalesce(userConf.Doc.Flatten.RemoteRefs, defaultConf.Doc.Flatten.RemoteRefs)
	res.Doc.Flatten.Indent = Coalesce(userConf.Doc.Flatten.Indent, defaultConf.Doc.Flatten.Indent)
	res.Doc.Flatten.Format = Coalesce(userConf.Doc.Flatten.Format, defaultConf.Doc.Flatten.Format)
	res.Locator.AllowRemoteReferences = Coalesce(res.Locator.AllowRemoteReferences, res.Doc.Flatten.RemoteRefs) // FIXME: rmeove and make a separate config param in locator constructor
	res.Doc.Cp.ShallowRefs = Coalesce(userConf.Doc.Cp.ShallowRefs, defaultConf.Doc.Cp.ShallowRefs)
	res.Doc.Cp.FollowRefs = Coalesce(userConf.Doc.Cp.FollowRefs, defaultConf.Doc.Cp.FollowRefs)
	res.Doc.Cp.Headless = Coalesce(userConf.Doc.Cp.Headless, defaultConf.Doc.Cp.Headless)
	res.Doc.Cp.Force = Coalesce(userConf.Doc.Cp.Force, defaultConf.Doc.Cp.Force)
	res.Doc.Cp.Interactive = Coalesce(userConf.Doc.Cp.Interactive, defaultConf.Doc.Cp.Interactive)
	res.Doc.Cp.DisableRewriting = Coalesce(userConf.Doc.Cp.DisableRewriting, defaultConf.Doc.Cp.DisableRewriting)
	res.Doc.Cp.Indent = Coalesce(userConf.Doc.Cp.Indent, defaultConf.Doc.Cp.Indent)
	res.Doc.Cp.Format = Coalesce(userConf.Doc.Cp.Format, defaultConf.Doc.Cp.Format)
	res.Doc.Mv.ShallowRefs = Coalesce(userConf.Doc.Mv.ShallowRefs, defaultConf.Doc.Mv.ShallowRefs)
	res.Doc.Mv.FollowRefs = Coalesce(userConf.Doc.Mv.FollowRefs, defaultConf.Doc.Mv.FollowRefs)
	res.Doc.Mv.Headless = Coalesce(userConf.Doc.Mv.Headless, defaultConf.Doc.Mv.Headless)
	res.Doc.Mv.Force = Coalesce(userConf.Doc.Mv.Force, defaultConf.Doc.Mv.Force)
	res.Doc.Mv.Interactive = Coalesce(userConf.Doc.Mv.Interactive, defaultConf.Doc.Mv.Interactive)
	res.Doc.Mv.DisableRewriting = Coalesce(userConf.Doc.Mv.DisableRewriting, defaultConf.Doc.Mv.DisableRewriting)
	res.Doc.Mv.Indent = Coalesce(userConf.Doc.Mv.Indent, defaultConf.Doc.Mv.Indent)
	res.Doc.Mv.Format = Coalesce(userConf.Doc.Mv.Format, defaultConf.Doc.Mv.Format)
	res.Doc.GenExamples.OutputFile = Coalesce(userConf.Doc.GenExamples.OutputFile, defaultConf.Doc.GenExamples.OutputFile)
	res.Doc.GenExamples.OnlyMessages = Coalesce(userConf.Doc.GenExamples.OnlyMessages, defaultConf.Doc.GenExamples.OnlyMessages)
	res.Doc.GenExamples.OnlySchemas = Coalesce(userConf.Doc.GenExamples.OnlySchemas, defaultConf.Doc.GenExamples.OnlySchemas)
	res.Doc.GenExamples.Append = Coalesce(userConf.Doc.GenExamples.Append, defaultConf.Doc.GenExamples.Append)
	res.Doc.GenExamples.Count = Coalesce(userConf.Doc.GenExamples.Count, defaultConf.Doc.GenExamples.Count)
	res.Doc.GenExamples.DateFormat = Coalesce(userConf.Doc.GenExamples.DateFormat, defaultConf.Doc.GenExamples.DateFormat)
	res.Doc.GenExamples.TimeFormat = Coalesce(userConf.Doc.GenExamples.TimeFormat, defaultConf.Doc.GenExamples.TimeFormat)
	res.Doc.GenExamples.DateTimeFormat = Coalesce(userConf.Doc.GenExamples.DateTimeFormat, defaultConf.Doc.GenExamples.DateTimeFormat)
	res.Doc.GenExamples.Indent = Coalesce(userConf.Doc.GenExamples.Indent, defaultConf.Doc.GenExamples.Indent)
	res.Doc.GenExamples.Format = Coalesce(userConf.Doc.GenExamples.Format, defaultConf.Doc.GenExamples.Format)
	res.Locator.AllowRemoteReferences = Coalesce(res.Locator.AllowRemoteReferences, res.Doc.GenExamples.AllowRemoteReferences) // FIXME: rmeove and make a separate config param in locator constructor
	res.Doc.Nodes.Entities = Coalesce(userConf.Doc.Nodes.Entities, defaultConf.Doc.Nodes.Entities)
	res.Doc.Nodes.Expand = Coalesce(userConf.Doc.Nodes.Expand, defaultConf.Doc.Nodes.Expand)
	res.Doc.Nodes.ExpandAll = Coalesce(userConf.Doc.Nodes.ExpandAll, defaultConf.Doc.Nodes.ExpandAll)
	res.Doc.Nodes.Components = Coalesce(userConf.Doc.Nodes.Components, defaultConf.Doc.Nodes.Components)
	res.Doc.Nodes.TopLevel = Coalesce(userConf.Doc.Nodes.TopLevel, defaultConf.Doc.Nodes.TopLevel)
	res.Doc.Nodes.FollowExternalRefs = Coalesce(userConf.Doc.Nodes.FollowExternalRefs, defaultConf.Doc.Nodes.FollowExternalRefs)
	res.Doc.Nodes.AllowRemoteReferences = Coalesce(userConf.Doc.Nodes.AllowRemoteReferences, defaultConf.Doc.Nodes.AllowRemoteReferences)
	res.Doc.Nodes.Tree = Coalesce(userConf.Doc.Nodes.Tree, defaultConf.Doc.Nodes.Tree)
	res.Doc.Nodes.EntryStyle = Coalesce(userConf.Doc.Nodes.EntryStyle, defaultConf.Doc.Nodes.EntryStyle)
	res.Doc.Deps.Tree = Coalesce(userConf.Doc.Deps.Tree, defaultConf.Doc.Deps.Tree)
	res.Locator.AllowRemoteReferences = Coalesce(res.Locator.AllowRemoteReferences, res.Doc.Deps.AllowRemoteReferences) // FIXME: rmeove and make a separate config param in locator constructor

	return res
}
