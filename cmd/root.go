package cmd

import (
	"os"

	"github.com/emiliopalmerini/grimoire/cmd/conjure"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "grimoire",
	Short: "A spellbook of developer incantations",
	Long:  `Grimoire is a CLI spellbook containing cantrips and spells for scaffolding, automation, and productivity.`,
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.AddCommand(conjure.Cmd)
}
