package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var to string

var greetCmd = &cobra.Command{
	Use:   "greet",
	Short: "Prints a hello message",
	Long:  "This ",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("Hello: from %s to %s!\n", name, to)
	},
}

func init() {
	greetCmd.Flags().StringVarP(&to, "to", "t", "", "Greet to sb")
	rootCmd.AddCommand(greetCmd)
}
