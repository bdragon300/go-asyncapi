package main

import (
	"fmt"
	"os"
)

func listImplementations() {
	manifest, err := getImplementationsManifest()
	if err != nil {
		panic(err.Error())
	}
	for proto, implInfo := range manifest {
		_, _ = os.Stdout.WriteString(proto + ":\n")
		for implName, info := range implInfo {
			_, _ = os.Stdout.WriteString(fmt.Sprintf("* %s (%s)\n", implName, info.URL))
		}
		_, _ = os.Stdout.WriteString("\n")
	}
}
