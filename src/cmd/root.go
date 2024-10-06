package cmd

import (
	"fmt"
	"os"
	"stash/src/actions"
	"strconv"

	"github.com/common-nighthawk/go-figure"
	"github.com/spf13/cobra"
)

func init() {
	// TODO
	// - flatten tree
	// - individual apply
	// - restore

	rootCmd.AddCommand(apply)
	rootCmd.AddCommand(version)
	rootCmd.AddCommand(restore)
	rootCmd.AddCommand(revert)

	initArgs()
}

var version = &cobra.Command{
	Use:   "version",
	Short: "Prints app version",
	Long:  "Prints app version",
	Run: func(cmd *cobra.Command, args []string) {
		actions.PrintVersion()
	},
}

var apply = &cobra.Command{
	Use:   "apply",
	Short: "Applies configs to the system",
	Long:  "Applies configs to the system",
	Run: func(cmd *cobra.Command, args []string) {
		// actions.CheckForFileEdits()
		actions.Apply()
	},
}

var revert = &cobra.Command{
	Use:   "revert",
	Short: "reverts back to original state",
	Long:  "reverts back to original state",
	Run: func(cmd *cobra.Command, args []string) {
		actions.CheckForFileEdits()
		actions.Revert()
	},
}

var restore_list bool

var restore = &cobra.Command{
	Use:   "restore",
	Short: "restores to a previous state",
	Long:  "restores to a previous state",
	Run: func(cmd *cobra.Command, args []string) {
		if restore_list {
			actions.PrintRestorePoints()
		} else {
			if len(args) == 0 {
				fmt.Println("No restore index param provided!")
				os.Exit(1)
			}

			// index, _ := strconv.ParseInt(args[0], 10, 0)
			index, _ := strconv.Atoi(args[0])
			actions.Restore(index)
		}
	},
}

func initArgs() {
	// restore args
	// var list string
	restore.Flags().BoolVarP(&restore_list, "list", "l", false, "Displays list of restore points")

}

var rootCmd = &cobra.Command{
	Use:   "stash",
	Short: "Declaration system configuration manager",
	Long:  "Declaration system configuration manager",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Created by: Dom Di Nenno")
		actions.PrintVersion()
		fmt.Println("See --help for more info")
	},
}

func Execute() {
	// Print Logo
	figure.NewColorFigure("STASH", "cricket", "yellow", true).Print()
	fmt.Println("")

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
