package handler

import (
	"context"
	"net/http"
	"time"

	"auth-service/pkg/utils"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

type HealthHandler struct {
	db  *pgxpool.Pool
	rdb *redis.Client
}

func NewHealthHandler(db *pgxpool.Pool, rdb *redis.Client) *HealthHandler {
	return &HealthHandler{db: db, rdb: rdb}
}

func (h *HealthHandler) Healthz(w http.ResponseWriter, r *http.Request) {
	utils.WriteJSON(w, http.StatusOK, map[string]string{"status": "UP"}, "service is alive")
}

func (h *HealthHandler) Readyz(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()

	if err := h.db.Ping(ctx); err != nil {
		utils.WriteError(w, http.StatusServiceUnavailable, "database connection pool unreachable")
		return
	}

	if err := h.rdb.Ping(ctx).Err(); err != nil {
		utils.WriteError(w, http.StatusServiceUnavailable, "redis cache unreachable")
		return
	}

	utils.WriteJSON(w, http.StatusOK, map[string]string{"status": "READY"}, "all dependencies operational")
}