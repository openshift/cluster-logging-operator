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

func NewPath(parts ...string) Path {
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
	if err := os.MkdirAll(p.string, 0755); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", p.string, err)
	}
	return nil
}

func (p Path) WriteFile(data []byte) error {
	// Create parent directory if it doesn't exist
	dir := filepath.Dir(p.string)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", dir, err)
	}

	if err := os.WriteFile(p.String(), data, 0644); err != nil {
		return fmt.Errorf("failed to write file %s: %w", p, err)
	}
	return nil
}
