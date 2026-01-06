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
  grimorio stats              # All time (default)
  grimorio stats --days 7     # Last 7 days
  grimorio stats --days 30    # Last 30 days`,
	RunE: runStats,
}

func init() {
	Cmd.Flags().IntVarP(&days, "days", "d", 0, "Number of days to show stats for (0 = all time)")
}

func runStats(cmd *cobra.Command, args []string) error {
	var filter metrics.Filter
	var periodLabel string

	if days > 0 {
		filter.From = time.Now().AddDate(0, 0, -days)
		periodLabel = fmt.Sprintf("last %d days", days)
	} else {
		filter.From = time.Time{}
		periodLabel = "all time"
	}

	summary, err := metrics.Default.GetSummary(context.Background(), filter)
	if err != nil {
		return fmt.Errorf("failed to get stats: %w", err)
	}

	fmt.Printf("Usage Statistics (%s)\n", periodLabel)
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
