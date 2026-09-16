package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"auth-service/internal/domain"
	"auth-service/internal/usecase"
	"auth-service/pkg/utils"
)

type AuthHandler struct {
	authUC     usecase.AuthUseCase
	keyManager *utils.KeyManager
}

func NewAuthHandler(authUC usecase.AuthUseCase, keyManager *utils.KeyManager) *AuthHandler {
	return &AuthHandler{
		authUC:     authUC,
		keyManager: keyManager,
	}
}

type registerRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type refreshRequest struct {
	RefreshToken string `json:"refresh_token"`
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20) // 1MB payload bound
	var req registerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteError(w, http.StatusBadRequest, "invalid request payload: must be valid JSON under 1MB")
		return
	}

	err := h.authUC.Register(r.Context(), req.Email, req.Password, r.RemoteAddr)
	if err != nil {
		if errors.Is(err, domain.ErrConflict) {
			utils.WriteError(w, http.StatusConflict, "user with this email already exists")
			return
		}
		utils.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	utils.WriteJSON(w, http.StatusCreated, nil, "user registered successfully")
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	var req loginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteError(w, http.StatusBadRequest, "invalid request payload: must be valid JSON under 1MB")
		return
	}

	tokens, err := h.authUC.Login(r.Context(), req.Email, req.Password, r.RemoteAddr)
	if err != nil {
		if errors.Is(err, domain.ErrUnauthorized) {
			utils.WriteError(w, http.StatusUnauthorized, "invalid email or password")
			return
		}
		utils.WriteError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	utils.WriteJSON(w, http.StatusOK, tokens, "login successful")
}

func (h *AuthHandler) Refresh(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	var req refreshRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.RefreshToken == "" {
		utils.WriteError(w, http.StatusBadRequest, "invalid request payload: missing refresh_token")
		return
	}

	tokens, err := h.authUC.RefreshToken(r.Context(), req.RefreshToken, r.RemoteAddr)
	if err != nil {
		if errors.Is(err, domain.ErrInvalidToken) || errors.Is(err, domain.ErrUnauthorized) {
			utils.WriteError(w, http.StatusUnauthorized, "invalid or expired refresh token")
			return
		}
		utils.WriteError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	utils.WriteJSON(w, http.StatusOK, tokens, "token refreshed successfully")
}

func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	var req refreshRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err == nil && req.RefreshToken != "" {
		_ = h.authUC.Logout(r.Context(), req.RefreshToken, r.RemoteAddr)
	}

	utils.WriteJSON(w, http.StatusOK, nil, "session revoked successfully")
}

func (h *AuthHandler) JWKS(w http.ResponseWriter, r *http.Request) {
	jwks := h.keyManager.GetJWKS()
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Cache-Control", "public, max-age=3600")
	_ = json.NewEncoder(w).Encode(jwks)
}