package common

import (
	"fmt"

	"github.com/bdragon300/go-asyncapi/internal/jsonpointer"
	"github.com/samber/lo"
)

// ArtifactKind is an enumeration of compiled artifact kind.
//
// Basically, this enables us to run different rendering logic (templates subtree) for different artifact kinds, e.g.
// channel or schema model. All artifacts that produce the separate code should have their own ArtifactKind.
//
// Other artifacts that don't produce their own code and affect only the code produced by other artifacts
// (e.g. correlation id or message trait) have ArtifactKindOther kind.
type ArtifactKind string

const (
	ArtifactKindSchema    ArtifactKind = "schema"
	ArtifactKindServer    ArtifactKind = "server"
	ArtifactKindChannel   ArtifactKind = "channel"
	ArtifactKindOperation ArtifactKind = "operation"
	ArtifactKindMessage   ArtifactKind = "message"
	ArtifactKindParameter ArtifactKind = "parameter"
	ArtifactKindSecurity  ArtifactKind = "security"
	// ArtifactKindAsyncAPI represents the root AsyncAPI object.
	ArtifactKindAsyncAPI ArtifactKind = "asyncapi"
	// ArtifactKindOther is a utility language object, not intended to be selected in Selector (e.g. type, value, interface, etc.)
	ArtifactKindOther ArtifactKind = ""
)

// Artifact is a compiled object that is meant to be rendered in the template.
type Artifact interface {
	// Name returns the original name of the object in the document. It can be an entity name, x-go-name field, etc.
	Name() string
	Kind() ArtifactKind
	// Selectable returns true if object is available to be selected in code layout rules and passed to the root
	// template further.
	// Because not all the plenty of generated tiny Go types and their usages
	// (e.g. an `int` field in a struct in depths of the code) deserve a separate root template call.
	// So they are rendered recursively by the "selectable" objects, containing the Promises that point to them.
	Selectable() bool
	// Visible returns false if object is set not to be rendered because of configuration, x-ignore field, etc.
	Visible() bool
	// Pointer returns the JSON pointer to the document URL and the position where an object is located.
	Pointer() jsonpointer.JSONPointer
	// String is just a string representation of the object for logging and debugging purposes.
	String() string
}

// GolangType is an Artifact, that represents a primitive Go type, such as struct, map, type alias, etc.
// All of these types are located in [render/lang] package.
type GolangType interface {
	Artifact
	// CanBeAddressed returns true if we're able to define a pointer to this type and take its value's address by
	// applying the & operator.
	//
	// Values that always *not addressable* are `nil`, values of interface type, constants, etc.
	CanBeAddressed() bool
	// GoTemplate returns a template name that renders this particular type.
	GoTemplate() string
}

// artifactReferrer is an interface for artifacts that contain a reference to another artifact (such as Promise, Ref).
type artifactReferrer interface {
	// ReferredArtifact returns the artifact that is referred by this artifact. Non-recursive.
	ReferredArtifact() Artifact
}

// DerefArtifact recursively extracts the target artifact from the reference a. If a is not a reference, it returns a.
func DerefArtifact[T Artifact](a Artifact) T {
	// TODO: detect ref loops to avoid infinite recursion
	w, ok := a.(artifactReferrer)
	for ok {
		a = w.ReferredArtifact()
		if lo.IsNil(a) {
			panic(fmt.Sprintf("reference points to nil, cannot dereference: %T", w))
		}
		w, ok = a.(artifactReferrer)
	}
	return a.(T)
}

// golangTypeWrapper is an interface for Go types (such as pointers, enums) that are able to
// wrap another [common.GolangType] type.
type golangTypeWrapper interface {
	// WrappedGolangType returns the GolangType that is wrapped by this wrapper type. Non-recursive.
	WrappedGolangType() GolangType
}

// UnwrapGolangType recursively unwraps the target Go type from Go type t (such as a pointer, enum, etc.).
// If t is not a wrapper type, it returns t.
func UnwrapGolangType[T GolangType](t GolangType) T {
	// TODO: detect ref loops to avoid infinite recursion
	w, ok := t.(golangTypeWrapper)
	for ok {
		t = w.WrappedGolangType()
		if lo.IsNil(t) {
			panic(fmt.Sprintf("wrapper type points to nil, cannot unwrap: %T", w))
		}
		w, ok = t.(golangTypeWrapper)
	}
	return t.(T)
}

// CheckSameArtifacts checks if two artifacts are the same object or points to the same object through references.
func CheckSameArtifacts(a, b Artifact) bool {
	return DerefArtifact[Artifact](a) == DerefArtifact[Artifact](b)
}
