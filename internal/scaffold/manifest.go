package scaffold

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// ManifestFile is the name of the manifest goforge writes into every project it
// scaffolds, so later commands (generate domain) can read back the choices
// made at `new` time instead of guessing from directory contents.
const ManifestFile = ".goforge.yaml"

// Manifest records the choices made when a project was scaffolded.
type Manifest struct {
	Module  string `yaml:"module"`
	Project string `yaml:"project"`
	DB      bool   `yaml:"db"`
	Authz   string `yaml:"authz,omitempty"` // "" or "openfga"
	Deploy  string `yaml:"deploy,omitempty"` // "" or "nginx"
}

// Save writes the manifest to <dir>/.goforge.yaml.
func (m *Manifest) Save(dir string) error {
	data, err := yaml.Marshal(m)
	if err != nil {
		return fmt.Errorf("marshal manifest: %w", err)
	}

	path := filepath.Join(dir, ManifestFile)
	if err := os.WriteFile(path, data, 0o644); err != nil {
		return fmt.Errorf("write %s: %w", path, err)
	}
	return nil
}

// LoadManifest reads <dir>/.goforge.yaml, produced by a previous `goforge new`.
func LoadManifest(dir string) (*Manifest, error) {
	path := filepath.Join(dir, ManifestFile)

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("%s not found: run this from the root of a project scaffolded with `goforge new`", ManifestFile)
		}
		return nil, fmt.Errorf("read %s: %w", path, err)
	}

	var m Manifest
	if err := yaml.Unmarshal(data, &m); err != nil {
		return nil, fmt.Errorf("parse %s: %w", path, err)
	}
	return &m, nil
}
