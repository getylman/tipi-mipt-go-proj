package handler

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/cloud-pricer/ingestion/internal/client"
	"github.com/cloud-pricer/ingestion/internal/invalid"
	"github.com/cloud-pricer/ingestion/internal/validator"
	"github.com/cloud-pricer/shared/apierror"
	"github.com/cloud-pricer/shared/types"
)

type Handler struct {
	log       *slog.Logger
	validator *validator.Validator
	pricing   client.PricingAPI
	invalid   invalid.Store
}

func New(
	log *slog.Logger,
	v *validator.Validator,
	p client.PricingAPI,
	inv invalid.Store,
) *Handler {
	return &Handler{log: log, validator: v, pricing: p, invalid: inv}
}

func (h *Handler) Usage(w http.ResponseWriter, r *http.Request) {
	var raw json.RawMessage
	if err := json.NewDecoder(r.Body).Decode(&raw); err != nil {
		if h.log != nil {
			h.log.Warn("usage: decode failed", "error", err)
		}
		apierror.ValidationError(w, "invalid JSON body", nil)
		return
	}

	var req types.UsageRequest
	if err := json.Unmarshal(raw, &req); err != nil {
		h.saveInvalid(raw, "failed to parse usage request: "+err.Error())
		apierror.ValidationError(w, "invalid request format", nil)
		return
	}

	if err := h.validator.ValidateUsage(&req); err != nil {
		h.saveInvalid(raw, err.Error())
		apierror.ValidationError(w, err.Error(), nil)
		return
	}

	if h.log != nil {
		h.log.Info("usage: forwarding to pricing engine", "user_id", req.UserID)
	}

	resp, err := h.pricing.Usage(r.Context(), &req)
	if err != nil {
		if h.log != nil {
			h.log.Error("usage: pricing engine error", "error", err)
		}
		apierror.UpstreamError(w, "pricing engine unavailable")
		return
	}

	if h.log != nil {
		h.log.Info("usage: done", "user_id", req.UserID, "total", resp.TotalPrice)
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *Handler) Estimate(w http.ResponseWriter, r *http.Request) {
	var raw json.RawMessage
	if err := json.NewDecoder(r.Body).Decode(&raw); err != nil {
		apierror.ValidationError(w, "invalid JSON body", nil)
		return
	}

	var req types.EstimateRequest
	if err := json.Unmarshal(raw, &req); err != nil {
		h.saveInvalid(raw, "failed to parse estimate request: "+err.Error())
		apierror.ValidationError(w, "invalid request format", nil)
		return
	}

	if err := h.validator.ValidateEstimate(&req); err != nil {
		h.saveInvalid(raw, err.Error())
		apierror.ValidationError(w, err.Error(), nil)
		return
	}

	resp, err := h.pricing.Estimate(r.Context(), &req)
	if err != nil {
		if h.log != nil {
			h.log.Error("estimate: pricing engine error", "error", err)
		}
		apierror.UpstreamError(w, "pricing engine unavailable")
		return
	}

	if h.log != nil {
		h.log.Info("estimate: calculated", "total", resp.TotalPrice)
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *Handler) Health(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func (h *Handler) saveInvalid(payload interface{}, reason string) {
	if err := h.invalid.Save(payload, reason); err != nil {
		if h.log != nil {
			h.log.Error("failed to save invalid metric", "error", err)
		}
	}
}
