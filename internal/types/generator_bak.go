package types

//
// import (
//	"bytes"
//	"fmt"
//	"github.com/a-h/generate"
//	"github.com/bdragon300/asyncapi-codegen/internal/interfaces"
//	"strings"
//	"unicode"
//)
//
//
//// Generator will produce structs from the JSON schema.
// type Generator struct {
//	schemas  []*generate.Schema
//	resolver *RefResolver
//	Structs  map[string]Struct
//	Aliases  map[string]Field
//	// cache for reference types; k=url v=type
//	refs      map[string]string
//	anonCount int
//}
//
//// New creates an instance of a generator which will produce structs.
// func New(schemas ...*Schema) *Generator {
//	return &Generator{
//		schemas:  schemas,
//		resolver: NewRefResolver(schemas),
//		Structs:  make(map[string]Struct),
//		Aliases:  make(map[string]Field),
//		refs:     make(map[string]string),
//	}
//}
//
//// CreateTypes creates types from the JSON schemas, keyed by the golang name.
// func (g *Generator) CreateTypes() (err error) {
//	if err := g.resolver.Init(); err != nil {
//		return err
//	}
//
//	// extract the types
//	for _, schema := range g.schemas {
//		name := g.getSchemaName("", schema)
//		rootType, err := g.processSchema(name, schema)
//		if err != nil {
//			return err
//		}
//		// ugh: if it was anything but a struct the type will not be the name...
//		if rootType != "*"+name {
//			a := Field{
//				PreferredName:        name,
//				JSONName:    "",
//				Type:        rootType,
//				Required:    false,
//				Description: schema.Description,
//			}
//			g.Aliases[a.PreferredName] = a
//		}
//	}
//	return
//}
//
//
// func contains(s []string, e string) bool {
//	for _, a := range s {
//		if a == e {
//			return true
//		}
//	}
//	return false
//}
//
//
//// return a name for this (sub-)schema.
// func (g *Generator) getSchemaName(keyName string, schema *Schema) string {
//	if len(schema.Title) > 0 {
//		return getGolangName(schema.Title)
//	}
//	if keyName != "" {
//		return getGolangName(keyName)
//	}
//	if schema.Parent == nil {
//		return "Root"
//	}
//	if schema.JSONKey != "" {
//		return getGolangName(schema.JSONKey)
//	}
//	if schema.Parent != nil && schema.Parent.JSONKey != "" {
//		return getGolangName(schema.Parent.JSONKey + "Item")
//	}
//	g.anonCount++
//	return fmt.Sprintf("Anonymous%d", g.anonCount)
//}
//
//// getGolangName strips invalid characters out of golang struct or field names.
// func getGolangName(s string) string {
//	buf := bytes.NewBuffer([]byte{})
//	for i, v := range splitOnAll(s, isNotAGoNameCharacter) {
//		if i == 0 && strings.IndexAny(v, "0123456789") == 0 {
//			// Go types are not allowed to start with a number, lets prefix with an underscore.
//			buf.WriteRune('_')
//		}
//		buf.WriteString(capitaliseFirstLetter(v))
//	}
//	return buf.String()
//}
//
// func splitOnAll(s string, shouldSplit func(r rune) bool) []string {
//	rv := []string{}
//	buf := bytes.NewBuffer([]byte{})
//	for _, c := range s {
//		if shouldSplit(c) {
//			rv = append(rv, buf.String())
//			buf.Reset()
//		} else {
//			buf.WriteRune(c)
//		}
//	}
//	if buf.Len() > 0 {
//		rv = append(rv, buf.String())
//	}
//	return rv
//}
//
// func isNotAGoNameCharacter(r rune) bool {
//	if unicode.IsLetter(r) || unicode.IsDigit(r) {
//		return false
//	}
//	return true
//}
//
// func capitaliseFirstLetter(s string) string {
//	if s == "" {
//		return s
//	}
//	prefix := s[0:1]
//	suffix := s[1:]
//	return strings.ToUpper(prefix) + suffix
//}
