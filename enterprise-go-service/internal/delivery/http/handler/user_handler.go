package handler

import (
	"net/http"
	"strconv"

	"enterprise-go-service/internal/delivery/http/middleware"
	"enterprise-go-service/internal/domain"
	"enterprise-go-service/internal/usecase"
	"enterprise-go-service/pkg/utils"
)

type UserHandler struct {
	userUC usecase.UserUseCase
}

func NewUserHandler(userUC usecase.UserUseCase) *UserHandler {
	return &UserHandler{userUC: userUC}
}

func (h *UserHandler) GetProfile(w http.ResponseWriter, r *http.Request) {
	claims, ok := r.Context().Value(middleware.UserContextKey).(*domain.JwtCustomClaims)
	if !ok {
		utils.WriteError(w, http.StatusUnauthorized, "unauthorized context missing")
		return
	}

	user, err := h.userUC.GetProfile(r.Context(), claims.UserID)
	if err != nil {
		utils.WriteError(w, http.StatusNotFound, err.Error())
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
		limit = l
	}

	pq := domain.PaginationQuery{
		Page:  page,
		Limit: limit,
	}

	res, err := h.userUC.ListUsers(r.Context(), pq)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}

	utils.WriteJSON(w, http.StatusOK, res, "users list fetched successfully")
}