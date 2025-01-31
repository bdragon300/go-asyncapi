package renderer

import (
	"bytes"
	"github.com/bdragon300/go-asyncapi/internal/common"
	"github.com/bdragon300/go-asyncapi/internal/render/context"
	"github.com/bdragon300/go-asyncapi/internal/tmpl"
	"text/template"
)

func renderObjectInlineTemplate(item renderQueueItem, opts common.RenderOpts, text string) (string, error) {
	ctx := &context.RenderContextImpl{
		RenderOpts:             opts,
		CurrentSelectionConfig: item.selection,
		PackageName:            item.selection.Render.Package,  // May be empty string here
		Imports:                &context.ImportsList{},
		Object:                 item.object,
	}
	tmpl.SetContext(ctx)

	tplCtx := tmpl.NewTemplateContext(ctx, item.object.Renderable, ctx.Imports)

	return renderInlineTemplate(text, tplCtx)
}

func renderInlineTemplate(text string, tplCtx any) (string, error) {
	var res bytes.Buffer
	tpl, err := template.New("").Funcs(tmpl.GetTemplateFunctions()).Parse(text)
	if err != nil {
		return "", err
	}
	if err = tpl.Execute(&res, tplCtx); err != nil {
		return "", err
	}
	return res.String(), nil
}

func renderObjectFileTemplate(preambleTpl *template.Template, opts common.RenderOpts, renderState fileRenderState) (*bytes.Buffer, error) {
	var res bytes.Buffer

	ctx := &context.RenderContextImpl{
		RenderOpts:  opts,
		PackageName: renderState.packageName,
		Imports:     &renderState.imports,
	}
	tmpl.SetContext(ctx)
	tplCtx := tmpl.NewTemplateContext(ctx, nil, &renderState.imports)

	if err := preambleTpl.Execute(&res, tplCtx); err != nil {
		return nil, err
	}
	res.WriteRune('\n')
	if _, err := res.Write(renderState.buf.Bytes()); err != nil {
		return nil, err
	}

	return &res, nil
}
