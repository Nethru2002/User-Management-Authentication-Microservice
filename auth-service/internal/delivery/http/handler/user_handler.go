package handler

import (
	"errors"
	"net/http"
	"strconv"

	"auth-service/internal/delivery/http/middleware"
	"auth-service/internal/domain"
	"auth-service/internal/usecase"
	"auth-service/pkg/utils"
)

type UserHandler struct {
	userUC usecase.UserUseCase
}

func NewUserHandler(userUC usecase.UserUseCase) *UserHandler {
	return &UserHandler{userUC: userUC}
}

func (h *UserHandler) GetProfile(w http.ResponseWriter, r *http.Request) {
	claims, ok := r.Context().Value(middleware.UserContextKey).(*domain.JwtCustomClaims)
	if !ok || claims == nil {
		utils.WriteError(w, http.StatusUnauthorized, "unauthorized: missing or invalid context claims")
		return
	}

	user, err := h.userUC.GetProfile(r.Context(), claims.UserID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			utils.WriteError(w, http.StatusNotFound, "user profile not found")
			return
		}
		utils.WriteError(w, http.StatusInternalServerError, "failed to retrieve profile")
		return
	}

	utils.WriteJSON(w, http.StatusOK, user, "profile fetched successfully")
}

func (h *UserHandler) ListUsers(w http.ResponseWriter, r *http.Request) {
	pageStr := r.URL.Query().Get("page")
	limitStr := r.URL.Query().Get("limit")

	page := 1
	limit := 10

	if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
		page = p
	}
	if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
		if l > 100 {
			limit = 100
		} else {
			limit = l
		}
	}

	pq := domain.PaginationQuery{
		Page:  page,
		Limit: limit,
	}

	res, err := h.userUC.ListUsers(r.Context(), pq)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, "failed to retrieve user list")
		return
	}

	utils.WriteJSON(w, http.StatusOK, res, "users list fetched successfully")
}