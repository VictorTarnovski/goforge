package scaffold

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// WireDomain inserts the three lines cmd/api/main.go needs to serve a newly
// generated domain: its import, an entry in the `domains` slice, and an
// entry in `classifiers` — see the `// goforge:` markers left there by the
// base template. It is idempotent: re-running it for a domain already
// wired is a no-op.
func WireDomain(projectDir, module, pkg string) error {
	path := filepath.Join(projectDir, "cmd", "api", "main.go")

	content, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read %s: %w", path, err)
	}

	text := string(content)

	importLine := fmt.Sprintf("\t%q", module+"/internal/"+pkg)
	if strings.Contains(text, importLine) {
		return nil
	}

	text = insertBefore(text, "// goforge:import", importLine+"\n")
	text = insertBefore(text, "// goforge:domain", fmt.Sprintf("\t%s.Wire,\n", pkg))
	text = insertBefore(text, "// goforge:classifier", fmt.Sprintf("\t%s.ClassifyError,\n", pkg))

	if err := os.WriteFile(path, []byte(text), 0o644); err != nil {
		return fmt.Errorf("write %s: %w", path, err)
	}
	return nil
}

// insertBefore inserts line immediately before the start of the line
// containing marker. It is a no-op if marker is not found.
func insertBefore(text, marker, line string) string {
	idx := strings.Index(text, marker)
	if idx < 0 {
		return text
	}

	lineStart := strings.LastIndexByte(text[:idx], '\n') + 1
	return text[:lineStart] + line + text[lineStart:]
}
