package renderer

import (
	"fmt"
	"io/fs"
	"os"
	"path"

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
	utilCodeSrcDir        = "util"
	unknownProtocolSrcDir = "unknown"
)

func RenderUtilCode(protocols []string, opts common.RenderOpts, mng *manager.TemplateRenderManager, tplBase fs.FS) error {
	logger := log.GetLogger(log.LoggerPrefixRendering)
	prevTplLoader := mng.TemplateLoader
	defer func() {
		mng.TemplateLoader = prevTplLoader
	}()

	for _, protocol := range protocols {
		logger.Debug("Render util code", "protocol", protocol)
		ctx := tmpl.CodeExtraTemplateContext{RenderOpts: opts, Protocol: protocol}
		directory, err := renderInlineTemplate(opts.UtilCodeOpts.Directory, ctx, mng)
		if err != nil {
			return fmt.Errorf("render directory expression: %w", err)
		}
		logger.Trace("-> Directory", "result", directory)

		pkgName := utils.GetPackageName(directory)
		ctx = tmpl.CodeExtraTemplateContext{RenderOpts: opts, Protocol: protocol, Directory: directory, PackageName: pkgName}
		logger.Trace("-> Package name", "name", pkgName)

		srcDir, found := getUtilSourceDir(tplBase, protocol)
		if !found {
			logger.Warn("-> Unknown protocol found, generate the fallback code", "protocol", protocol)
		}
		logger.Trace("-> Util code source directory", "srcDir", srcDir)

		userConfig, _ := lo.Find(opts.UtilCodeOpts.Custom, func(item common.UtilCodeCustomOpts) bool {
			return item.Protocol == protocol
		})

		var templates []string
		if userConfig.TemplateDirectory != "" {
			logger.Trace("-> Using the custom template directory", "directory", userConfig.TemplateDirectory)
			ld := tmpl.NewTemplateLoader("", os.DirFS(userConfig.TemplateDirectory))
			mng.TemplateLoader = ld

			templates, err = ld.ParseDir(".", mng)
			if err != nil {
				return fmt.Errorf("parse templates from directory %q: %w", userConfig.TemplateDirectory, err)
			}
		} else {
			logger.Trace("-> Using the built-in util code templates")
			ld := tmpl.NewTemplateLoader("", tplBase)
			mng.TemplateLoader = ld

			templates, err = ld.ParseDir(srcDir, mng)
			if err != nil {
				return fmt.Errorf("parse templates from built-in util code for protocol %q: %w", protocol, err)
			}
		}

		logger.Trace("-> Templates found", "files", templates)
		if err = renderCodeExtraTemplates(templates, ctx, mng, nil); err != nil {
			return fmt.Errorf("render templates for protocol %q: %w", protocol, err)
		}
	}
	return nil
}

func RenderImplementationCode(protocols []string, opts common.RenderOpts, mng *manager.TemplateRenderManager, tplBase fs.FS) error {
	logger := log.GetLogger(log.LoggerPrefixRendering)
	prevTplLoader := mng.TemplateLoader
	defer func() {
		mng.TemplateLoader = prevTplLoader
	}()
	manifests := lo.Must(LoadImplementationsManifests(tplBase))

	for _, protocol := range protocols {
		logger.Debug("Render implementation code", "protocol", protocol)

		userConfig, _ := lo.Find(opts.ImplementationCodeOpts.Custom, func(item common.ImplementationCodeCustomOpts) bool {
			return item.Protocol == protocol
		})
		if userConfig.Disable {
			logger.Debug("-> Implementation is disabled for protocol, skipping")
			continue
		}

		ctx := tmpl.CodeExtraTemplateContext{RenderOpts: opts, Protocol: protocol}
		directory, err := renderInlineTemplate(opts.ImplementationCodeOpts.Directory, ctx, mng)
		if err != nil {
			return fmt.Errorf("render directory expression: %w", err)
		}
		logger.Trace("-> Directory", "result", directory)

		pkgName, _ := lo.Coalesce(userConfig.Package, utils.GetPackageName(directory))
		ctx = tmpl.CodeExtraTemplateContext{RenderOpts: opts, Protocol: protocol, Directory: directory, PackageName: pkgName}
		logger.Trace("-> Package name", "name", pkgName)

		var templates []string
		if userConfig.TemplateDirectory != "" {
			logger.Trace("-> Using the custom template directory", "directory", userConfig.TemplateDirectory)
			ld := tmpl.NewTemplateLoader("", os.DirFS(userConfig.TemplateDirectory))
			mng.TemplateLoader = ld

			templates, err = ld.ParseDir(".", mng)
			if err != nil {
				return fmt.Errorf("parse templates from directory %q: %w", userConfig.TemplateDirectory, err)
			}
		} else {
			logger.Trace("-> Using the built-in implementations", "name", lo.CoalesceOrEmpty(userConfig.Name, "<default>"))
			ld := tmpl.NewTemplateLoader("", tplBase)
			mng.TemplateLoader = ld

			man, found := lo.Find(manifests, func(item codeextra.ImplementationManifest) bool {
				return item.Protocol == protocol && (userConfig.Name == "" || item.Name == userConfig.Name)
			})
			if !found {
				logger.Warn("-> No implementation found for protocol, skipping", "protocol", protocol, "name", lo.CoalesceOrEmpty(userConfig.Name, "<default>"))
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
		if err = renderCodeExtraTemplates(templates, ctx, mng, &userConfig); err != nil {
			return fmt.Errorf("render templates for protocol %q: %w", protocol, err)
		}
	}
	return nil
}

func renderCodeExtraTemplates(templates []string, ctx tmpl.CodeExtraTemplateContext, mng *manager.TemplateRenderManager, implConf *common.ImplementationCodeCustomOpts) error {
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

// getUtilSourceDir returns the directory with util code sources for the given protocol. If found, returns its path and true,
// otherwise, it returns the default unknown protocol directory and false.
func getUtilSourceDir(filesystem fs.FS, protocol string) (string, bool) {
	d := path.Join(protocol, utilCodeSrcDir)
	if f, err := filesystem.Open(d); err == nil {
		defer f.Close()
		return d, true
	}

	return path.Join(unknownProtocolSrcDir, utilCodeSrcDir), false
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
