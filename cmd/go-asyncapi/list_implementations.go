package main

import (
	"fmt"
	"os"
	"slices"

	"github.com/bdragon300/go-asyncapi/internal/renderer"
	"github.com/bdragon300/go-asyncapi/templates/codeextra"

	"github.com/samber/lo"
)

func listImplementations() {
	manifest := lo.Must(renderer.LoadImplementationsManifests(codeextra.TemplateFS))
	implGroups := lo.GroupBy(manifest, func(item codeextra.ImplementationManifest) string {
		return item.Protocol
	})
	protos := lo.Keys(implGroups)
	slices.Sort(protos)
	for _, proto := range protos {
		_, _ = os.Stdout.WriteString(proto + ":\n")
		for _, info := range implGroups[proto] {
			_, _ = fmt.Fprintf(os.Stdout,
				"* %s (%s) %s\n", info.Name, info.URL, lo.Ternary(info.Default, "[default]", ""))
		}
		_, _ = os.Stdout.WriteString("\n")
	}
}
