package main

import (
	"fmt"
	"io"
	"os"
	"path"
	"strings"

	"github.com/bdragon300/asyncapi-codegen-go/internal/asyncapi"
	"github.com/bdragon300/asyncapi-codegen-go/internal/common"
	"github.com/bdragon300/asyncapi-codegen-go/internal/compiler"
	"github.com/bdragon300/asyncapi-codegen-go/internal/linker"
	"github.com/bdragon300/asyncapi-codegen-go/internal/utils"
	"github.com/bdragon300/asyncapi-codegen-go/internal/writer"
	"github.com/samber/lo"
	"golang.org/x/mod/modfile"
)

func generate(cmd *GenerateCmd) error {
	importBase := cmd.ProjectModule
	if importBase == "" {
		b, err := getImportBase()
		if err != nil {
			return fmt.Errorf("extraction module name from go.mod (you can specify it by -M argument): %w", err)
		}
		importBase = b
	}
	targetPkg, _ := lo.Coalesce(cmd.TargetPackage, cmd.TargetDir)
	logger.Debugf("Target package name is %s", targetPkg)
	importBase = path.Join(importBase, targetPkg)
	logger.Debugf("Target import base is %s", importBase)
	implDir, _ := lo.Coalesce(cmd.ImplDir, path.Join(cmd.TargetDir, "impl"))
	logger.Debugf("Target implementations directory is %s", implDir)

	// Compilation
	specID, _ := utils.SplitSpecPath(cmd.Spec)
	modules, err := generationCompile(specID)
	if err != nil {
		return err
	}
	objSources := lo.MapValues(modules, func(value *compiler.Module, _ string) linker.ObjectSource { return value })
	mainModule := modules[specID]

	// Linking
	logger.Info("Run linking", "path", specID)
	if err = generationLinking(objSources); err != nil {
		return err
	}

	// Rendering
	logger.Info("Run rendering")
	protoRenderers := lo.MapValues(protocolBuilders(), func(value asyncapi.ProtocolBuilder, _ string) common.ProtocolRenderer { return value })
	files, err := writer.RenderPackages(mainModule, protoRenderers, importBase, cmd.TargetDir)
	if err != nil {
		return fmt.Errorf("schema render: %w", err)
	}

	// Writing
	logger.Info("Run writing")
	if err = writer.WriteToFiles(files, cmd.TargetDir); err != nil {
		return fmt.Errorf("writing code to files: %w", err)
	}

	// Rendering the selected implementations
	logger.Info("Run writing selected implementations")
	selectedImpls := getSelectedImplementations(cmd.ImplementationsOpts)
	if err = generationRenderImplementations(selectedImpls, mainModule, implDir); err != nil {
		return err
	}

	logger.Info("Finished")
	return nil
}

func generationRenderImplementations(selectedImpls map[string]string, mainModule *compiler.Module, implDir string) error {
	implManifest, err := getImplementationsManifest()
	if err != nil {
		panic(err.Error())
	}
	var total int
	for _, p := range mainModule.Protocols() {
		implName := selectedImpls[p]
		if implName == "no" || implName == "" {
			logger.Debug("Implementation has been unselected", "protocol", p)
			continue
		}
		if _, ok := implManifest[p][implName]; !ok {
			return fmt.Errorf("unknown implementation %q for %q protocol, use list-implementations command to see possible values", implName, p)
		}
		logger.Debug("Writing implementation", "protocol", p, "name", implName)
		n, err := writer.WriteImplementation(implManifest[p][implName].Dir, path.Join(implDir, p))
		if err != nil {
			return fmt.Errorf("implementation rendering for protocol %q: %w", p, err)
		}
		total += n
	}
	logger.WithPrefix("Writing 📝").Debugf(
		"Implementations writer stats: total bytes: %d, protocols: %s",
		total, strings.Join(mainModule.Protocols(), ","),
	)
	return nil
}

func generationLinking(objSources map[string]linker.ObjectSource) error {
	// Linking refs
	linker.AssignRefs(objSources)
	danglingRefs := linker.DanglingRefs(objSources)
	logger.Debugf("Linker stats: %s", linker.Stats(objSources))
	if len(danglingRefs) > 0 {
		logger.Error("Some refs remain dangling", "refs", danglingRefs)
		return fmt.Errorf("cannot finish linking")
	}

	// Linking list promises
	logger.Debug("Run linking the list promises")
	linker.AssignListPromises(objSources)
	danglingPromises := linker.DanglingPromisesCount(objSources)
	logger.Debugf("Linker stats: %s", linker.Stats(objSources))
	if danglingPromises > 0 {
		logger.Error("Cannot assign internal list promises", "promises", danglingPromises)
		return fmt.Errorf("cannot finish linking")
	}
	logger.Info("Linking completed", "files", len(objSources))
	return nil
}

func generationCompile(specID string) (map[string]*compiler.Module, error) {
	asyncapi.ProtocolBuilders = protocolBuilders()
	compileQueue := []string{specID}             // Queue of specIDs to compile
	modules := make(map[string]*compiler.Module) // Compilers by spec id
	for len(compileQueue) > 0 {
		specID = compileQueue[0]          // Pop from the queue
		compileQueue = compileQueue[1:]   //
		if _, ok := modules[specID]; ok { // Skip if specID has been already modules
			continue
		}

		logger.Info("Run compilation", "path", specID)
		module := compiler.NewModule(specID)
		modules[specID] = module

		logger.Debug("Loading spec path", "path", specID)
		if err := module.Load(); err != nil {
			return nil, fmt.Errorf("load the spec: %w", err)
		}
		logger.Debug("Compilation a loaded file", "path", specID)
		if err := module.Compile(common.NewCompileContext(specID)); err != nil {
			return nil, fmt.Errorf("compilation the spec: %w", err)
		}
		logger.Debugf("Compiler stats: %s", module.Stats())
		compileQueue = lo.Flatten([][]string{compileQueue, module.RemoteSpecIDs()}) // Extend queue with remote specIDs
	}
	logger.Info("Compilation completed", "files", len(modules))

	return modules, nil
}

func getImportBase() (string, error) {
	pwd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("cannot get current working directory: %w", err)
	}
	fn := path.Join(pwd, "go.mod")
	f, err := os.Open(fn)
	if err != nil {
		return "", fmt.Errorf("unable to open %q: %w", fn, err)
	}
	data, err := io.ReadAll(f)
	if err != nil {
		return "", fmt.Errorf("unable read %q file: %w", fn, err)
	}
	modpath := modfile.ModulePath(data)
	if modpath == "" {
		return "", fmt.Errorf("module path not found in %q", fn)
	}
	return modpath, nil
}
