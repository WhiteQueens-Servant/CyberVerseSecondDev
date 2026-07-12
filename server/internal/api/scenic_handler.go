package api

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/cyberverse/server/internal/scenic"
)

// ── Attractions ──

func (r *Router) handleListAttractions(w http.ResponseWriter, req *http.Request) {
	ctx := req.Context()
	list, err := r.scenicDB.ListAttractions(ctx)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, list)
}

func (r *Router) handleGetAttraction(w http.ResponseWriter, req *http.Request) {
	id := req.PathValue("id")
	ctx := req.Context()
	a, err := r.scenicDB.GetAttraction(ctx, id)
	if err != nil {
		writeJSON(w, http.StatusNotFound, ErrorResponse{Error: err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, a)
}

func (r *Router) handleCreateAttraction(w http.ResponseWriter, req *http.Request) {
	var form scenic.AttractionForm
	if err := json.NewDecoder(req.Body).Decode(&form); err != nil {
		writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: "invalid JSON: " + err.Error()})
		return
	}
	if form.Name == "" || form.Description == "" {
		writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: "name and description are required"})
		return
	}
	ctx := req.Context()
	a, err := r.scenicDB.CreateAttraction(ctx, form)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}
	writeJSON(w, http.StatusCreated, a)
}

func (r *Router) handleUpdateAttraction(w http.ResponseWriter, req *http.Request) {
	id := req.PathValue("id")
	var form scenic.AttractionForm
	if err := json.NewDecoder(req.Body).Decode(&form); err != nil {
		writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: "invalid JSON: " + err.Error()})
		return
	}
	if form.Name == "" || form.Description == "" {
		writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: "name and description are required"})
		return
	}
	ctx := req.Context()
	a, err := r.scenicDB.UpdateAttraction(ctx, id, form)
	if err != nil {
		writeJSON(w, http.StatusNotFound, ErrorResponse{Error: err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, a)
}

func (r *Router) handleDeleteAttraction(w http.ResponseWriter, req *http.Request) {
	id := req.PathValue("id")
	ctx := req.Context()
	err := r.scenicDB.DeleteAttraction(ctx, id)
	if err != nil {
		if errors.Is(err, scenic.ErrAttractionInUse) {
			writeJSON(w, http.StatusConflict, ErrorResponse{Error: err.Error()})
			return
		}
		writeJSON(w, http.StatusNotFound, ErrorResponse{Error: err.Error()})
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// ── Routes ──

func (r *Router) handleListRoutes(w http.ResponseWriter, req *http.Request) {
	ctx := req.Context()
	list, err := r.scenicDB.ListRoutes(ctx)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, list)
}

func (r *Router) handleGetRoute(w http.ResponseWriter, req *http.Request) {
	id := req.PathValue("id")
	ctx := req.Context()
	rt, err := r.scenicDB.GetRoute(ctx, id)
	if err != nil {
		writeJSON(w, http.StatusNotFound, ErrorResponse{Error: err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, rt)
}

func (r *Router) handleCreateRoute(w http.ResponseWriter, req *http.Request) {
	var form scenic.RouteForm
	if err := json.NewDecoder(req.Body).Decode(&form); err != nil {
		writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: "invalid JSON: " + err.Error()})
		return
	}
	if form.Name == "" || form.Description == "" || form.Duration == "" {
		writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: "name, description, and duration are required"})
		return
	}
	if len(form.Steps) == 0 {
		writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: "at least one step is required"})
		return
	}
	ctx := req.Context()
	rt, err := r.scenicDB.CreateRoute(ctx, form)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}
	writeJSON(w, http.StatusCreated, rt)
}

func (r *Router) handleUpdateRoute(w http.ResponseWriter, req *http.Request) {
	id := req.PathValue("id")
	var form scenic.RouteForm
	if err := json.NewDecoder(req.Body).Decode(&form); err != nil {
		writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: "invalid JSON: " + err.Error()})
		return
	}
	if form.Name == "" || form.Description == "" || form.Duration == "" {
		writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: "name, description, and duration are required"})
		return
	}
	ctx := req.Context()
	rt, err := r.scenicDB.UpdateRoute(ctx, id, form)
	if err != nil {
		writeJSON(w, http.StatusNotFound, ErrorResponse{Error: err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, rt)
}

func (r *Router) handleDeleteRoute(w http.ResponseWriter, req *http.Request) {
	id := req.PathValue("id")

	// Check character references (characters are in file-based store, not cyberverse.db)
	chars := r.charStore.List()
	for _, c := range chars {
		if c.RecommendedRoutes != nil {
			for _, rid := range c.RecommendedRoutes {
				if rid == id {
					writeJSON(w, http.StatusConflict, ErrorResponse{
						Error: "route is referenced by character: " + c.Name,
					})
					return
				}
			}
		}
	}

	ctx := req.Context()
	err := r.scenicDB.DeleteRoute(ctx, id)
	if err != nil {
		writeJSON(w, http.StatusNotFound, ErrorResponse{Error: err.Error()})
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
