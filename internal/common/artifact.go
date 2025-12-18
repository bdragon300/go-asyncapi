package common

import (
	"github.com/bdragon300/go-asyncapi/internal/jsonpointer"
)

// ArtifactKind is an enumeration of compiled artifact kinds.
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
	// ArtifactKindOther is a utility language object, not intended for selection (type, value, interface, etc.)
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
	// CanBeDereferenced returns true if this type is a pointer, so we can dereference it by using * operator.
	CanBeDereferenced() bool
	// GoTemplate returns a template name that renders this particular type.
	GoTemplate() string
}

type artifactWrapper interface {
	Unwrap() Artifact
}

// DerefArtifact returns the artifact, unwrapping it if it's wrapped in a Promise or Ref.
func DerefArtifact(obj Artifact) Artifact {
	// TODO: detect ref loops to avoid infinite recursion
	if w, ok := obj.(artifactWrapper); ok {
		return w.Unwrap()
	}
	return obj
}

// CheckSameArtifacts checks if two artifacts are the same object. If they are wrapped in Promises or Refs, it unwraps
// them first.
func CheckSameArtifacts(a, b Artifact) bool {
	return DerefArtifact(a) == DerefArtifact(b)
}
