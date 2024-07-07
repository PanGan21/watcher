package main

import (
	"fmt"
	"os"
	"time"
	watcher "watcher/internal"

	"github.com/spf13/cobra"
)

const (
	defaultTarget  = "./"
	defaultPattern = "**"
	waitForTerm    = 5 * time.Second
)

// var (
// 	targets   []string
// 	patterns  []string
// 	ignores   []string
// 	delay     time.Duration
// 	restart   bool
// 	sigopt    string
// 	filteropt []string
// 	verbose   bool
// )

func main() {
	var watcherOptions watcher.WatcherOptions

	var rootCmd = &cobra.Command{
		Use:   "watcher [COMMAND]",
		Short: "A file watching and command restarting tool",
		Long:  `Run the COMMAND and restart when a file matching the pattern has been modified.`,
		Args:  cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			watcher.Execute(cmd, args, watcherOptions)
		},
	}

	rootCmd.PersistentFlags().StringArrayVarP(&watcherOptions.Targets, "target", "t", []string{defaultTarget}, "observation target `path` (default \"./\")")
	rootCmd.PersistentFlags().StringArrayVarP(&watcherOptions.Patterns, "pattern", "p", []string{defaultPattern}, "trigger pathname `glob` pattern (default \"**\")")
	rootCmd.PersistentFlags().StringArrayVarP(&watcherOptions.Ignores, "ignore", "i", nil, "ignore pathname `glob` pattern")
	rootCmd.PersistentFlags().DurationVarP(&watcherOptions.Delay, "delay", "d", time.Second, "`duration` to delay the restart of the command")
	rootCmd.PersistentFlags().BoolVarP(&watcherOptions.Restart, "restart", "r", false, "restart the command on exit")
	rootCmd.PersistentFlags().StringVarP(&watcherOptions.Sigopt, "signal", "s", "", "`signal` used to stop the command (default \"SIGTERM\")")
	rootCmd.PersistentFlags().BoolVarP(&watcherOptions.Verbose, "verbose", "v", false, "verbose output")
	rootCmd.PersistentFlags().StringArrayVarP(&watcherOptions.Filteropt, "filter", "f", nil, "filter file system `event` (CREATE|WRITE|REMOVE|RENAME|CHMOD)")

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
