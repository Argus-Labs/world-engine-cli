package root

import (
	"os"

	"github.com/spf13/cobra"

	"pkg.world.dev/world-cli/cmd/world/cardinal"
	"pkg.world.dev/world-cli/cmd/world/evm"
	"pkg.world.dev/world-cli/common/config"
	"pkg.world.dev/world-cli/common/logger"
	"pkg.world.dev/world-cli/tea/style"
)

// rootCmd represents the base command
// Usage: `world`
var rootCmd = &cobra.Command{
	Use:   "world",
	Short: "A swiss army knife for World Engine development",
	Long:  style.CLIHeader("World CLI", "A swiss army knife for World Engine projects"),
}

func init() {
	// Enable case-insensitive commands
	cobra.EnableCaseInsensitive = true

	// Register groups
	rootCmd.AddGroup(&cobra.Group{ID: "starter", Title: "Getting Started:"})
	rootCmd.AddGroup(&cobra.Group{ID: "core", Title: "Tools:"})

	// Register base commands
	doctorCmd := getDoctorCmd(os.Stdout)
	createCmd := getCreateCmd(os.Stdout)
	rootCmd.AddCommand(createCmd, doctorCmd, versionCmd)

	// Register subcommands
	rootCmd.AddCommand(cardinal.BaseCmd)
	rootCmd.AddCommand(evm.BaseCmd)

	config.AddConfigFlag(rootCmd)

	// Add --debug flag
	logger.AddLogFlag(createCmd)
	logger.AddLogFlag(doctorCmd)
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		logger.Errors(err)
	}
	// print log stack
	logger.PrintLogs()
}
