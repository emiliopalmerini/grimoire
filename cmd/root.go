package cmd

import (
	"os"

	"github.com/emiliopalmerini/grimorio/cmd/augury"
	"github.com/emiliopalmerini/grimorio/cmd/conjure"
	"github.com/emiliopalmerini/grimorio/cmd/dashboard"
	"github.com/emiliopalmerini/grimorio/cmd/identify"
	"github.com/emiliopalmerini/grimorio/cmd/mending"
	modifymemory "github.com/emiliopalmerini/grimorio/cmd/modify-memory"
	"github.com/emiliopalmerini/grimorio/cmd/polymorph"
	"github.com/emiliopalmerini/grimorio/cmd/prepare"
	"github.com/emiliopalmerini/grimorio/cmd/scrying"
	"github.com/emiliopalmerini/grimorio/cmd/sending"
	"github.com/emiliopalmerini/grimorio/cmd/stats"
	"github.com/emiliopalmerini/grimorio/cmd/summon"
	"github.com/emiliopalmerini/grimorio/internal/metrics"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "grimorio",
	Short: "A spellbook of developer incantations",
	Long:  `Grimorio is a CLI spellbook containing cantrips and spells for scaffolding, automation, and productivity.`,
}

func Execute() {
	dbPath, err := metrics.DefaultDBPath()
	if err == nil {
		tracker := metrics.NewSQLiteTracker(dbPath)
		metrics.Default = tracker
		defer tracker.Close()
	}

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.AddCommand(augury.Cmd)
	rootCmd.AddCommand(conjure.Cmd)
	rootCmd.AddCommand(dashboard.Cmd)
	rootCmd.AddCommand(identify.Cmd)
	rootCmd.AddCommand(mending.Cmd)
	rootCmd.AddCommand(modifymemory.Cmd)
	rootCmd.AddCommand(polymorph.Cmd)
	rootCmd.AddCommand(prepare.Cmd)
	rootCmd.AddCommand(scrying.Cmd)
	rootCmd.AddCommand(sending.Cmd)
	rootCmd.AddCommand(stats.Cmd)
	rootCmd.AddCommand(summon.Cmd)
}
