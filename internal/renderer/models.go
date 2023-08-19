package renderer

import (
	"bytes"
	"embed"
	"errors"
	"html/template"
	"path"
	"strconv"
	"strings"

	"github.com/bdragon300/asyncapi-codegen/internal/assets/types"
	"github.com/bdragon300/asyncapi-codegen/internal/scanner"
)

type ModelsTplArgs struct {
	Definitions string
}

func RenderTypes(bucket *types.LangTypeBucket, tplDir embed.FS) (files map[string]bytes.Buffer, err error) {
	var defBuilder strings.Builder

	names := make(map[string]scanner.LangRenderer)
	var itemsToRender []scanner.LangRenderer
	for _, item := range bucket.Items() {
		if item.SkipRender() {
			continue
		}
		name := getUniqueName(item, names)
		item.PrepareRender(name)
		itemsToRender = append(itemsToRender, item)
	}

	for _, item := range itemsToRender {
		rendered := item.RenderDefinition()
		for _, stmt := range rendered {
			if e := stmt.Render(&defBuilder); e != nil {
				err = errors.Join(err, e)
			}
			defBuilder.WriteRune('\n')
		}
	}
	if err != nil {
		return
	}

	tplBuilder, err := renderTemplate(tplDir, "models/models.gotmpl", ModelsTplArgs{Definitions: defBuilder.String()})
	if err != nil {
		return
	}

	files = map[string]bytes.Buffer{
		"models/models.go": tplBuilder,
	}

	return
}

func renderTemplate(tplFS embed.FS, tplPath string, tplArgs any) (bytes.Buffer, error) {
	var buf bytes.Buffer
	tpl, err := template.New(path.Base(tplPath)).ParseFS(tplFS, path.Join(path.Dir(tplPath), "*.gotmpl"))
	if err != nil {
		return buf, err
	}
	if err = tpl.Execute(&buf, tplArgs); err != nil {
		return buf, err
	}
	return buf, nil
}

func getUniqueName(typ scanner.LangRenderer, names map[string]scanner.LangRenderer) string {
	langName := typ.GetDefaultName()
	findName := langName

	// Use type's name or append a number such as MyType2, MyType3, ...
	for i := 1; ; i++ {
		if _, ok := names[findName]; !ok {
			names[findName] = typ
			return findName
		}
		findName = langName + strconv.Itoa(i)
	}
}
