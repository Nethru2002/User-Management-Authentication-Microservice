package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"auth-service/internal/delivery/http/middleware"
	"auth-service/internal/domain"
	"auth-service/internal/usecase"

	"github.com/google/uuid"
)

type SCIMHandler struct {
	scimUC usecase.SCIMUseCase
}

func NewSCIMHandler(scimUC usecase.SCIMUseCase) *SCIMHandler {
	return &SCIMHandler{scimUC: scimUC}
}

func (h *SCIMHandler) ServiceProviderConfig(w http.ResponseWriter, r *http.Request) {
	cfg := domain.SCIMServiceProviderConfig{
		Schemas: []string{"urn:ietf:params:scim:schemas:core:2.0:ServiceProviderConfig"},
	}
	cfg.AuthenticationSchemes = []struct {
		Name        string `json:"name"`
		Description string `json:"description"`
		Type        string `json:"type"`
	}{
		{Name: "OAuth Bearer Token", Description: "Standard OAuth Bearer Token", Type: "oauthbearertoken"},
	}

	w.Header().Set("Content-Type", "application/scim+json; charset=utf-8")
	_ = json.NewEncoder(w).Encode(cfg)
}

func (h *SCIMHandler) ListUsers(w http.ResponseWriter, r *http.Request) {
	tenantID, _ := r.Context().Value(middleware.TenantContextKey).(uuid.UUID)

	page, _ := strconv.Atoi(r.URL.Query().Get("startIndex"))
	limit, _ := strconv.Atoi(r.URL.Query().Get("count"))
	if limit < 1 {
		limit = 10
	}
	if page < 1 {
		page = 1
	}
	pq := domain.PaginationQuery{Page: page, Limit: limit}

	res, err := h.scimUC.ListUsers(r.Context(), tenantID, pq)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/scim+json; charset=utf-8")
	_ = json.NewEncoder(w).Encode(res)
}

func (h *SCIMHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
	tenantID, _ := r.Context().Value(middleware.TenantContextKey).(uuid.UUID)

	var req domain.SCIMUser
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	res, err := h.scimUC.CreateUser(r.Context(), tenantID, &req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/scim+json; charset=utf-8")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(res)
}

func (h *SCIMHandler) GetUserByID(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, "invalid user id format", http.StatusBadRequest)
		return
	}

	u, err := h.scimUC.GetUser(r.Context(), id)
	if err != nil {
		http.Error(w, "scim user not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/scim+json; charset=utf-8")
	_ = json.NewEncoder(w).Encode(u)
}

func (h *SCIMHandler) UpdateUserByID(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, "invalid user id format", http.StatusBadRequest)
		return
	}

	var req domain.SCIMUser
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	u, err := h.scimUC.UpdateUser(r.Context(), id, &req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/scim+json; charset=utf-8")
	_ = json.NewEncoder(w).Encode(u)
}

func (h *SCIMHandler) DeleteUserByID(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, "invalid user id format", http.StatusBadRequest)
		return
	}

	if err := h.scimUC.DeleteUser(r.Context(), id); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}