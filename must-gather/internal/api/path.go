package api

import (
	"fmt"
	"os"
	"path/filepath"

	"k8s.io/apimachinery/pkg/runtime/schema"
)

type Path struct {
	string
}

func NewArtifactPath(parts ...string) Path {
	return Path{filepath.Join(parts...)}
}

func (p Path) Add(parts ...string) Path {
	parts = append([]string{p.string}, parts...)
	p.string = filepath.Join(parts...)
	return p
}

func (p Path) ForResource(gvr schema.GroupVersionResource) Path {
	p.string = filepath.Join(p.string, gvr.Group, gvr.Resource)
	return p
}

func (p Path) String() string {
	return p.string
}

func (p Path) MkdirAll() error {
	dir := filepath.Dir(p.string)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", dir, err)
	}
	return nil
}

func (p Path) WriteFile(filePath Path, data []byte) error {
	if err := os.WriteFile(filePath.String(), data, 0644); err != nil {
		return fmt.Errorf("failed to write file %s: %w", filePath, err)
	}
	return nil
}
