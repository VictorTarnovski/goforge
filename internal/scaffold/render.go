package scaffold

import (
	"fmt"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"strings"
	"text/template"
)

// ProjectData is the template data available to every file in the base
// project tree.
type ProjectData struct {
	Module  string
	Project string
	DB      bool
	Authz   string // "" or "openfga"
	Deploy  string // "" or "nginx"
}

func (d ProjectData) HasAuthz() bool  { return d.Authz != "" }
func (d ProjectData) HasDeploy() bool { return d.Deploy != "" }
func (d ProjectData) Compose() bool   { return d.DB || d.HasAuthz() }

// EnvPrefix is the SCREAMING_SNAKE_CASE environment variable prefix viper
// binds to, derived from the project name (e.g. "widget-api" -> "WIDGET_API").
func (d ProjectData) EnvPrefix() string {
	return strings.ToUpper(strings.NewReplacer("-", "_", " ", "_").Replace(d.Project))
}

// conditionalPaths are subtrees of the base template that are only rendered
// when their predicate holds. A path not listed here is always rendered.
// Paths use forward slashes and are matched against the template-relative
// path (before any .tmpl suffix is stripped).
var conditionalPaths = []struct {
	path string
	when func(ProjectData) bool
}{
	{"internal/db", func(d ProjectData) bool { return d.DB }},
	{"internal/uow", func(d ProjectData) bool { return d.DB }},
	{"migrations", func(d ProjectData) bool { return d.DB }},
	{"cmd/ctl/migrate.go.tmpl", func(d ProjectData) bool { return d.DB }},
	{"internal/authz", ProjectData.HasAuthz},
	{"deploy", ProjectData.HasDeploy},
	{"docker-compose.yml.tmpl", ProjectData.Compose},
}

func included(rel string, data ProjectData) bool {
	rel = filepath.ToSlash(rel)
	for _, c := range conditionalPaths {
		if rel == c.path || strings.HasPrefix(rel, c.path+"/") {
			return c.when(data)
		}
	}
	return true
}

// RenderTree walks srcFS (rooted at root) and writes the rendered result
// under destDir. Files ending in .tmpl are executed as Go text/template
// (with the suffix stripped from the output name); everything else is
// copied byte-for-byte. Subtrees listed in conditionalPaths are skipped
// when their predicate is false.
func RenderTree(srcFS fs.FS, root string, destDir string, data ProjectData) error {
	return fs.WalkDir(srcFS, root, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		rel, err := filepath.Rel(root, p)
		if err != nil {
			return err
		}
		if rel == "." {
			return nil
		}
		rel = filepath.ToSlash(rel)

		if !included(rel, data) {
			if d.IsDir() {
				return fs.SkipDir
			}
			return nil
		}

		destPath := filepath.Join(destDir, filepath.FromSlash(rel))

		if d.IsDir() {
			return os.MkdirAll(destPath, 0o755)
		}

		if err := os.MkdirAll(filepath.Dir(destPath), 0o755); err != nil {
			return err
		}

		content, err := fs.ReadFile(srcFS, p)
		if err != nil {
			return fmt.Errorf("read template %s: %w", p, err)
		}

		if strings.HasSuffix(destPath, ".tmpl") {
			return renderFile(path.Base(p), content, strings.TrimSuffix(destPath, ".tmpl"), data)
		}

		return os.WriteFile(destPath, content, 0o644)
	})
}

func renderFile(name string, content []byte, destPath string, data any) error {
	tmpl, err := template.New(name).Parse(string(content))
	if err != nil {
		return fmt.Errorf("parse template %s: %w", name, err)
	}

	f, err := os.Create(destPath)
	if err != nil {
		return fmt.Errorf("create %s: %w", destPath, err)
	}
	defer f.Close()

	if err := tmpl.Execute(f, data); err != nil {
		return fmt.Errorf("render %s: %w", destPath, err)
	}
	return nil
}

// CopyTree copies srcFS (rooted at root) into destDir verbatim, with no
// templating and no conditionals — used for vendoring the skills library.
func CopyTree(srcFS fs.FS, root string, destDir string) error {
	return fs.WalkDir(srcFS, root, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		rel, err := filepath.Rel(root, p)
		if err != nil {
			return err
		}
		if rel == "." {
			return nil
		}

		destPath := filepath.Join(destDir, filepath.FromSlash(filepath.ToSlash(rel)))

		if d.IsDir() {
			return os.MkdirAll(destPath, 0o755)
		}

		if err := os.MkdirAll(filepath.Dir(destPath), 0o755); err != nil {
			return err
		}

		content, err := fs.ReadFile(srcFS, p)
		if err != nil {
			return fmt.Errorf("read %s: %w", p, err)
		}

		return os.WriteFile(destPath, content, 0o644)
	})
}
