package manager

import (
	"fmt"
	"slices"
	"strings"

	"github.com/bdragon300/go-asyncapi/internal/common"
	"github.com/samber/lo"
)

// NamespaceManager manages the template namespace, that is used for conditional rendering functionality in the templates.
// It keeps the rendered declarations of artifacts and names that was explicitly defined in the template code.
// This functionality could remind the preprocessor namespace in C/C++, but for Go templates.
type NamespaceManager struct {
	artifacts []NamespaceArtifactItem
	names     []string
}

// DeclareArtifact adds an artifact declaration to the namespace.
//
// Rendered and pinned object declaration flags are cumulative, meaning that once each one became true, it will remain true in
// next calls, ignoring the passed value. All these flags affects the conditional rendering functions in templates`.
//
// If passed is true, it indicates that the control flow in templates has passed this artifact. E.g. `once` function
// has been called with this artifact as argument.
//
// If rendered is true, it indicates that the artifact (e.g. if it is [common.GolangType]) has been rendered in the template.
func (s *NamespaceManager) DeclareArtifact(a common.Artifact, renderManager *TemplateRenderManager, passed, rendered bool) {
	for i := range s.artifacts {
		if s.artifacts[i].Object == a {
			// once true, always true
			s.artifacts[i].Rendered = s.artifacts[i].Rendered || rendered
			s.artifacts[i].Passed = s.artifacts[i].Passed || passed
			return
		}
	}

	s.artifacts = append(s.artifacts, NamespaceArtifactItem{
		Object:      a,
		Layout:      renderManager.CurrentLayoutItem,
		Rendered:    rendered,
		Passed:      passed,
		FileName:    renderManager.FileName,
		PackageName: renderManager.PackageName,
	})
}

// DeclareName adds the name to the namespace.
func (s *NamespaceManager) DeclareName(name string) {
	if !lo.Contains(s.names, name) {
		s.names = append(s.names, name)
	}
}

// FindArtifactDeclaration looks the namespace for the artifact declaration. If found, returns the declaration and true.
func (s *NamespaceManager) FindArtifactDeclaration(obj common.Artifact) (NamespaceArtifactItem, bool) {
	return lo.Find(s.artifacts, func(def NamespaceArtifactItem) bool {
		return def.Object == obj
	})
}

// IsNameDeclared checks if the name is defined in the namespace.
func (s *NamespaceManager) IsNameDeclared(name string) bool {
	return lo.Contains(s.names, name)
}

func (s *NamespaceManager) Clone() *NamespaceManager {
	return &NamespaceManager{
		artifacts: slices.Clone(s.artifacts),
		names:     slices.Clone(s.names),
	}
}

func (s *NamespaceManager) String() string {
	defs := strings.Join(lo.Map(s.artifacts, func(item NamespaceArtifactItem, _ int) string {
		return fmt.Sprintf("[%[1]p] %[1]s", item.Object)
	}), "; ")
	return fmt.Sprintf("names: %s | artifacts: %s", strings.Join(s.names, "; "), defs)
}

type NamespaceArtifactItem struct {
	Object      common.Artifact
	Layout      common.ConfigLayoutItem
	FileName    string
	PackageName string
	Rendered    bool
	Passed      bool
}
