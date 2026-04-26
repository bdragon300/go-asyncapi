package common

const SchemaTagName = "cgen"

// SchemaTag marks entity in certain place in the AsyncAPI document to set quirks for compilation logic when this
// entity is placed by that path. For example, we mark by tag an asyncapi.Object located in message payload and headers as "data model",
// which additionally causes every field in their Go definition to have field tags, althrough all other schemas defined in
// other places don't have this.
//
// Tags may be inherited -- once met it is applied to marked entity and to all nested entities down during compilation.
// For example, all subschemas of "data model" schema are also "data models" and have the appropriate tag applied,
// regardless of whether they have the tag or not.
//
// If tag is not inherited, then it is applied only to the entity in place it is met.
// Note: the entity place in document may contain $ref, so the tag is applied only for it, but not for the referenced entity.
type SchemaTag string

const (
	// SchemaTagSelectable marks the objects that should get to selections, i.e. basically objects that are
	// rendered directly by feeding to the root template. See [renderer.RenderArtifacts]. Non-inherited tag.
	SchemaTagSelectable SchemaTag = "selectable"

	// SchemaTagDataModel marks the jsonschema objects and all its nested objects as data models.
	// In particular, the tags in fields in data models are driven by document and message content types.
	//
	// Typically, this tag marks the message payload and header entities and also common-used document schemas.
	// Inherited tag.
	SchemaTagDataModel SchemaTag = "data_model"
)
