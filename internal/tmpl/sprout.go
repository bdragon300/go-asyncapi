package tmpl

import (
	"github.com/go-sprout/sprout"
	"github.com/go-sprout/sprout/registry/checksum"
	"github.com/go-sprout/sprout/registry/conversion"
	"github.com/go-sprout/sprout/registry/encoding"
	"github.com/go-sprout/sprout/registry/env"
	"github.com/go-sprout/sprout/registry/filesystem"
	"github.com/go-sprout/sprout/registry/maps"
	"github.com/go-sprout/sprout/registry/numeric"
	"github.com/go-sprout/sprout/registry/random"
	"github.com/go-sprout/sprout/registry/reflect"
	"github.com/go-sprout/sprout/registry/regexp"
	"github.com/go-sprout/sprout/registry/semver"
	"github.com/go-sprout/sprout/registry/slices"
	"github.com/go-sprout/sprout/registry/std"
	"github.com/go-sprout/sprout/registry/strings"
	"github.com/go-sprout/sprout/registry/time"
	"github.com/go-sprout/sprout/registry/uniqueid"
	"github.com/samber/lo"
)

var sproutFunctions sprout.FunctionMap

func init() {
	// https://docs.atom.codes/sprout/registries/list-of-all-registries
	handler := sprout.New()
	lo.Must0(handler.AddRegistries(
		checksum.NewRegistry(),
		conversion.NewRegistry(),
		encoding.NewRegistry(),
		env.NewRegistry(),
		filesystem.NewRegistry(),
		maps.NewRegistry(),
		numeric.NewRegistry(),
		random.NewRegistry(),
		reflect.NewRegistry(),
		regexp.NewRegistry(),
		semver.NewRegistry(),
		slices.NewRegistry(),
		std.NewRegistry(),
		strings.NewRegistry(),
		time.NewRegistry(),
		uniqueid.NewRegistry(),
	))
	sproutFunctions = handler.Build()
}
