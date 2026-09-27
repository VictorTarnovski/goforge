package scaffold

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/VictorTarnovski/goforge/internal/templates"
)

// NewOptions are the choices behind `goforge new`.
type NewOptions struct {
	Dir     string // destination directory
	Module  string
	Project string
	DB      bool
	Authz   string // "" or "openfga"
	Deploy  string // "" or "nginx"
}

// New scaffolds a project into opts.Dir: the base skeleton, the vendored
// conventions library, a manifest, and a baked-in "gopher" example domain
// demonstrating the pattern end-to-end.
func New(opts NewOptions) error {
	if err := os.MkdirAll(opts.Dir, 0o755); err != nil {
		return fmt.Errorf("create %s: %w", opts.Dir, err)
	}

	projectData := ProjectData{
		Module:  opts.Module,
		Project: opts.Project,
		DB:      opts.DB,
		Authz:   opts.Authz,
		Deploy:  opts.Deploy,
	}

	if err := RenderTree(templates.Base, "base", opts.Dir, projectData); err != nil {
		return fmt.Errorf("render project skeleton: %w", err)
	}

	if err := CopyTree(templates.Skills, "skills", opts.Dir); err != nil {
		return fmt.Errorf("vendor conventions library: %w", err)
	}

	manifest := &Manifest{
		Module:  opts.Module,
		Project: opts.Project,
		DB:      opts.DB,
		Authz:   opts.Authz,
		Deploy:  opts.Deploy,
	}
	if err := manifest.Save(opts.Dir); err != nil {
		return fmt.Errorf("write manifest: %w", err)
	}

	if err := GenerateDomain(opts.Dir, manifest, "gopher", "name:string,color:string,favorite_food:string"); err != nil {
		return fmt.Errorf("generate example domain: %w", err)
	}

	return nil
}

// TidyModules runs `go mod tidy` in dir if a go toolchain is on PATH. It is
// best-effort: goforge's own templates carry no go.sum, so without this step
// the generated project simply won't build until the caller runs it by hand.
func TidyModules(dir string) error {
	if _, err := exec.LookPath("go"); err != nil {
		return errGoNotFound
	}

	cmd := exec.Command("go", "mod", "tidy")
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("go mod tidy: %w\n%s", err, out)
	}
	return nil
}

var errGoNotFound = fmt.Errorf("go toolchain not found on PATH")

// ErrGoNotFound reports whether err is the sentinel TidyModules returns when
// there is no go binary to run.
func ErrGoNotFound(err error) bool {
	return err == errGoNotFound
}

// FormatGenerated best-effort formats and auto-fixes generated code: `gofmt
// -w` if a go toolchain is on PATH, then `golangci-lint run --fix` if that's
// on PATH too. Templates are hand-written and occasionally drift from exact
// gofmt/wsl_v5 spacing; this step (not the templates) is what keeps
// generated output clean, so it silently does nothing when either tool is
// absent rather than failing the scaffold.
func FormatGenerated(dir string) {
	if _, err := exec.LookPath("gofmt"); err == nil {
		cmd := exec.Command("gofmt", "-w", ".")
		cmd.Dir = dir
		_ = cmd.Run()
	}

	if _, err := exec.LookPath("golangci-lint"); err == nil {
		cmd := exec.Command("golangci-lint", "run", "--fix", "./...")
		cmd.Dir = dir
		_ = cmd.Run()
	}
}
