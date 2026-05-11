package handler

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/cloud-pricer/pricing/internal/estimate"
	"github.com/cloud-pricer/pricing/internal/usage"
	"github.com/cloud-pricer/shared/apierror"
	"github.com/cloud-pricer/shared/types"
)

type Handler struct {
	log      *slog.Logger
	usage    *usage.Service
	estimate *estimate.Service
	products ProductLister
}

func New(
	log *slog.Logger,
	usageSvc *usage.Service,
	estimateSvc *estimate.Service,
	products ProductLister,
) *Handler {
	return &Handler{log: log, usage: usageSvc, estimate: estimateSvc, products: products}
}

func (h *Handler) Usage(w http.ResponseWriter, r *http.Request) {
	var req types.UsageRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.log.Warn("usage: decode failed", "error", err)
		apierror.ValidationError(w, "invalid JSON body", nil)
		return
	}

	log := h.log.With("user_id", req.UserID, "items_count", len(req.Items))
	log.Info("usage request received")

	resp, err := h.usage.Process(r.Context(), &req)
	if err != nil {
		log.Error("usage processing failed", "error", err)
		apierror.InternalError(w, err.Error())
		return
	}

	log.Info("usage saved", "total_price", resp.TotalPrice)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *Handler) Estimate(w http.ResponseWriter, r *http.Request) {
	var req types.EstimateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.log.Warn("estimate: decode failed", "error", err)
		apierror.ValidationError(w, "invalid JSON body", nil)
		return
	}

	h.log.Info("estimate request received", "items_count", len(req.Items))

	resp, err := h.estimate.Calculate(r.Context(), &req)
	if err != nil {
		h.log.Error("estimate failed", "error", err)
		apierror.NotFoundError(w, err.Error(), nil)
		return
	}

	h.log.Info("estimate calculated", "total_price", resp.TotalPrice)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *Handler) ListProducts(w http.ResponseWriter, r *http.Request) {
	products, err := h.products.ListAll(r.Context())
	if err != nil {
		h.log.Error("list products failed", "error", err)
		apierror.InternalError(w, "failed to list products")
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{"count": len(products), "products": products})
}

func (h *Handler) Health(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}
