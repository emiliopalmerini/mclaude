package web

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/emiliopalmerini/mclaude/internal/domain"
	"github.com/emiliopalmerini/mclaude/internal/web/templates"
)

func (s *Server) handleSettings(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get pricing via port interface
	pricing, _ := s.pricingRepo.List(ctx)

	models := make([]templates.ModelPricing, 0, len(pricing))
	for _, p := range pricing {
		model := templates.ModelPricing{
			ID:              p.ID,
			DisplayName:     p.DisplayName,
			InputPerMillion: p.InputPerMillion,
			OutputPerMillion: p.OutputPerMillion,
			IsDefault:       p.IsDefault,
		}
		if p.CacheReadPerMillion != nil {
			model.CacheReadPerMillion = *p.CacheReadPerMillion
		}
		if p.CacheWritePerMillion != nil {
			model.CacheWritePerMillion = *p.CacheWritePerMillion
		}
		models = append(models, model)
	}

	// Get plan config
	var planView *templates.PlanConfigView
	if planConfig, err := s.planConfigRepo.Get(ctx); err == nil && planConfig != nil {
		planView = &templates.PlanConfigView{
			PlanType:                planConfig.PlanType,
			WindowHours:             planConfig.WindowHours,
			LearnedTokenLimit:       planConfig.LearnedTokenLimit,
			WeeklyLearnedTokenLimit: planConfig.WeeklyLearnedTokenLimit,
		}
	}

	pageData := templates.SettingsPageData{
		Pricing:    models,
		PlanConfig: planView,
	}

	templates.SettingsPage(pageData).Render(ctx, w)
}

func (s *Server) handleAPICreatePricing(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	if err := r.ParseForm(); err != nil {
		http.Error(w, "Invalid form data", http.StatusBadRequest)
		return
	}

	modelID := strings.TrimSpace(r.FormValue("model_id"))
	if modelID == "" {
		http.Error(w, "Model ID is required", http.StatusBadRequest)
		return
	}

	input, err := strconv.ParseFloat(r.FormValue("input"), 64)
	if err != nil || input <= 0 {
		http.Error(w, "Valid input price is required", http.StatusBadRequest)
		return
	}
	output, err := strconv.ParseFloat(r.FormValue("output"), 64)
	if err != nil || output <= 0 {
		http.Error(w, "Valid output price is required", http.StatusBadRequest)
		return
	}

	displayName := strings.TrimSpace(r.FormValue("display_name"))
	if displayName == "" {
		displayName = modelID
	}

	pricing := &domain.ModelPricing{
		ID:               modelID,
		DisplayName:      displayName,
		InputPerMillion:  input,
		OutputPerMillion: output,
		CreatedAt:        time.Now().UTC(),
	}

	if v, err := strconv.ParseFloat(r.FormValue("cache_read"), 64); err == nil && v > 0 {
		pricing.CacheReadPerMillion = &v
	}
	if v, err := strconv.ParseFloat(r.FormValue("cache_write"), 64); err == nil && v > 0 {
		pricing.CacheWritePerMillion = &v
	}

	// Check if model exists (update) or create new
	existing, _ := s.pricingRepo.GetByID(ctx, modelID)
	if existing != nil {
		pricing.IsDefault = existing.IsDefault
		pricing.CreatedAt = existing.CreatedAt
		if err := s.pricingRepo.Update(ctx, pricing); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	} else {
		// If first model, set as default
		allPricing, _ := s.pricingRepo.List(ctx)
		if len(allPricing) == 0 {
			pricing.IsDefault = true
		}
		if err := s.pricingRepo.Create(ctx, pricing); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	w.Header().Set("HX-Redirect", "/settings")
	w.WriteHeader(http.StatusOK)
}

func (s *Server) handleAPISetDefaultPricing(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id := r.PathValue("id")

	if err := s.pricingRepo.SetDefault(ctx, id); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("HX-Redirect", "/settings")
	w.WriteHeader(http.StatusOK)
}

func (s *Server) handleAPIDeletePricing(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id := r.PathValue("id")

	if err := s.pricingRepo.Delete(ctx, id); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("HX-Redirect", "/settings")
	w.WriteHeader(http.StatusOK)
}

func (s *Server) handleAPISetPlan(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	if err := r.ParseForm(); err != nil {
		http.Error(w, "Invalid form data", http.StatusBadRequest)
		return
	}

	planType := strings.ToLower(strings.TrimSpace(r.FormValue("plan_type")))
	if _, ok := domain.PlanPresets[planType]; !ok {
		http.Error(w, "Invalid plan type (valid: pro, max_5x, max_20x)", http.StatusBadRequest)
		return
	}

	config := &domain.PlanConfig{
		PlanType:    planType,
		WindowHours: 5,
	}

	if err := s.planConfigRepo.Upsert(ctx, config); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("HX-Redirect", "/settings")
	w.WriteHeader(http.StatusOK)
}

func (s *Server) handleAPILearnLimit(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	config, err := s.planConfigRepo.Get(ctx)
	if err != nil || config == nil {
		http.Error(w, "No plan configured", http.StatusBadRequest)
		return
	}

	summary, err := s.planConfigRepo.GetRollingWindowSummary(ctx, config.WindowHours)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if summary.TotalTokens == 0 {
		http.Error(w, "No token usage recorded in the current window", http.StatusBadRequest)
		return
	}

	if err := s.planConfigRepo.UpdateLearnedLimit(ctx, summary.TotalTokens); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("HX-Redirect", "/settings")
	w.WriteHeader(http.StatusOK)
}

func (s *Server) handleAPILearnWeeklyLimit(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	config, err := s.planConfigRepo.Get(ctx)
	if err != nil || config == nil {
		http.Error(w, "No plan configured", http.StatusBadRequest)
		return
	}

	summary, err := s.planConfigRepo.GetWeeklyWindowSummary(ctx)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if summary.TotalTokens == 0 {
		http.Error(w, "No token usage recorded in the last 7 days", http.StatusBadRequest)
		return
	}

	if err := s.planConfigRepo.UpdateWeeklyLearnedLimit(ctx, summary.TotalTokens); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("HX-Redirect", "/settings")
	w.WriteHeader(http.StatusOK)
}
