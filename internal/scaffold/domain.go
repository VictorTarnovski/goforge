package scaffold

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"text/template"

	"github.com/VictorTarnovski/goforge/internal/templates"
)

// DomainField is a Field augmented with the extra values the domain
// templates need: its positional SQL placeholder and an example literal
// used in generated tests.
type DomainField struct {
	Field
	Placeholder int
	Example     string
}

// DomainData is the template data for the generic domain template set, used
// both for `goforge new`'s baked-in example and for `goforge generate domain`.
type DomainData struct {
	Module       string
	Project      string
	Package      string
	Type         string
	Table        string
	Route        string
	Fields       []DomainField
	DB           bool
	HasAuthz     bool
	HasTimeField bool
	// FirstStringField is nil when the domain has no string field, so the
	// generated test can skip the "empty field" case entirely.
	FirstStringField *DomainField
}

// GenerateDomain renders a full vertical slice — model, service, repository,
// HTTP handler, and a table-driven test — into internal/<name>/, adapting to
// the project's manifest: with db enabled it also writes a numbered goose
// migration; without it, the repository is an in-memory map and no
// migration is written.
func GenerateDomain(projectDir string, m *Manifest, name string, fieldsSpec string) error {
	fields, err := ParseFields(fieldsSpec)
	if err != nil {
		return err
	}

	data := buildDomainData(m, name, fields)

	destDir := filepath.Join(projectDir, "internal", data.Package)

	files := map[string]string{
		"model.go.tmpl":     data.Package + ".go",
		"httpx.go.tmpl":     "httpx.go",
		"model_test.go.tmpl": data.Package + "_test.go",
	}
	if m.DB {
		files["repository_db.go.tmpl"] = "repository.go"
		files["service_db.go.tmpl"] = "service.go"
	} else {
		files["repository_memory.go.tmpl"] = "repository.go"
		files["service_memory.go.tmpl"] = "service.go"
	}

	for src, out := range files {
		if err := renderDomainFile(src, filepath.Join(destDir, out), data); err != nil {
			return err
		}
	}

	if m.DB {
		migrationsDir := filepath.Join(projectDir, "migrations")
		seq, err := nextMigrationSeq(migrationsDir)
		if err != nil {
			return fmt.Errorf("determine next migration number: %w", err)
		}

		migrationPath := filepath.Join(migrationsDir, fmt.Sprintf("%05d_create_%s.sql", seq, data.Table))
		if err := renderDomainFile("migration.sql.tmpl", migrationPath, data); err != nil {
			return err
		}
	}

	return nil
}

// PackageName derives the Go package name a domain named name is generated
// under, e.g. "blog_post" -> "blogpost".
func PackageName(name string) string {
	return strings.ToLower(PascalCase(name))
}

func buildDomainData(m *Manifest, name string, fields []Field) DomainData {
	domainFields := make([]DomainField, len(fields))
	var firstString *DomainField
	hasTime := false

	for i, f := range fields {
		df := DomainField{
			Field:       f,
			Placeholder: i + 2,
			Example:     exampleValue(f.GoType),
		}
		domainFields[i] = df

		if f.GoType == "string" && firstString == nil {
			firstString = &domainFields[i]
		}
		if f.GoType == "time.Time" {
			hasTime = true
		}
	}

	pkg := PackageName(name)

	return DomainData{
		Module:           m.Module,
		Project:          m.Project,
		Package:          pkg,
		Type:             PascalCase(name),
		Table:            pkg + "s",
		Route:            pkg + "s",
		Fields:           domainFields,
		DB:               m.DB,
		HasAuthz:         m.Authz != "",
		HasTimeField:     hasTime,
		FirstStringField: firstString,
	}
}

func renderDomainFile(srcName, destPath string, data DomainData) error {
	content, err := fs.ReadFile(templates.Domain, "domain/"+srcName)
	if err != nil {
		return fmt.Errorf("read domain template %s: %w", srcName, err)
	}

	tmpl, err := template.New(srcName).Parse(string(content))
	if err != nil {
		return fmt.Errorf("parse domain template %s: %w", srcName, err)
	}

	if err := os.MkdirAll(filepath.Dir(destPath), 0o755); err != nil {
		return err
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

func exampleValue(goType string) string {
	switch goType {
	case "string":
		return `"example"`
	case "int":
		return "1"
	case "bool":
		return "true"
	case "float64":
		return "1.5"
	case "time.Time":
		return "time.Now().UTC()"
	default:
		return `""`
	}
}

// nextMigrationSeq scans dir for goose-style "NNNNN_description.sql" files
// and returns one more than the highest sequence number found, or 1 if the
// directory has none yet.
func nextMigrationSeq(dir string) (int, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return 1, nil
		}
		return 0, err
	}

	max := 0
	for _, e := range entries {
		name := e.Name()
		i := strings.IndexByte(name, '_')
		if i <= 0 {
			continue
		}
		n, err := strconv.Atoi(name[:i])
		if err != nil {
			continue
		}
		if n > max {
			max = n
		}
	}
	return max + 1, nil
}
