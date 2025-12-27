package renderer

import (
	"fmt"
	"io/fs"
	"os"
	"path"
	"text/template"

	"github.com/bdragon300/go-asyncapi/internal/common"
	"github.com/bdragon300/go-asyncapi/internal/log"
	"github.com/bdragon300/go-asyncapi/internal/tmpl"
	"github.com/bdragon300/go-asyncapi/internal/tmpl/manager"
	"github.com/bdragon300/go-asyncapi/internal/utils"
	"github.com/bdragon300/go-asyncapi/templates/codeextra"
	"github.com/samber/lo"
	"gopkg.in/yaml.v3"
)

const (
	extraCodeSrcDir       = "extra"
	unknownProtocolSrcDir = "unknown"
)

type dirTemplateLoader interface {
	ParseDir(subDir string, renderManager *manager.TemplateRenderManager) ([]string, error)
	LoadTemplate(name string) (*template.Template, error)
}

func RenderExtraCode(protocols []string, opts common.ConfigCodeExtraOpts, mng *manager.TemplateRenderManager, tplBase fs.FS) error {
	logger := log.GetLogger(log.LoggerPrefixRendering)
	tplLoader := mng.TemplateLoader.(dirTemplateLoader)

	for _, protocol := range protocols {
		logger.Debug("Render extra code", "protocol", protocol)
		ctx := tmpl.CodeExtraTemplateContext{LayoutOpts: opts, Protocol: protocol}
		directory, err := renderInlineTemplate(opts.Directory, ctx, mng)
		if err != nil {
			return fmt.Errorf("render directory expression: %w", err)
		}
		logger.Trace("-> Directory", "result", directory)

		pkgName := utils.GetPackageName(directory)
		ctx = tmpl.CodeExtraTemplateContext{LayoutOpts: opts, Protocol: protocol, Directory: directory, PackageName: pkgName}
		logger.Trace("-> Package name", "name", pkgName)

		srcDir, found := getExtraSourceDir(tplBase, protocol)
		if !found {
			logger.Warn("-> Unknown protocol found, generate the fallback code", "protocol", protocol)
		}
		logger.Trace("-> Extra code source directory", "srcDir", srcDir)

		templates, err := tplLoader.ParseDir(srcDir, mng)
		if err != nil {
			return err
		}
		logger.Trace("-> Templates found", "files", templates)
		if err = renderCodeExtraTemplates(templates, ctx, mng, nil); err != nil {
			return fmt.Errorf("render templates for protocol %q: %w", protocol, err)
		}
	}
	return nil
}

func RenderImplementationsCode(protocols []string, opts common.ConfigCodeExtraOpts, mng *manager.TemplateRenderManager, tplBase fs.FS) error {
	logger := log.GetLogger(log.LoggerPrefixRendering)
	prevTplLoader := mng.TemplateLoader
	defer func() {
		mng.TemplateLoader = prevTplLoader
	}()
	manifests := lo.Must(LoadImplementationsManifests(tplBase))

	for _, protocol := range protocols {
		logger.Debug("Render implementation code", "protocol", protocol)

		conf, _ := lo.Find(opts.Implementations, func(item common.ConfigImplementationProtocol) bool {
			return item.Protocol == protocol
		})
		if conf.Disable {
			logger.Debug("-> Implementation is disabled for protocol, skipping")
			continue
		}

		ctx := tmpl.CodeExtraTemplateContext{LayoutOpts: opts, Protocol: protocol}
		dirExpr, _ := lo.Coalesce(conf.Directory, opts.Directory)
		directory, err := renderInlineTemplate(dirExpr, ctx, mng)
		if err != nil {
			return fmt.Errorf("render directory expression: %w", err)
		}
		logger.Trace("-> Directory", "result", directory)

		pkgName, _ := lo.Coalesce(conf.Package, utils.GetPackageName(directory))
		ctx = tmpl.CodeExtraTemplateContext{LayoutOpts: opts, Protocol: protocol, Directory: directory, PackageName: pkgName}
		logger.Trace("-> Package name", "name", pkgName)

		var templates []string
		if conf.TemplateDirectory != "" {
			logger.Trace("-> Using the custom template directory", "directory", conf.TemplateDirectory)
			ld := tmpl.NewTemplateLoader("", os.DirFS(conf.TemplateDirectory))
			mng.TemplateLoader = ld

			templates, err = ld.ParseDir(".", mng)
			if err != nil {
				return fmt.Errorf("parse templates from directory %q: %w", conf.TemplateDirectory, err)
			}
		} else {
			logger.Trace("-> Using the built-in implementations", "name", lo.CoalesceOrEmpty(conf.Name, "<default>"))
			ld := tmpl.NewTemplateLoader("", tplBase)
			mng.TemplateLoader = ld

			man, found := lo.Find(manifests, func(item codeextra.ImplementationManifest) bool {
				return item.Protocol == protocol && (conf.Name == "" || item.Name == conf.Name)
			})
			if !found {
				logger.Warn("-> No implementation found for protocol, skipping", "protocol", protocol)
				continue
			}
			logger.Debug("-> Using built-in implementation", "protocol", protocol)
			ctx.Manifest = &man

			templates, err = ld.ParseDir(man.Dir, mng)
			if err != nil {
				return fmt.Errorf("parse templates from built-in implementation %q, this is a bug: %w", man.Name, err)
			}
		}

		logger.Trace("-> Templates found", "files", templates)
		if err = renderCodeExtraTemplates(templates, ctx, mng, &conf); err != nil {
			return fmt.Errorf("render templates for protocol %q: %w", protocol, err)
		}
	}
	return nil
}

func renderCodeExtraTemplates(templates []string, ctx tmpl.CodeExtraTemplateContext, mng *manager.TemplateRenderManager, implConf *common.ConfigImplementationProtocol) error {
	logger := log.GetLogger(log.LoggerPrefixRendering)

	for _, templateFile := range templates {
		normFileName := utils.ToGoFilePath(path.Join(ctx.Directory, templateFile))
		mng.BeginFile(normFileName, ctx.PackageName)
		mng.ExtraCodeProtocol = ctx.Protocol
		if implConf != nil {
			mng.ImplementationConfig = implConf
			mng.ImplementationManifest = ctx.Manifest
		}

		logger.Debug("-> Render file", "name", normFileName)
		tpl, err := mng.TemplateLoader.LoadTemplate(templateFile)
		if err != nil {
			return fmt.Errorf("load template %q: %w", templateFile, err)
		}
		if err := tpl.Execute(mng.Buffer, ctx); err != nil {
			return fmt.Errorf("execute template %q: %w", templateFile, err)
		}

		mng.Commit()
	}
	return nil
}

// getExtraSourceDir tries to find extra code source directory for the given protocol. If found, returns its path and true.
// otherwise, it returns the path for unknown protocol and false.
func getExtraSourceDir(filesystem fs.FS, protocol string) (string, bool) {
	d := path.Join(protocol, extraCodeSrcDir)
	if f, err := filesystem.Open(d); err == nil {
		defer f.Close()
		return d, true
	}

	return path.Join(unknownProtocolSrcDir, extraCodeSrcDir), false
}

// LoadImplementationsManifests loads the built-in implementations manifests file.
func LoadImplementationsManifests(tplFS fs.FS) (codeextra.ImplementationManifests, error) {
	f, err := tplFS.Open("manifests.yaml")
	if err != nil {
		return nil, fmt.Errorf("cannot open manifests.yaml: %w", err)
	}
	defer f.Close()

	dec := yaml.NewDecoder(f)
	var meta codeextra.ImplementationManifests
	if err = dec.Decode(&meta); err != nil {
		return nil, fmt.Errorf("cannot parse manifests.yaml: %w", err)
	}

	return meta, nil
}
