package stats

import (
	"context"
	"fmt"
	"time"

	"github.com/emiliopalmerini/grimorio/internal/metrics"
	"github.com/emiliopalmerini/grimorio/internal/metrics/turso"
	"github.com/spf13/cobra"
)

var (
	days   int
	remote bool
)

var Cmd = &cobra.Command{
	Use:   "stats",
	Short: "[Cantrip] View usage statistics",
	Long: `Stats displays usage statistics and insights for grimorio commands.

Examples:
  grimorio stats              # All time (default)
  grimorio stats --days 7     # Last 7 days
  grimorio stats --days 30    # Last 30 days
  grimorio stats --remote     # Sync and show aggregated stats from all machines`,
	RunE: runStats,
}

func init() {
	Cmd.Flags().IntVarP(&days, "days", "d", 0, "Number of days to show stats for (0 = all time)")
	Cmd.Flags().BoolVarP(&remote, "remote", "r", false, "Sync and show aggregated stats from Turso")
}

func runStats(cmd *cobra.Command, args []string) error {
	if remote {
		return runRemoteStats(cmd, args)
	}
	return runLocalStats(cmd, args)
}

func runLocalStats(cmd *cobra.Command, args []string) error {
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

func runRemoteStats(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	config := turso.ConfigFromEnv()
	if config == nil {
		return fmt.Errorf("TURSO_DATABASE_URL and TURSO_AUTH_TOKEN environment variables are required for remote stats")
	}

	client := turso.NewClient(config)

	queries, err := metrics.Default.Queries(ctx)
	if err != nil {
		return fmt.Errorf("failed to get database queries: %w", err)
	}

	syncer := turso.NewSyncer(client, queries)

	fmt.Println("Syncing to Turso...")
	result, err := syncer.Sync(ctx)
	if err != nil {
		return fmt.Errorf("failed to sync: %w", err)
	}
	if result.CommandsSynced > 0 || result.AISynced > 0 {
		fmt.Printf("Synced %d commands, %d AI invocations\n", result.CommandsSynced, result.AISynced)
	}
	fmt.Println()

	var from time.Time
	var periodLabel string
	if days > 0 {
		from = time.Now().AddDate(0, 0, -days)
		periodLabel = fmt.Sprintf("last %d days", days)
	} else {
		periodLabel = "all time"
	}

	summary, err := syncer.GetRemoteSummary(ctx, from, time.Time{})
	if err != nil {
		return fmt.Errorf("failed to get remote stats: %w", err)
	}

	fmt.Printf("Usage Statistics - All Machines (%s)\n", periodLabel)
	fmt.Println("==========================================")
	fmt.Println()

	fmt.Printf("Total commands:    %d\n", summary.TotalCommands)
	fmt.Printf("Failed commands:   %d\n", summary.TotalFailures)
	if summary.TotalCommands > 0 {
		successRate := float64(summary.TotalCommands-summary.TotalFailures) / float64(summary.TotalCommands) * 100
		fmt.Printf("Success rate:      %.1f%%\n", successRate)
	}
	fmt.Println()

	if len(summary.MachineStats) > 0 {
		fmt.Println("By Machine")
		fmt.Println("----------")
		for _, ms := range summary.MachineStats {
			fmt.Printf("%-16s %3d commands\n", ms.MachineID, ms.Count)
		}
		fmt.Println()
	}

	if summary.AIStats.TotalCalls > 0 {
		fmt.Println("AI Usage")
		fmt.Println("--------")
		fmt.Printf("API calls:         %d\n", summary.AIStats.TotalCalls)
		fmt.Printf("Prompt tokens:     %d\n", summary.AIStats.TotalPromptTokens)
		fmt.Printf("Response tokens:   %d\n", summary.AIStats.TotalResponseTokens)
		fmt.Printf("Avg latency:       %.0fms\n", summary.AIStats.AvgLatencyMs)
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
