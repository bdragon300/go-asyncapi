// Package renderer contains the rendering stage logic.
//
// Function in this package (renderers) take the artifacts queue and implementations, pass them to the
// templates and produce the final output contents.
//
// Renderer functions rely on the manager.TemplateRenderManager. This
// is an assistant object that keeps the files state consistent (including the imports, namespace, etc.) and accumulates
// their rendered content. See TemplateRenderManager docs for more information.
//
// Roughly, the rendering process looks like this:
//
//  1. Before the rendering, the caller builds the objects queue by applying the config "selections" to the
//     artifacts in compiled documents. Also, it loads the templates.
//  2. The caller calls the renderer, passing the queue, implementations, and manager.TemplateRenderManager and
//     the options.
//  3. The renderer function iterates over the queue, invokes the templates, and commits the results to the rendering manager.
//  4. In the same way, the caller may additionally run other renderers, for example, to attach some output to the result.
//  3. Calling the FinishFiles extracts the output from the render manager, postprocesses it, and returns the file's content.
//
// Thus, after the FinishFiles is called, the renderer stage is considered finished.
package renderer

import (
	"bytes"
	"fmt"
	"slices"
	"strings"
	"text/template"
	"unicode"

	"github.com/bdragon300/go-asyncapi/internal/common"
	"github.com/bdragon300/go-asyncapi/internal/log"
	"github.com/bdragon300/go-asyncapi/internal/tmpl"
	"github.com/bdragon300/go-asyncapi/internal/tmpl/manager"
	"github.com/samber/lo"
)

// FinishFiles extracts file's content from TemplateRenderManager, filters out the empty files, renders the preamble
// before each *.go file. Function returns the finished content buffers for all files.
func FinishFiles(mng *manager.TemplateRenderManager) (map[string]*bytes.Buffer, error) {
	states := mng.CommittedStates()

	res := make(map[string]*bytes.Buffer, len(states))
	logger := log.GetLogger(log.LoggerPrefixRendering)
	logger.Debug("Finish files rendering", "files", len(states))

	tpl, err := mng.TemplateLoader.LoadTemplate(mng.RenderOpts.PreambleTemplate)
	if err != nil {
		return nil, fmt.Errorf("template %q: %w", mng.RenderOpts.PreambleTemplate, err)
	}

	keys := lo.Keys(states)
	slices.Sort(keys)
	for _, fileName := range keys {
		state := states[fileName]
		logger.Debug("Render file", "file", fileName, "package", state.PackageName, "imports", state.Imports.String())
		// Skip empty buffers. Spaces and tabs are not considered as a content.
		if !bytes.ContainsFunc(state.Buffer.Bytes(), unicode.IsLetter) {
			logger.Debug("-> Skip empty file", "file", fileName)
			continue
		}

		b := new(bytes.Buffer)
		// Prepend the content with preamble if it's a Go file.
		if strings.HasSuffix(fileName, ".go") {
			logger.Debug("-> Render preamble", "file", fileName)
			b, err = renderPreambleTemplate(tpl, mng.RenderOpts, state)
			if err != nil {
				return nil, err
			}
		}
		if _, err := b.ReadFrom(state.Buffer); err != nil {
			return nil, err
		}
		res[fileName] = b
	}

	return res, nil
}

// renderObjectInlineTemplate renders a small template snippet, passing the object, CodeTemplateContext and all
// standard template functions. The result puts to the given template manager.
func renderObjectInlineTemplate(item RenderQueueItem, text string, mng *manager.TemplateRenderManager) (string, error) {
	tplCtx := &tmpl.CodeTemplateContext{
		RenderOpts:       mng.RenderOpts,
		CurrentSelection: item.Selection,
		PackageName:      item.Selection.Render.Package,
		Object:           item.Object.Renderable,
		ImportsManager:   mng.ImportsManager,
	}

	return renderInlineTemplate(text, tplCtx, mng)
}

// renderInlineTemplate renders a small template snippet, with any context and with standard template functions.
// The result puts to the given template manager.
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

func renderPreambleTemplate(tpl *template.Template, opts common.RenderOpts, state manager.FileRenderState) (*bytes.Buffer, error) {
	var res bytes.Buffer

	tplCtx := &tmpl.CodeTemplateContext{
		RenderOpts:     opts,
		PackageName:    state.PackageName,
		ImportsManager: state.Imports,
	}

	if err := tpl.Execute(&res, tplCtx); err != nil {
		return nil, err
	}
	res.WriteRune('\n')

	return &res, nil
}
