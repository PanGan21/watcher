package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func main() {
	var rootCmd = &cobra.Command{
		Use:   "watcher [COMMAND]",
		Short: "A file watching and command restarting tool",
		Long:  `Run the COMMAND and restart when a file matching the pattern has been modified.`,
		Args:  cobra.MaximumNArgs(1),
		Run:   execute,
	}

	if len(os.Args) < 2 {
		rootCmd.Help()
		os.Exit(0)
	}

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func execute(cmd *cobra.Command, args []string) {

}
