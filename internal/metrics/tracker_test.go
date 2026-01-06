package metrics

import (
	"context"
	"testing"
	"time"
)

func TestSQLiteTracker(t *testing.T) {
	tracker := NewSQLiteTracker(":memory:")
	defer tracker.Close()

	ctx := context.Background()

	t.Run("record and query commands", func(t *testing.T) {
		err := tracker.RecordCommand(ctx, "conjure", Cantrip, 100, 0, `{"transport":"http"}`)
		if err != nil {
			t.Fatalf("RecordCommand: %v", err)
		}

		err = tracker.RecordCommand(ctx, "divine", Spell, 5000, 0, "")
		if err != nil {
			t.Fatalf("RecordCommand: %v", err)
		}

		err = tracker.RecordCommand(ctx, "mend", Cantrip, 50, 1, "")
		if err != nil {
			t.Fatalf("RecordCommand (failure): %v", err)
		}

		summary, err := tracker.GetSummary(ctx, Filter{From: time.Now().Add(-time.Hour)})
		if err != nil {
			t.Fatalf("GetSummary: %v", err)
		}

		if summary.TotalCommands != 3 {
			t.Errorf("TotalCommands = %d, want 3", summary.TotalCommands)
		}
		if summary.TotalFailures != 1 {
			t.Errorf("TotalFailures = %d, want 1", summary.TotalFailures)
		}
		if len(summary.CommandStats) != 3 {
			t.Errorf("CommandStats count = %d, want 3", len(summary.CommandStats))
		}
	})

	t.Run("record ai invocations", func(t *testing.T) {
		err := tracker.RecordAI(ctx, "divine", "opus", 1000, 500, 2000, true, "")
		if err != nil {
			t.Fatalf("RecordAI: %v", err)
		}

		err = tracker.RecordAI(ctx, "scry", "opus", 800, 0, 100, false, "timeout")
		if err != nil {
			t.Fatalf("RecordAI (failure): %v", err)
		}

		summary, err := tracker.GetSummary(ctx, Filter{From: time.Now().Add(-time.Hour)})
		if err != nil {
			t.Fatalf("GetSummary: %v", err)
		}

		if summary.TotalAICalls != 2 {
			t.Errorf("TotalAICalls = %d, want 2", summary.TotalAICalls)
		}
	})
}

func TestNoopTracker(t *testing.T) {
	tracker := NoopTracker{}
	ctx := context.Background()

	if err := tracker.RecordCommand(ctx, "test", Cantrip, 100, 0, ""); err != nil {
		t.Errorf("RecordCommand: %v", err)
	}

	if err := tracker.RecordAI(ctx, "test", "sonnet", 100, 100, 100, true, ""); err != nil {
		t.Errorf("RecordAI: %v", err)
	}

	summary, err := tracker.GetSummary(ctx, Filter{})
	if err != nil {
		t.Errorf("GetSummary: %v", err)
	}
	if summary.TotalCommands != 0 {
		t.Errorf("TotalCommands = %d, want 0", summary.TotalCommands)
	}

	if err := tracker.Close(); err != nil {
		t.Errorf("Close: %v", err)
	}
}
