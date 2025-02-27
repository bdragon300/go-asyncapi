package common

import (
	"github.com/bdragon300/go-asyncapi/internal/jsonpointer"
)

// ArtifactKind is an enumeration of compiled artifact kinds.
type ArtifactKind string

const (
	// ArtifactKindOther is a utility language object, not intended for selection (type, value, interface, etc.)
	ArtifactKindOther     ArtifactKind = ""
	ArtifactKindSchema    ArtifactKind = "schema"
	ArtifactKindServer    ArtifactKind = "server"
	ArtifactKindChannel   ArtifactKind = "channel"
	ArtifactKindOperation ArtifactKind = "operation"
	ArtifactKindMessage   ArtifactKind = "message"
	ArtifactKindParameter ArtifactKind = "parameter"
	// ArtifactKindAsyncAPI represents the root AsyncAPI object.
	ArtifactKindAsyncAPI = "asyncapi"
)

// Artifact is an compiled object that can be rendered in the template.
type Artifact interface {
	// Name returns the name of the object as it was defined in the AsyncAPI document. This method is suitable
	// for rendering the object through a ref. So we can render the object under ref's Name, which is necessary,
	// for example, for rendering servers, channels, etc.
	Name() string
	Kind() ArtifactKind
	// Selectable returns true if object can be picked for selections to invoke the template. If false, the object
	// does not get to selections but still can be indirectly rendered inside the templates.
	Selectable() bool
	// Visible returns true if object contents is visible in rendered code.
	Visible() bool
	// Pointer returns the JSON pointer to the object in the document.
	Pointer() jsonpointer.JSONPointer
	// String is just a string representation of the object for logging and debugging purposes.
	String() string
}

// GolangType is an Artifact variation, that represents a primitive Go type, such as struct, map, type alias, etc.
// All of these types are located in [render/lang] package.
type GolangType interface {
	Artifact
	// CanBeAddressed returns true if value of this type could be addressed. Therefore, we're able to define a pointer
	// to this type, and we can take value's address by applying the & operator.
	//
	// Values that always *not addressable* typically are `nil`, values of interface type, constants, etc.
	CanBeAddressed() bool
	// CanBeDereferenced returns true if this type is a pointer, so we can dereference it by using * operator.
	// True basically means that the type is a pointer as well.
	CanBeDereferenced() bool
	// GoTemplate returns a template name that renders an object of this type.
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
