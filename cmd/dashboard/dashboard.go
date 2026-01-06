package dashboard

import (
	"fmt"

	"github.com/emiliopalmerini/grimorio/internal/dashboard"
	"github.com/emiliopalmerini/grimorio/internal/metrics"
	"github.com/spf13/cobra"
)

var port int

var Cmd = &cobra.Command{
	Use:   "dashboard",
	Short: "[Cantrip] Start the web dashboard",
	Long: `Dashboard starts a web server to visualize usage statistics.

Examples:
  grimorio dashboard           # Start on default port 8080
  grimorio dashboard --port 3000`,
	RunE: runDashboard,
}

func init() {
	Cmd.Flags().IntVarP(&port, "port", "p", 8080, "Port to run the dashboard on")
}

func runDashboard(cmd *cobra.Command, args []string) error {
	tracker, ok := metrics.Default.(*metrics.SQLiteTracker)
	if !ok {
		return fmt.Errorf("metrics tracking is not enabled")
	}

	srv := dashboard.NewServer(tracker, port)
	return srv.Start()
}
