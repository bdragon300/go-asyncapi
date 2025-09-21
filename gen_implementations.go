package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"slices"
	"strings"

	"github.com/bdragon300/go-asyncapi/implementations"
	"github.com/bdragon300/go-asyncapi/internal/common"
	"github.com/bdragon300/go-asyncapi/internal/log"
	"github.com/bdragon300/go-asyncapi/internal/tmpl"
	"github.com/bdragon300/go-asyncapi/internal/tmpl/manager"
	"github.com/bdragon300/go-asyncapi/internal/utils"
	"github.com/samber/lo"
	"gopkg.in/yaml.v3"
)

const (
	toolchainCommand = "go"
	runModuleDir     = "run"
	goModuleName     = "implementations"
)

func main() {
	outDir := flag.String("out", "implementations_go", "output implementations directory")
	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), "Dev tool that generates the implementations code from templates and save it to the output directory.\n\n")
		fmt.Fprintf(flag.CommandLine.Output(), "Usage of %s:\n", os.Args[0])
		flag.PrintDefaults()
	}
	flag.Parse()

	manifest, err := loadImplementationsManifest()
	if err != nil {
		panic(err)
	}

	renderManager := manager.NewTemplateRenderManager(common.RenderOpts{PreambleTemplate: "preamble.tmpl"})
	tplLoader := tmpl.NewTemplateLoader("", implementations.ImplementationFS)
	renderManager.TemplateLoader = tplLoader

	if err := render(manifest, tplLoader, renderManager); err != nil {
		panic(fmt.Sprintf("render implementations: %v", err))
	}
	if err := writeFiles(renderManager, *outDir); err != nil {
		panic(fmt.Sprintf("write files: %v", err))
	}
	if err := prepareGoModule(*outDir); err != nil {
		panic(fmt.Sprintf("prepare Go module: %v", err))
	}
	fmt.Println("All done")
}

func loadImplementationsManifest() (implementations.ImplManifest, error) {
	f, err := implementations.ImplementationFS.Open("manifest.yaml")
	if err != nil {
		return nil, fmt.Errorf("cannot open manifest.yaml: %w", err)
	}
	dec := yaml.NewDecoder(f)
	var meta implementations.ImplManifest
	if err = dec.Decode(&meta); err != nil {
		return nil, fmt.Errorf("cannot parse manifest.yaml: %w", err)
	}

	return meta, nil
}

func render(manifest implementations.ImplManifest, tplLoader *tmpl.TemplateLoader, renderManager *manager.TemplateRenderManager) error {
	implGroups := lo.GroupBy(manifest, func(item implementations.ImplManifestItem) string {
		return item.Protocol
	})
	protos := lo.Keys(implGroups)
	slices.Sort(protos)
	for _, proto := range protos {
		for _, implementation := range implGroups[proto] {
			fmt.Printf("Processing %s (%s)\n", implementation.Name, implementation.Protocol)
			ctx := tmpl.ImplTemplateContext{
				Manifest:  implementation,
				Directory: implementation.Dir,
				Package:   utils.ToGolangName(utils.GetPackageName(implementation.Dir), false),
			}
			fileNames, err := tplLoader.ParseDir(implementation.Dir, renderManager)
			if err != nil {
				return fmt.Errorf("parse templates %q: %w", implementation.Dir, err)
			}
			for _, fileName := range fileNames {
				normFileName := utils.ToGoFilePath(path.Join(implementation.Protocol, fileName))
				renderManager.BeginFile(normFileName, ctx.Package)

				// Render package header
				fmt.Fprintf(renderManager.Buffer, "package %s\n\n", ctx.Package)

				// Render the rest
				fmt.Printf("-> Render file %s\n", normFileName)
				tpl, err := tplLoader.LoadTemplate(fileName)
				if err != nil {
					return fmt.Errorf("load template %q: %w", fileName, err)
				}
				if err := tpl.Execute(renderManager.Buffer, ctx); err != nil {
					return fmt.Errorf("execute template %q: %w", fileName, err)
				}

				renderManager.Commit()
			}
		}
	}

	return nil
}

func writeFiles(renderManager *manager.TemplateRenderManager, outDir string) error {
	for fileName, state := range renderManager.CommittedStates() {
		fullFileName := path.Join(outDir, fileName)
		fmt.Printf("Writing file %s\n", fullFileName)
		if err := ensureDir(path.Dir(fullFileName)); err != nil {
			return fmt.Errorf("create directory: %w", err)
		}
		if err := os.WriteFile(fullFileName, state.Buffer.Bytes(), 0o644); err != nil {
			return fmt.Errorf("write: %w", err)
		}
	}
	return nil
}

// ensureDir ensures that the directory at the given path exists. If not, creates it recursively.
func ensureDir(path string) error {
	if info, err := os.Stat(path); os.IsNotExist(err) {
		if err2 := os.MkdirAll(path, 0o755); err2 != nil {
			return err2
		}
	} else if err != nil {
		return err
	} else if !info.IsDir() {
		return fmt.Errorf("path %q is not a directory", path)
	}

	return nil
}

func prepareGoModule(outDir string) error {
	logger := log.GetLogger("")
	toolchainPath, err := exec.LookPath(toolchainCommand)
	if err != nil {
		return fmt.Errorf("find toolchain %q: %w", toolchainCommand, err)
	}
	logger.Debug("Go toolchain found", "path", toolchainPath)

	workingDir := path.Join(".", outDir)
	repoDir, err := filepath.Rel(path.Join(".", outDir), ".")
	if err != nil {
		return fmt.Errorf("relative path: %w", err)
	}
	fmt.Println("Repo dir:", repoDir, "Working dir:", workingDir)

	// Go subcommands to run
	subcommands := [][]string{
		{"mod", "init", goModuleName},
		{"work", "init"},
		{"work", "use", "."},
		{"work", "use", path.Clean(path.Join(repoDir, runModuleDir))},
		{"get", "-v", "./..."},
	}

	for _, subcommand := range subcommands {
		cmdLine := toolchainPath + " " + strings.Join(subcommand, " ")
		fmt.Println("Run command: ", cmdLine)
		cmd := exec.Command(toolchainPath, subcommand...)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Stdin = os.Stdin
		cmd.Dir = workingDir
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("command %q: %w", cmdLine, err)
		}
	}

	return nil
}
