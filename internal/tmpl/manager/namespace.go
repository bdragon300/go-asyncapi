package manager

import (
	"fmt"
	"slices"
	"strings"

	"github.com/bdragon300/go-asyncapi/internal/common"
	"github.com/samber/lo"
)

// NamespaceManager manages the template namespace, that is used for conditional rendering functionality in the templates.
// It keeps the rendered declarations of artifacts and objects that was explicitly defined in the template code.
// This functionality could remind the preprocessor namespace in C/C++, but for Go templates.
type NamespaceManager struct {
	artifacts []NamespaceArtifactItem
	objects   []any
}

// DeclareArtifact adds an artifact declaration to the namespace.
//
// Rendered object declaration flag is cumulative, meaning that once it becomes true, it will remain true in
// next calls, ignoring the passed value.
//
// If rendered is true, it indicates that the artifact (e.g. if it is [common.GolangType]) has been rendered in the template.
func (s *NamespaceManager) DeclareArtifact(a common.Artifact, renderManager *TemplateRenderManager, rendered bool) {
	for i := range s.artifacts {
		if s.artifacts[i].Object == a {
			// once true, always true
			s.artifacts[i].Rendered = s.artifacts[i].Rendered || rendered
			return
		}
	}

	s.artifacts = append(s.artifacts, NamespaceArtifactItem{
		Object:      a,
		Layout:      renderManager.CurrentLayoutItem,
		Rendered:    rendered,
		FileName:    renderManager.FileName,
		PackageName: renderManager.PackageName,
	})
}

// FindArtifact looks the namespace for the artifact declaration. If found, returns the declaration and true.
func (s *NamespaceManager) FindArtifact(obj common.Artifact) (NamespaceArtifactItem, bool) {
	return lo.Find(s.artifacts, func(def NamespaceArtifactItem) bool {
		return def.Object == obj
	})
}

// AddObject adds any object to the namespace.
func (s *NamespaceManager) AddObject(o any) {
	if !lo.Contains(s.objects, o) {
		s.objects = append(s.objects, o)
	}
}

// ContainsObject checks if an object has been added to the namespace.
func (s *NamespaceManager) ContainsObject(o any) bool {
	return lo.Contains(s.objects, o)
}

func (s *NamespaceManager) Clone() *NamespaceManager {
	return &NamespaceManager{
		artifacts: slices.Clone(s.artifacts),
		objects:   slices.Clone(s.objects),
	}
}

func (s *NamespaceManager) String() string {
	a := lo.Map(s.artifacts, func(item NamespaceArtifactItem, _ int) string {
		return fmt.Sprintf("[%[1]p] %[1]s", item.Object)
	})
	o := lo.Map(s.objects, func(item any, _ int) string {
		return fmt.Sprintf("%[1]T[%[1]s]", item)
	})
	return fmt.Sprintf("objects: %s | artifacts: %s", strings.Join(o, "; "), strings.Join(a, "; "))
}

type NamespaceArtifactItem struct {
	Object      common.Artifact
	Layout      common.ConfigLayoutItem
	FileName    string
	PackageName string
	Rendered    bool
}
