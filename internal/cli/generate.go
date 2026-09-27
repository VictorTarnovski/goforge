package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/VictorTarnovski/goforge/internal/scaffold"
)

func newGenerateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "generate",
		Short: "Generate additional code into an existing goforge-scaffolded project",
	}

	cmd.AddCommand(newGenerateDomainCmd())

	return cmd
}

func newGenerateDomainCmd() *cobra.Command {
	var fields string

	cmd := &cobra.Command{
		Use:   "domain <name>",
		Short: "Generate a full vertical slice: model, service, repository, HTTP handler, and tests",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]

			dir, err := os.Getwd()
			if err != nil {
				return err
			}

			manifest, err := scaffold.LoadManifest(dir)
			if err != nil {
				return err
			}

			if err := scaffold.GenerateDomain(dir, manifest, name, fields); err != nil {
				return err
			}

			pkg := scaffold.PackageName(name)

			if err := scaffold.WireDomain(dir, manifest.Module, pkg); err != nil {
				return fmt.Errorf("generated internal/%s, but could not wire it into cmd/api/main.go: %w", pkg, err)
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Generated and wired internal/%s\n", pkg)
			if manifest.DB {
				fmt.Fprintln(cmd.OutOrStdout(), "Added a migration under migrations/ — run `make migrate` (or `go run ./cmd/ctl migrate`).")
			}

			scaffold.FormatGenerated(dir)

			return nil
		},
	}

	cmd.Flags().StringVar(&fields, "fields", "", "Comma-separated name:type pairs, e.g. name:string,count:int")

	return cmd
}
