package renderer

import (
	"bytes"
	"fmt"
	"github.com/bdragon300/go-asyncapi/internal/common"
	"github.com/bdragon300/go-asyncapi/internal/log"
	"github.com/bdragon300/go-asyncapi/internal/tmpl"
	"github.com/bdragon300/go-asyncapi/internal/tmpl/manager"
	"github.com/samber/lo"
	"slices"
	"text/template"
	"unicode"
)

func FinishFiles(mng *manager.TemplateRenderManager) (map[string]*bytes.Buffer, error) {
	states := mng.CommittedStates()

	var res = make(map[string]*bytes.Buffer, len(states))
	logger := log.GetLogger(log.LoggerPrefixRendering)
	logger.Debug("Render files", "files", len(states))

	tpl, err := mng.TemplateLoader.LoadTemplate(mng.RenderOpts.PreambleTemplate)
	if err != nil {
		return nil, fmt.Errorf("template %q: %w", mng.RenderOpts.PreambleTemplate, err)
	}

	keys := lo.Keys(states)
	slices.Sort(keys)
	for _, fileName := range keys {
		state := states[fileName]
		logger.Debug("Render file", "file", fileName, "package", state.PackageName, "imports", state.Imports.String())
		if !bytes.ContainsFunc(state.Buffer.Bytes(), unicode.IsLetter) {
			logger.Debug("-> Skip empty file", "file", fileName)
			continue
		}
		b, err := renderPreambleTemplate(tpl, mng.RenderOpts, state)
		if err != nil {
			return nil, err
		}
		if _, err := b.ReadFrom(state.Buffer); err != nil {
			return nil, err
		}
		res[fileName] = b
	}

	return res, nil
}

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
		RenderOpts:       opts,
		PackageName:      state.PackageName,
		ImportsManager:   state.Imports,
	}

	if err := tpl.Execute(&res, tplCtx); err != nil {
		return nil, err
	}
	res.WriteRune('\n')

	return &res, nil
}
