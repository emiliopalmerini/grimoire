package cmd

import (
	"os"

	"github.com/emiliopalmerini/grimorio/cmd/augur"
	"github.com/emiliopalmerini/grimorio/cmd/conjure"
	"github.com/emiliopalmerini/grimorio/cmd/divine"
	"github.com/emiliopalmerini/grimorio/cmd/mend"
	modifymemory "github.com/emiliopalmerini/grimorio/cmd/modify-memory"
	"github.com/emiliopalmerini/grimorio/cmd/scry"
	"github.com/emiliopalmerini/grimorio/cmd/summon"
	"github.com/emiliopalmerini/grimorio/cmd/transmute"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "grimorio",
	Short: "A spellbook of developer incantations",
	Long:  `Grimorio is a CLI spellbook containing cantrips and spells for scaffolding, automation, and productivity.`,
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.AddCommand(augur.Cmd)
	rootCmd.AddCommand(conjure.Cmd)
	rootCmd.AddCommand(divine.Cmd)
	rootCmd.AddCommand(mend.Cmd)
	rootCmd.AddCommand(modifymemory.Cmd)
	rootCmd.AddCommand(scry.Cmd)
	rootCmd.AddCommand(summon.Cmd)
	rootCmd.AddCommand(transmute.Cmd)
}
