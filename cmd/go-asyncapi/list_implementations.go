package main

import (
	"fmt"
	"github.com/bdragon300/go-asyncapi/implementations"
	"os"

	"github.com/samber/lo"
	"golang.org/x/exp/slices"
)

func listImplementations() {
	manifest := lo.Must(loadImplementationsManifest())
	implGroups := lo.GroupBy(manifest, func(item implementations.ImplManifestItem) string {
		return item.Protocol
	})
	protos := lo.Keys(implGroups)
	slices.Sort(protos)
	for _, proto := range protos {
		_, _ = os.Stdout.WriteString(proto + ":\n")
		for _, info := range implGroups[proto] {
			_, _ = os.Stdout.WriteString(
				fmt.Sprintf("* %s (%s) %s\n", info.Name, info.URL, lo.Ternary(info.Default, "[default]", "")),
			)
		}
		_, _ = os.Stdout.WriteString("\n")
	}
}
