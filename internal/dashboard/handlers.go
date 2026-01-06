package dashboard

import (
	"net/http"
	"time"

	"github.com/emiliopalmerini/grimorio/internal/dashboard/views"
	"github.com/emiliopalmerini/grimorio/internal/metrics"
	"github.com/emiliopalmerini/grimorio/internal/metrics/db"
)

func (s *Server) parseFilter(r *http.Request) metrics.Filter {
	filter := metrics.Filter{}

	if from := r.URL.Query().Get("from"); from != "" {
		if t, err := time.Parse("2006-01-02", from); err == nil {
			filter.From = t
		}
	}
	if to := r.URL.Query().Get("to"); to != "" {
		if t, err := time.Parse("2006-01-02", to); err == nil {
			filter.To = t.Add(24*time.Hour - time.Second)
		}
	}
	filter.Command = r.URL.Query().Get("command")

	return filter
}

func (s *Server) handleDashboard(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	filter := s.parseFilter(r)

	summary, err := s.tracker.GetSummary(ctx, filter)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	queries, err := s.tracker.Queries(ctx)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	fromStr := filter.From.Format("2006-01-02 15:04:05")
	toStr := "9999-12-31 23:59:59"
	if !filter.To.IsZero() {
		toStr = filter.To.Format("2006-01-02 15:04:05")
	}

	modelStats, _ := queries.GetAIStatsByModel(ctx, db.GetAIStatsByModelParams{
		FromDate: fromStr,
		ToDate:   toStr,
	})
	recentCmds, _ := queries.GetRecentCommands(ctx, db.GetRecentCommandsParams{
		CommandFilter: filter.Command,
		LimitCount:    10,
	})
	distinctCmds, _ := queries.GetDistinctCommands(ctx)

	data := views.DashboardData{
		Summary:          summary,
		ModelStats:       modelStats,
		RecentCommands:   recentCmds,
		DistinctCommands: distinctCmds,
		Filter:           filter,
	}

	views.Layout(views.Dashboard(data)).Render(ctx, w)
}

func (s *Server) handleStats(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	filter := s.parseFilter(r)

	summary, err := s.tracker.GetSummary(ctx, filter)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	views.Stats(summary).Render(ctx, w)
}

func (s *Server) handleCommands(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	filter := s.parseFilter(r)

	summary, err := s.tracker.GetSummary(ctx, filter)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	views.Commands(summary.CommandStats).Render(ctx, w)
}

func (s *Server) handleModels(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	filter := s.parseFilter(r)

	queries, err := s.tracker.Queries(ctx)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	fromStr := filter.From.Format("2006-01-02 15:04:05")
	toStr := "9999-12-31 23:59:59"
	if !filter.To.IsZero() {
		toStr = filter.To.Format("2006-01-02 15:04:05")
	}

	modelStats, err := queries.GetAIStatsByModel(ctx, db.GetAIStatsByModelParams{
		FromDate: fromStr,
		ToDate:   toStr,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	views.Models(modelStats).Render(ctx, w)
}

func (s *Server) handleHistory(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	filter := s.parseFilter(r)

	queries, err := s.tracker.Queries(ctx)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	recentCmds, err := queries.GetRecentCommands(ctx, db.GetRecentCommandsParams{
		CommandFilter: filter.Command,
		LimitCount:    10,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	views.History(recentCmds).Render(ctx, w)
}
