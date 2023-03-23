// Package file provides types that represent a repository interface and it's
// methods, as well as go types.
package file

import (
	"fmt"
	"path/filepath"
	"strings"
)

// ImportManager is a helper used to manager conflicting imports.
type ImportManager struct {
	// Imports is a map of import names to their paths, used to prevent
	// collisions.
	Imports map[string]string // map[importName]importPath
}

func NewImportManager() *ImportManager {
	return &ImportManager{Imports: make(map[string]string)}
}

func (m *ImportManager) Add(path string) string {
	if path == "" {
		return ""
	}

	name := filepath.Base(path)
	name = strings.SplitN(name, "-", 2)[0]
	name = strings.SplitN(name, "_", 2)[0]

	// fast path -- no collision
	if cmpPath, ok := m.Imports[name]; !ok || path == cmpPath {
		m.Imports[name] = path
		return name
	}

	for i := 0; ; i++ {
		safeName := fmt.Sprintf("%s%d", name, i)
		if cmpPath, ok := m.Imports[safeName]; ok && path != cmpPath {
			continue
		}

		m.Imports[safeName] = path
		return safeName
	}
}
