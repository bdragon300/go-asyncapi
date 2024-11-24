package common

// ObjectKind is an enumeration of all possible object kinds used in the AsyncAPI specification.
type ObjectKind string

const (
	ObjectKindLang ObjectKind = "lang" // Utility language object, not a spec component (type, value, interface, etc.)
	ObjectKindSchema ObjectKind = "schema"
	ObjectKindServer ObjectKind = "server"
	ObjectKindServerVariable ObjectKind = "serverVariable"
	ObjectKindChannel	ObjectKind = "channel"
	ObjectKindMessage ObjectKind = "message"
	ObjectKindParameter ObjectKind = "parameter"
	ObjectKindCorrelationID ObjectKind = "correlationID"
	ObjectKindServerBindings ObjectKind = "serverBindings"
	ObjectKindChannelBindings ObjectKind = "channelBindings"
	ObjectKindMessageBindings ObjectKind = "messageBindings"
	ObjectKindOperationBindings ObjectKind = "operationBindings"
)

type Renderer interface {
	Kind() ObjectKind
	// Selectable returns true if object can be selected to pass to the templates for rendering.
	Selectable() bool
}

type RenderOpts struct {
	RuntimeModule string
	ImportBase    string
	TargetPackage string
	TargetDir    string
}

type RenderContext interface {
	RuntimeModule(subPackage string) string
	GeneratedModule(subPackage string) string
	QualifiedName(packageExpr string) string
	QualifiedGeneratedName(subPackage, name string) string
	QualifiedRuntimeName(subPackage, name string) string
	SpecProtocols() []string
}
