package cli

import "github.com/spf13/cobra"

func Execute() error {
	return newRootCmd().Execute()
}

func newRootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:           "goforge",
		Short:         "Scaffold Go projects following your own conventions",
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	cmd.AddCommand(newNewCmd())
	cmd.AddCommand(newGenerateCmd())

	return cmd
}
