package main

import (
	"fmt"
	"os"

	"github.com/samber/lo"
	"golang.org/x/exp/slices"
)

func listImplementations() {
	manifest := lo.Must(loadImplementationsManifest())
	protos := lo.Keys(manifest)
	slices.Sort(protos)
	for _, proto := range protos {
		_, _ = os.Stdout.WriteString(proto + ":\n")
		for implName, info := range manifest[proto] {
			_, _ = os.Stdout.WriteString(fmt.Sprintf("* %s (%s)\n", implName, info.URL))
		}
		_, _ = os.Stdout.WriteString("\n")
	}
}
