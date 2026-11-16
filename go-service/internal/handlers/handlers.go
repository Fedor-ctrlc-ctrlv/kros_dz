package handlers

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"

	"loglab/go-service/internal/middleware"
	"loglab/go-service/internal/store"
)

type API struct {
	Log    *slog.Logger
	Orders *store.Orders
}

type createOrder struct {
	Product  string `json:"product"`
	Quantity int    `json:"quantity"`
}

func jsonResponse(w http.ResponseWriter, code int, body any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(body)
}
func (a *API) logger(r *http.Request) *slog.Logger {
	return a.Log.With("request_id", middleware.RequestID(r.Context()))
}
func (a *API) Routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", a.health)
	mux.HandleFunc("POST /orders", a.create)
	mux.HandleFunc("GET /orders/{id}", a.get)
	mux.HandleFunc("GET /simulate-error", a.simulateError)
	mux.HandleFunc("GET /simulate-panic", a.simulatePanic)
	return mux
}
func (a *API) health(w http.ResponseWriter, r *http.Request) {
	a.logger(r).Debug("health check")
	jsonResponse(w, 200, map[string]string{"status": "ok", "service": "go-api"})
}
func (a *API) create(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	var input createOrder
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		a.logger(r).Warn("invalid request JSON", "error", err.Error())
		jsonResponse(w, 400, map[string]string{"error": "invalid JSON"})
		return
	}
	input.Product = strings.TrimSpace(input.Product)
	if input.Product == "" || input.Quantity <= 0 || input.Quantity > 1000 {
		a.logger(r).Warn("order validation failed", "product", input.Product, "quantity", input.Quantity)
		jsonResponse(w, 400, map[string]string{"error": "product must be nonempty and quantity must be 1..1000"})
		return
	}
	order := a.Orders.Create(input.Product, input.Quantity)
	a.logger(r).Info("order created", "order_id", order.ID, "product", order.Product, "quantity", order.Quantity)
	jsonResponse(w, 201, order)
}
func (a *API) get(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || id <= 0 {
		a.logger(r).Warn("invalid order ID", "id", r.PathValue("id"))
		jsonResponse(w, 400, map[string]string{"error": "invalid order ID"})
		return
	}
	order, err := a.Orders.Get(id)
	if err != nil {
		a.logger(r).Warn("order not found", "order_id", id)
		jsonResponse(w, 404, map[string]string{"error": "order not found"})
		return
	}
	a.logger(r).Debug("order retrieved", "order_id", id)
	jsonResponse(w, 200, order)
}
func (a *API) simulateError(w http.ResponseWriter, r *http.Request) {
	a.logger(r).Error("simulated database timeout", "component", "database", "retryable", true)
	jsonResponse(w, 500, map[string]string{"error": "simulated database timeout"})
}
func (a *API) simulatePanic(w http.ResponseWriter, r *http.Request) {
	a.logger(r).Debug("starting simulated panic", "at", time.Now().Format(time.RFC3339))
	panic("simulated unexpected failure")
}
