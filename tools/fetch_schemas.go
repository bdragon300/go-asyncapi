//go:build ignore

// Command fetch_schemas downloads the AsyncAPI JSON Schema files used by
// the "go-asyncapi doc validate" subcommand from the official spec repository
// (https://github.com/asyncapi/spec-json-schemas) and writes them into the
// assets directory so they can be embedded into the binary.
//
// It is meant to be run by a developer via "go generate" (see assets/schema.go)
// or "task fetch-schemas"; it is not part of the regular build (see the
// //go:build ignore tag above).
package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// schemaURLTemplate points to the bundled, self-contained AsyncAPI schema for a
// given version in the spec-json-schemas repository. The bundled file inlines
// all of its $ref'd definitions, so it can be used for validation offline.
const schemaURLTemplate = "https://raw.githubusercontent.com/asyncapi/spec-json-schemas/master/schemas/%s.json"

func main() {
	output := flag.String("output", "schemas", "directory to write the schema files into")
	versionsFlag := flag.String("versions", "", "comma-separated list of AsyncAPI versions to fetch schemas for")
	flag.Parse()

	if *versionsFlag == "" {
		log.Fatal("no versions specified, use -versions flag")
	}

	if err := os.MkdirAll(*output, 0o755); err != nil {
		log.Fatalf("create output directory %q: %v", *output, err)
	}
	client := &http.Client{Timeout: 60 * time.Second}

	versions := strings.Split(*versionsFlag, ",")
	for _, version := range versions {
		version = strings.TrimSpace(version)
		url := fmt.Sprintf(schemaURLTemplate, version)
		dst := filepath.Join(*output, version+".json")
		log.Printf("fetching AsyncAPI %s schema from %s", version, url)
		if err := download(client, url, dst); err != nil {
			log.Fatalf("fetch AsyncAPI %s schema: %v", version, err)
		}
		log.Printf("wrote %s", dst)
	}
	log.Printf("done, fetched %d schema(s)", len(versions))
}

func download(client *http.Client, url, dst string) error {
	resp, err := client.Get(url)
	if err != nil {
		return fmt.Errorf("request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status %s", resp.Status)
	}

	f, err := os.Create(dst)
	if err != nil {
		return fmt.Errorf("create %q: %w", dst, err)
	}
	defer f.Close()

	if _, err := io.Copy(f, resp.Body); err != nil {
		return fmt.Errorf("write %q: %w", dst, err)
	}
	return nil
}
