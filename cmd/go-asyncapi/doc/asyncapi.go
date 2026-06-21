package doc

import "strings"

// asyncapiObjectsStructure describes the structure of the AsyncAPI objects, so the object kind can be determined by a given
// node path in the document.
//
// Every root key is either a key in the document (e.g. "channels", "components") or an object kind (the latter starts
// with the ">" character, e.g. ">channel"). A value is a map of nested keys, where every value refers either to an
// object kind (starts with ">") or to another root key. The empty "" nested key matches any key or array index at that
// level and is used for maps and arrays of objects.
var asyncapiObjectsStructure = map[string]map[string]string{
	// Document and "components" section container keys (maps and arrays of objects)
	"info": {
		"contact":      ">contact",
		"license":      ">license",
		"tags":         "tags",
		"externalDocs": ">externalDocs",
	},
	"servers": {
		"": ">server",
	},
	"channels": {
		"": ">channel",
	},
	"operations": {
		"": ">operation",
	},
	"components": {
		"schemas":           "schemas",
		"servers":           "servers",
		"channels":          "channels",
		"operations":        "operations",
		"messages":          "messages",
		"securitySchemes":   "securitySchemes",
		"serverVariables":   "serverVariables",
		"parameters":        "parameters",
		"correlationIds":    "correlationIds",
		"replies":           "replies",
		"replyAddresses":    "replyAddresses",
		"externalDocs":      "externalDocs",
		"tags":              "tags",
		"operationTraits":   "operationTraits",
		"messageTraits":     "messageTraits",
		"serverBindings":    "serverBindings",
		"channelBindings":   "channelBindings",
		"operationBindings": "operationBindings",
		"messageBindings":   "messageBindings",
	},
	"schemas":           {"": ">schema"},
	"messages":          {"": ">message"},
	"securitySchemes":   {"": ">securityScheme"},
	"serverVariables":   {"": ">serverVariable"},
	"parameters":        {"": ">parameter"},
	"correlationIds":    {"": ">correlationId"},
	"replies":           {"": ">operationReply"},
	"replyAddresses":    {"": ">operationReplyAddress"},
	"externalDocs":      {"": ">externalDocs"},
	"tags":              {"": ">tag"},
	"operationTraits":   {"": ">operationTrait"},
	"messageTraits":     {"": ">messageTrait"},
	"messageExamples":   {"": ">messageExample"},
	"serverBindings":    {"": ">serverBindings"},
	"channelBindings":   {"": ">channelBindings"},
	"operationBindings": {"": ">operationBindings"},
	"messageBindings":   {"": ">messageBindings"},

	// Objects
	">contact": {},
	">license": {},
	">server": {
		"variables":    "serverVariables",
		"security":     "securitySchemes",
		"tags":         "tags",
		"externalDocs": ">externalDocs",
		"bindings":     ">serverBindings",
	},
	">serverVariable": {},
	">channel": {
		"messages":     "messages",
		"servers":      "servers",
		"parameters":   "parameters",
		"tags":         "tags",
		"externalDocs": ">externalDocs",
		"bindings":     ">channelBindings",
	},
	">parameter": {},
	">operation": {
		"channel":      ">channel",
		"security":     "securitySchemes",
		"tags":         "tags",
		"externalDocs": ">externalDocs",
		"bindings":     ">operationBindings",
		"traits":       "operationTraits",
		"messages":     "messages",
		"reply":        ">operationReply",
	},
	">operationTrait": {
		"security":     "securitySchemes",
		"tags":         "tags",
		"externalDocs": ">externalDocs",
		"bindings":     ">operationBindings",
	},
	">operationReply": {
		"address":  ">operationReplyAddress",
		"channel":  ">channel",
		"messages": "messages",
	},
	">operationReplyAddress": {},
	">message": {
		"headers":       ">schema",
		"payload":       ">schema",
		"correlationId": ">correlationId",
		"tags":          "tags",
		"externalDocs":  ">externalDocs",
		"bindings":      ">messageBindings",
		"examples":      "messageExamples",
		"traits":        "messageTraits",
	},
	">messageTrait": {
		"headers":       ">schema",
		"correlationId": ">correlationId",
		"tags":          "tags",
		"externalDocs":  ">externalDocs",
		"bindings":      ">messageBindings",
		"examples":      "messageExamples",
	},
	">messageExample": {},
	">correlationId":  {},
	">tag": {
		"externalDocs": ">externalDocs",
	},
	">externalDocs": {},
	">securityScheme": {
		"flows": ">oauthFlows",
	},
	">oauthFlows": {
		"implicit":          ">oauthFlow",
		"password":          ">oauthFlow",
		"clientCredentials": ">oauthFlow",
		"authorizationCode": ">oauthFlow",
	},
	">oauthFlow": {},
	">schema": {
		"properties":           "schemas",
		"patternProperties":    "schemas",
		"definitions":          "schemas",
		"$defs":                "schemas",
		"allOf":                "schemas",
		"anyOf":                "schemas",
		"oneOf":                "schemas",
		"items":                ">schema",
		"additionalItems":      ">schema",
		"additionalProperties": ">schema",
		"propertyNames":        ">schema",
		"contains":             ">schema",
		"not":                  ">schema",
		"if":                   ">schema",
		"then":                 ">schema",
		"else":                 ">schema",
		"externalDocs":         ">externalDocs",
	},
	// Protocol-specific bindings objects (their children are keyed by protocol name and are leaf config)
	">serverBindings":    {},
	">channelBindings":   {},
	">operationBindings": {},
	">messageBindings":   {},
}

// componentsKeyByKind maps an object kind (a ">"-prefixed key of [asyncapiLocations]) to the key of the document
// "components" section that holds objects of this kind. Object kinds that have no dedicated "components" section
// (e.g. the Contact Object) are absent.
var componentsKeyByKind = map[string]string{
	">server":                "servers",
	">channel":               "channels",
	">operation":             "operations",
	">message":               "messages",
	">schema":                "schemas",
	">securityScheme":        "securitySchemes",
	">serverVariable":        "serverVariables",
	">parameter":             "parameters",
	">correlationId":         "correlationIds",
	">operationReply":        "replies",
	">operationReplyAddress": "replyAddresses",
	">externalDocs":          "externalDocs",
	">tag":                   "tags",
	">operationTrait":        "operationTraits",
	">messageTrait":          "messageTraits",
	">serverBindings":        "serverBindings",
	">channelBindings":       "channelBindings",
	">operationBindings":     "operationBindings",
	">messageBindings":       "messageBindings",
}

// asyncapiResolveComponentsKey resolves a node path of any depth into the key of the document "components" section that holds
// objects of the same kind as the node the path points to. The first path item is a top-level document key (e.g.
// "channels" or "components").
//
// For example, the path ["operations", "myop", "reply", "channel"] points to a Channel Object, so the result is
// "channels"; the path ["channels", "foobar", "servers", "0"] points to a Server Object, so the result is "servers".
//
// It returns an empty string if the path can't be resolved against the AsyncAPI schema, or if the object the path
// points to has no dedicated "components" section (e.g. the Info Object).
func asyncapiResolveComponentsKey(p []string) string {
	entity := resolveEntity(p)
	if entity == "" {
		return ""
	}

	if strings.HasPrefix(entity, ">") {
		// "" if this object kind has no dedicated "components" section
		return componentsKeyByKind[entity]
	}
	// The path ends on a container key (e.g. on the "servers" array itself). It maps to itself if it's a
	// "components" section key.
	for _, ck := range componentsKeyByKind {
		if ck == entity {
			return ck
		}
	}
	return ""
}

func asyncapiEntitiesSectionPaths() ([]string, []string) {
	rootKeys := []string{"servers", "channels", "operations"}
	componentsKeys := []string{
		"schemas",
		"servers", "channels", "operations", "messages",
		"securitySchemes", "serverVariables", "parameters", "correlationIds", "replies", "replyAddresses", "externalDocs", "tags",
		"operationTraits", "messageTraits",
		"serverBindings", "channelBindings", "operationBindings", "messageBindings",
	}
	return rootKeys, componentsKeys
}

// asyncapiUnresolvableRefPaths returns the $ref node paths that must not be resolved into the object they point to
// because the AsyncAPI schema explicitly requires the $ref there.
// The result is a list of paths, where each path is a list of keys to navigate from the document root to the $ref node.
// An empty item in the path matches any key at that level.
func asyncapiUnresolvableRefPaths() [][]string {
	return [][]string{
		{"channels", "", "servers", ""},
		{"components", "channels", "", "servers", ""},
		{"operations", "", "channel"},
		{"operations", "", "messages", ""},
		{"components", "operations", "", "channel"},
		{"components", "operations", "", "messages", ""},
		{"operations", "", "reply", "channel"},
		{"operations", "", "reply", "messages", ""},
		{"components", "operations", "", "reply", "channel"},
		{"components", "operations", "", "reply", "messages", ""},
		{"components", "replies", "", "channel"},
		{"components", "replies", "", "messages", ""},
	}
}

func asyncapiMandatoryRootPaths() []string {
	return []string{"asyncapi", "info"}
}

// resolveEntity resolves a node path of any depth into the tag denotes the entity the path points to.
// The returned value may be:
//   - a valid AsyncAPI key if path points to a container of objects (e.g. "channels" or "components")
//   - a ">"-prefixed key if path points to an object (e.g. ">channel" or ">message")
//   - an empty string if the path can't be resolved against the AsyncAPI schema or it contains invalid keys.
func resolveEntity(p []string) string {
	if len(p) == 0 {
		return ""
	}

	// kind holds the [asyncapiObjectsStructure] key that describes the node reached so far. The first path item is a
	// top-level document key, the rest navigate down the schema.
	kind := p[0]
	for _, seg := range p[1:] {
		loc, ok := asyncapiObjectsStructure[kind]
		if !ok {
			// Reached a leaf or an unknown object before the whole path was consumed
			return ""
		}
		next, ok := loc[seg]
		if !ok {
			// The empty key matches any map key or array index
			if next, ok = loc[""]; !ok {
				return ""
			}
		}
		kind = next
	}

	return kind
}
