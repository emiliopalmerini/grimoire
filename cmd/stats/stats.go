package stats

import (
	"context"
	"fmt"
	"time"

	"github.com/emiliopalmerini/grimorio/internal/metrics"
	"github.com/spf13/cobra"
)

var days int

var Cmd = &cobra.Command{
	Use:   "stats",
	Short: "[Cantrip] View usage statistics",
	Long: `Stats displays usage statistics and insights for grimorio commands.

Examples:
  grimorio stats              # Last 7 days
  grimorio stats --days 30    # Last 30 days`,
	RunE: runStats,
}

func init() {
	Cmd.Flags().IntVarP(&days, "days", "d", 7, "Number of days to show stats for")
}

func runStats(cmd *cobra.Command, args []string) error {
	since := time.Now().AddDate(0, 0, -days)
	summary, err := metrics.Default.GetSummary(context.Background(), since)
	if err != nil {
		return fmt.Errorf("failed to get stats: %w", err)
	}

	fmt.Printf("Usage Statistics (last %d days)\n", days)
	fmt.Println("================================")
	fmt.Println()

	fmt.Printf("Total commands:    %d\n", summary.TotalCommands)
	fmt.Printf("Failed commands:   %d\n", summary.TotalFailures)
	if summary.TotalCommands > 0 {
		successRate := float64(summary.TotalCommands-summary.TotalFailures) / float64(summary.TotalCommands) * 100
		fmt.Printf("Success rate:      %.1f%%\n", successRate)
	}
	fmt.Println()

	if summary.TotalAICalls > 0 {
		fmt.Println("AI Usage")
		fmt.Println("--------")
		fmt.Printf("API calls:         %d\n", summary.TotalAICalls)
		fmt.Printf("Prompt tokens:     %d\n", summary.TotalPromptTokens)
		fmt.Printf("Response tokens:   %d\n", summary.TotalResponseTokens)
		fmt.Printf("Avg latency:       %.0fms\n", summary.AvgLatencyMs)
		fmt.Println()
	}

	if len(summary.CommandStats) > 0 {
		fmt.Println("Commands by Usage")
		fmt.Println("-----------------")
		for _, cs := range summary.CommandStats {
			fmt.Printf("%-15s %3d calls  (avg %.0fms)\n", cs.Command, cs.Count, cs.AvgDurationMs)
		}
	}

	return nil
}
