package main

import (
	"bytes"
	"flag"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"slices"
	"strings"
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
	toolchainCommand  = "go"
	runModuleDir      = "run"
	goModuleName      = "codeextra"
	templateExtension = ".tmpl"
)

const preambleTemplate = `package {{ .PackageName }}

{{- with .ImportsManager.Imports }}
import (
{{- range . }}
    {{ if .Alias }}{{ .Alias }} {{ end }}"{{ .PackagePath }}"
{{- end }}
)
{{- end }}
// ------------------------------ Inflated template contents starts here ------------------------------
`

func main() {
	outDir := flag.String("out", "codeextra_go", "output directory")
	noGoWork := flag.Bool("no-go-work", false, "do not generate the go.work file")
	autoRemove := flag.Bool("auto-remove", false, "automatically remove existing output directory before generation")
	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), "Dev tool that inflates the codeextra templates into Go code.\n\n")
		fmt.Fprintf(flag.CommandLine.Output(), "Usage of %s:\n", os.Args[0])
		flag.PrintDefaults()
	}
	flag.Parse()

	if *autoRemove {
		if err := os.RemoveAll(*outDir); err != nil {
			panic(fmt.Sprintf("remove existing output directory %q: %v", *outDir, err))
		}
	}

	renderOpts := common.RenderOpts{
		RuntimeModule:          "github.com/bdragon300/go-asyncapi/run",
		ImportBase:             goModuleName,
		UtilCodeOpts:           common.UtilCodeOpts{Directory: "."},
		ImplementationCodeOpts: common.ImplementationCodeOpts{Directory: "."},
	}
	renderManager := manager.NewTemplateRenderManager(renderOpts)

	dirs, err := listAllTemplateDirs(codeextra.TemplateFS)
	if err != nil {
		panic(fmt.Sprintf("list template dirs: %v", err))
	}
	if err := render(dirs, renderManager); err != nil {
		panic(fmt.Sprintf("render implementations: %v", err))
	}
	if err := writeFiles(renderManager, *outDir); err != nil {
		panic(fmt.Sprintf("write files: %v", err))
	}
	if err := prepareGoModule(*outDir, !*noGoWork); err != nil {
		panic(fmt.Sprintf("prepare Go module: %v", err))
	}
	fmt.Println("All done")
}

func listAllTemplateDirs(filesystem fs.ReadDirFS) ([]string, error) {
	var dirs []string
	err := fs.WalkDir(filesystem, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() && path != "." {
			dirs = append(dirs, path)
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("walk dir: %w", err)
	}
	// Filter only dirs that contain .tmpl files
	dirs = lo.Filter(dirs, func(dir string, _ int) bool {
		matches, err := fs.Glob(filesystem, path.Join(dir, "*"+templateExtension))
		if err != nil {
			return false
		}
		return len(matches) > 0
	})

	// Sort dirs to have utils first and then the rest. This is to ensure that util code is generated before
	// implementation code that may depend on it.
	slices.SortFunc(dirs, func(a, b string) int {
		am, err := path.Match("*/util", a)
		if err != nil {
			panic(err)
		}
		bm, err := path.Match("*/util", b)
		if err != nil {
			panic(err)
		}
		if am && !bm {
			return -1
		}
		if !am && bm {
			return 1
		}
		return strings.Compare(a, b)
	})
	return dirs, nil
}

func render(dirs []string, renderManager *manager.TemplateRenderManager) error {
	implementationManifests, err := loadImplementationsManifest()
	if err != nil {
		return fmt.Errorf("load implementations manifest: %w", err)
	}

	findImpl := func(dir string) *codeextra.ImplementationManifest {
		m, ok := lo.Find(implementationManifests, func(m codeextra.ImplementationManifest) bool {
			return m.Dir == dir
		})
		if ok {
			return &m
		}
		return nil
	}

	for _, dir := range dirs {
		tplLoader := tmpl.NewTemplateLoader("", codeextra.TemplateFS)
		renderManager.TemplateLoader = tplLoader
		templates, err := tplLoader.ParseDir(dir, renderManager)
		if err != nil {
			return fmt.Errorf("parse templates in dir %q: %w", dir, err)
		}

		for _, templateName := range templates {
			ctx := tmpl.CodeExtraTemplateContext{
				RenderOpts:  renderManager.RenderOpts,
				PackageName: utils.ToGolangName(path.Base(dir), false),
			}
			man := findImpl(dir)
			if man != nil {
				ctx.Protocol = man.Protocol
				ctx.Manifest = man
			} else {
				ctx.Protocol = strings.Split(dir, string(os.PathSeparator))[0]
			}
			err2 := renderTemplate(ctx, path.Join(dir, templateName), tplLoader, renderManager)
			if err2 != nil {
				return fmt.Errorf("render template %q: %w", templateName, err2)
			}
		}
	}

	return nil
}

func loadImplementationsManifest() (codeextra.ImplementationManifests, error) {
	f, err := codeextra.TemplateFS.Open("manifests.yaml")
	if err != nil {
		return nil, fmt.Errorf("cannot open manifests.yaml: %w", err)
	}
	dec := yaml.NewDecoder(f)
	var meta codeextra.ImplementationManifests
	if err = dec.Decode(&meta); err != nil {
		return nil, fmt.Errorf("cannot parse manifests.yaml: %w", err)
	}

	return meta, nil
}

func renderTemplate(ctx tmpl.CodeExtraTemplateContext, tplPath string, tplLoader *tmpl.TemplateLoader, renderManager *manager.TemplateRenderManager) error {
	fmt.Printf("Processing %s\n", tplPath)

	normFileName := utils.ToGoFilePath(tplPath)
	renderManager.BeginFile(normFileName, ctx.PackageName)
	renderManager.ExtraCodeProtocol = ctx.Protocol
	renderManager.ImplementationManifest = ctx.Manifest
	if ctx.Manifest != nil {
		renderManager.ImplementationConfig = &common.ImplementationCodeCustomOpts{
			Protocol: ctx.Protocol,
			Name:     ctx.Manifest.Name,
			Package:  ctx.PackageName,
		}
	}

	// Render the rest
	fmt.Printf("-> Render file %s\n", normFileName)
	tpl, err := tplLoader.LoadTemplate(path.Base(tplPath))
	if err != nil {
		return fmt.Errorf("load: %w", err)
	}
	if err := tpl.Execute(renderManager.Buffer, ctx); err != nil {
		return fmt.Errorf("execute: %w", err)
	}

	renderManager.Commit()
	return nil
}

func writeFiles(renderManager *manager.TemplateRenderManager, outDir string) error {
	for fileName, state := range renderManager.CommittedStates() {
		fullFileName := path.Join(outDir, fileName)
		fmt.Printf("Writing file %s\n", fullFileName)
		if err := ensureDir(path.Dir(fullFileName)); err != nil {
			return fmt.Errorf("create directory: %w", err)
		}

		tplCtx := tmpl.CodeTemplateContext{
			RenderOpts:     renderManager.RenderOpts,
			PackageName:    state.PackageName,
			ImportsManager: state.Imports,
		}
		preambleStr, err := renderInlineTemplate(preambleTemplate, tplCtx, renderManager)
		if err != nil {
			return fmt.Errorf("render preamble: %w", err)
		}

		var finalBuffer bytes.Buffer
		finalBuffer.WriteString(preambleStr)
		finalBuffer.Write(state.Buffer.Bytes())

		if err := os.WriteFile(fullFileName, finalBuffer.Bytes(), 0o644); err != nil {
			return fmt.Errorf("write: %w", err)
		}
	}
	return nil
}

func renderInlineTemplate(text string, tplCtx any, renderManager *manager.TemplateRenderManager) (string, error) {
	var res bytes.Buffer
	tpl, err := template.New("").Funcs(tmpl.GetTemplateFunctions(renderManager)).Parse(text)
	if err != nil {
		return "", err
	}
	if err = tpl.Execute(&res, tplCtx); err != nil {
		return "", err
	}
	return res.String(), nil
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

func prepareGoModule(outDir string, generateGoWork bool) error {
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
	devRunPath := path.Clean(path.Join(repoDir, runModuleDir))
	subcommands := [][]string{
		{"mod", "init", goModuleName},
		// Add replace directive to use local run module
		{"mod", "edit", "-replace", fmt.Sprintf("github.com/bdragon300/go-asyncapi/run=%s", devRunPath)},
	}
	if generateGoWork {
		subcommands = append(subcommands, [][]string{
			{"work", "init"},
			{"work", "use", "."},
			{"work", "use", devRunPath},
		}...)
	}
	subcommands = append(subcommands, []string{"get", "-v", "./..."})

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
