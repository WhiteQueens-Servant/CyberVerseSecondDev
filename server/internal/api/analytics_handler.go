package api

import (
	"encoding/json"
	"net/http"

	"github.com/cyberverse/server/internal/scenic"
)

// ── Dashboard ──

func (r *Router) handleGetDashboard(w http.ResponseWriter, req *http.Request) {
	ctx := req.Context()
	data, err := r.scenicDB.GetDashboard(ctx)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, data)
}

// ── Sessions (Analytics view) ──

func (r *Router) handleListAnalyticsSessions(w http.ResponseWriter, req *http.Request) {
	q := req.URL.Query()
	params := scenic.ListSessionsParams{
		CharacterID: q.Get("character_id"),
		DateFrom:    q.Get("date_from"),
		DateTo:      q.Get("date_to"),
		Sentiment:   q.Get("sentiment"),
	}
	ctx := req.Context()
	list, err := r.scenicDB.ListSessions(ctx, params)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, list)
}

func (r *Router) handleGetAnalyticsSessionDetail(w http.ResponseWriter, req *http.Request) {
	id := req.PathValue("id")
	ctx := req.Context()
	detail, err := r.scenicDB.GetSessionDetail(ctx, id)
	if err != nil {
		writeJSON(w, http.StatusNotFound, ErrorResponse{Error: err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, detail)
}

// ── Reports ──

func (r *Router) handleListReports(w http.ResponseWriter, req *http.Request) {
	ctx := req.Context()
	reports, err := r.scenicDB.ListReports(ctx)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, reports)
}

func (r *Router) handleGetReport(w http.ResponseWriter, req *http.Request) {
	id := req.PathValue("id")
	ctx := req.Context()
	report, err := r.scenicDB.GetReport(ctx, id)
	if err != nil {
		writeJSON(w, http.StatusNotFound, ErrorResponse{Error: err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, report)
}

func (r *Router) handleGenerateReport(w http.ResponseWriter, req *http.Request) {
	var form struct {
		DateFrom    string `json:"date_from"`
		DateTo      string `json:"date_to"`
		CharacterID string `json:"character_id"`
	}
	if err := json.NewDecoder(req.Body).Decode(&form); err != nil {
		// Allow empty body
		form = struct {
			DateFrom    string `json:"date_from"`
			DateTo      string `json:"date_to"`
			CharacterID string `json:"character_id"`
		}{}
	}
	ctx := req.Context()
	report, err := r.scenicDB.CreateReport(ctx, form.DateFrom, form.DateTo, form.CharacterID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}
	// TODO: trigger SubAgent task for LLM analysis (T-24 integration)
	writeJSON(w, http.StatusCreated, report)
}

func (r *Router) handleDeleteReport(w http.ResponseWriter, req *http.Request) {
	id := req.PathValue("id")
	ctx := req.Context()
	err := r.scenicDB.DeleteReport(ctx, id)
	if err != nil {
		writeJSON(w, http.StatusNotFound, ErrorResponse{Error: err.Error()})
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
