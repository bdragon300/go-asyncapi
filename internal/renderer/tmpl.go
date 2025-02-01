package renderer

import (
	"bytes"
	"github.com/bdragon300/go-asyncapi/internal/tmpl"
	"github.com/bdragon300/go-asyncapi/internal/tmpl/manager"
	"text/template"
)

func renderObjectInlineTemplate(item renderQueueItem, text string, mng *manager.TemplateRenderManager) (string, error) {
	tplCtx := &tmpl.CodeTemplateContext{
		RenderOpts:       mng.RenderOpts,
		CurrentSelection: item.selection,
		PackageName:      item.selection.Render.Package,
		Object:           item.object.Renderable,
		ImportsManager:   &mng.ImportsManager,
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

func renderPreambleTemplate(tpl *template.Template, mng *manager.TemplateRenderManager) (*bytes.Buffer, error) {
	var res bytes.Buffer

	tplCtx := &tmpl.CodeTemplateContext{
		RenderOpts:       mng.RenderOpts,
		PackageName:      mng.PackageName,
		ImportsManager:   &mng.ImportsManager,
	}

	if err := tpl.Execute(&res, tplCtx); err != nil {
		return nil, err
	}
	res.WriteRune('\n')
	if _, err := res.Write(mng.Buffer.Bytes()); err != nil {
		return nil, err
	}

	return &res, nil
}
