package common

type PackageItem struct {
	Typ  Renderer
	Path []string
}

type Package struct {
	items []PackageItem
}

func (m *Package) Put(obj Renderer, pathStack []string) {
	m.items = append(m.items, PackageItem{
		Typ:  obj,
		Path: pathStack,
	})
}

func (m *Package) Items() []PackageItem {
	return m.items
}
