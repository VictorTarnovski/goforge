package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/VictorTarnovski/goforge/internal/scaffold"
)

func newNewCmd() *cobra.Command {
	var module string
	var db bool
	var authz string
	var deploy string

	cmd := &cobra.Command{
		Use:   "new <project-name>",
		Short: "Scaffold a new project",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			project := args[0]

			if module == "" {
				return fmt.Errorf("--module is required, e.g. --module github.com/you/%s", project)
			}
			if authz != "" && authz != "openfga" {
				return fmt.Errorf("--authz must be \"openfga\" (or omitted)")
			}
			if deploy != "" && deploy != "nginx" {
				return fmt.Errorf("--deploy must be \"nginx\" (or omitted)")
			}

			dir, err := filepath.Abs(project)
			if err != nil {
				return err
			}
			if entries, readErr := os.ReadDir(dir); readErr == nil && len(entries) > 0 {
				return fmt.Errorf("%s already exists and is not empty", dir)
			}

			err = scaffold.New(scaffold.NewOptions{
				Dir:     dir,
				Module:  module,
				Project: project,
				DB:      db,
				Authz:   authz,
				Deploy:  deploy,
			})
			if err != nil {
				return err
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Scaffolded %s in %s\n", project, dir)

			if tidyErr := scaffold.TidyModules(dir); tidyErr != nil {
				if scaffold.ErrGoNotFound(tidyErr) {
					fmt.Fprintln(cmd.OutOrStdout(), "go toolchain not found on PATH: run `go mod tidy` inside the project yourself.")
				} else {
					fmt.Fprintf(cmd.OutOrStdout(), "go mod tidy failed, run it yourself to see why:\n%v\n", tidyErr)
				}
			}

			scaffold.FormatGenerated(dir)

			return nil
		},
	}

	cmd.Flags().StringVar(&module, "module", "", "Go module path for the new project (required)")
	cmd.Flags().BoolVar(&db, "db", false, "Include a Postgres-backed database layer (pgx + goose)")
	cmd.Flags().StringVar(&authz, "authz", "", "Authorization backend to include (openfga)")
	cmd.Flags().StringVar(&deploy, "deploy", "", "Deployment config to include (nginx)")

	return cmd
}
