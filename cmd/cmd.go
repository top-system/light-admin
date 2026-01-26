package cmd

import (
	"errors"
	"os"

	"github.com/top-system/light-admin/cmd/migrate"
	"github.com/top-system/light-admin/cmd/runserver"
	"github.com/top-system/light-admin/cmd/setup"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(runserver.StartCmd)
	rootCmd.AddCommand(migrate.StartCmd)
	rootCmd.AddCommand(setup.StartCmd)
}

var rootCmd = &cobra.Command{
	Use:          "light-admin",
	Short:        "light-admin",
	SilenceUsage: true,
	Long:         `light-admin`,
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) < 1 {
			return errors.New(
				"requires at least one arg, " +
					"you can view the available parameters through `--help`",
			)
		}
		return nil
	},
	PersistentPreRunE: func(*cobra.Command, []string) error { return nil },
	Run:               func(cmd *cobra.Command, args []string) {},
}

//Execute : apply commands
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(-1)
	}
}
