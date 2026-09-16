package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"auth-service/internal/delivery/http/middleware"
	"auth-service/internal/domain"
	"auth-service/internal/usecase"
	"auth-service/pkg/utils"

	"github.com/google/uuid"
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

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	tenantID, _ := r.Context().Value(middleware.TenantContextKey).(uuid.UUID)
	err := h.authUC.Register(r.Context(), tenantID, req.Email, req.Password, r.RemoteAddr)
	if err != nil {
		if errors.Is(err, domain.ErrConflict) {
			utils.WriteError(w, http.StatusConflict, "user with this email already exists")
			return
		}
		utils.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	utils.WriteJSON(w, http.StatusCreated, nil, "user registered; please check your email for verification")
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	tenantID, _ := r.Context().Value(middleware.TenantContextKey).(uuid.UUID)
	tokens, err := h.authUC.Login(r.Context(), tenantID, req.Email, req.Password, r.RemoteAddr)
	if err != nil {
		if errors.Is(err, domain.ErrUnauthorized) {
			utils.WriteError(w, http.StatusUnauthorized, "invalid credentials")
			return
		}
		utils.WriteError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	utils.WriteJSON(w, http.StatusOK, tokens, "authentication response")
}

func (h *AuthHandler) VerifyMFA(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	var req struct {
		MFAToken string `json:"mfa_token"`
		Code     string `json:"code"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.MFAToken == "" || req.Code == "" {
		utils.WriteError(w, http.StatusBadRequest, "mfa_token and code are required")
		return
	}

	tokens, err := h.authUC.VerifyMFAChallenge(r.Context(), req.MFAToken, req.Code, r.RemoteAddr)
	if err != nil {
		utils.WriteError(w, http.StatusUnauthorized, err.Error())
		return
	}

	utils.WriteJSON(w, http.StatusOK, tokens, "two-factor authentication verified")
}

func (h *AuthHandler) SetupMFA(w http.ResponseWriter, r *http.Request) {
	claims, _ := r.Context().Value(middleware.UserContextKey).(*domain.JwtCustomClaims)
	secret, uri, err := h.authUC.SetupMFA(r.Context(), claims.UserID)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}

	utils.WriteJSON(w, http.StatusOK, map[string]string{
		"secret": secret,
		"uri":    uri,
	}, "scan the URI into your authenticator app")
}

func (h *AuthHandler) ConfirmMFASetup(w http.ResponseWriter, r *http.Request) {
	claims, _ := r.Context().Value(middleware.UserContextKey).(*domain.JwtCustomClaims)
	var req struct {
		Code string `json:"code"`
	}
	_ = json.NewDecoder(r.Body).Decode(&req)
	if err := h.authUC.ConfirmMFASetup(r.Context(), claims.UserID, req.Code); err != nil {
		utils.WriteError(w, http.StatusBadRequest, "invalid confirmation code")
		return
	}

	utils.WriteJSON(w, http.StatusOK, nil, "mfa successfully enrolled and activated")
}

func (h *AuthHandler) VerifyEmail(w http.ResponseWriter, r *http.Request) {
	token := r.URL.Query().Get("token")
	if token == "" {
		utils.WriteError(w, http.StatusBadRequest, "token query parameter required")
		return
	}
	if err := h.authUC.ConfirmEmailVerification(r.Context(), token); err != nil {
		utils.WriteError(w, http.StatusBadRequest, "invalid or expired token")
		return
	}
	utils.WriteJSON(w, http.StatusOK, nil, "email address successfully verified")
}

func (h *AuthHandler) PasswordResetRequest(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email string `json:"email"`
	}
	_ = json.NewDecoder(r.Body).Decode(&req)
	tenantID, _ := r.Context().Value(middleware.TenantContextKey).(uuid.UUID)
	_ = h.authUC.RequestPasswordReset(r.Context(), tenantID, req.Email)
	utils.WriteJSON(w, http.StatusOK, nil, "if the email exists, a password reset link has been dispatched")
}

func (h *AuthHandler) PasswordResetConfirm(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Token       string `json:"token"`
		NewPassword string `json:"new_password"`
	}
	_ = json.NewDecoder(r.Body).Decode(&req)
	if err := h.authUC.ConfirmPasswordReset(r.Context(), req.Token, req.NewPassword); err != nil {
		utils.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	utils.WriteJSON(w, http.StatusOK, nil, "password has been updated successfully")
}

func (h *AuthHandler) OAuthURL(w http.ResponseWriter, r *http.Request) {
	provider := r.URL.Query().Get("provider")
	url, state, err := h.authUC.OAuthLoginURL(r.Context(), provider)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	utils.WriteJSON(w, http.StatusOK, map[string]string{
		"url":   url,
		"state": state,
	}, "oauth authorization url generated successfully")
}

func (h *AuthHandler) OAuthCallback(w http.ResponseWriter, r *http.Request) {
	provider := r.URL.Query().Get("provider")
	code := r.URL.Query().Get("code")
	state := r.URL.Query().Get("state")
	tenantID, _ := r.Context().Value(middleware.TenantContextKey).(uuid.UUID)

	tokens, err := h.authUC.HandleOAuthCallback(r.Context(), tenantID, provider, code, state, r.RemoteAddr)
	if err != nil {
		utils.WriteError(w, http.StatusUnauthorized, err.Error())
		return
	}

	utils.WriteJSON(w, http.StatusOK, tokens, "oauth authentication successful")
}

func (h *AuthHandler) Refresh(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	var req struct {
		RefreshToken string `json:"refresh_token"`
	}
	_ = json.NewDecoder(r.Body).Decode(&req)
	tokens, err := h.authUC.RefreshToken(r.Context(), req.RefreshToken, r.RemoteAddr)
	if err != nil {
		utils.WriteError(w, http.StatusUnauthorized, "invalid or expired refresh token")
		return
	}
	utils.WriteJSON(w, http.StatusOK, tokens, "token refreshed")
}

func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	var req struct {
		RefreshToken string `json:"refresh_token"`
	}
	_ = json.NewDecoder(r.Body).Decode(&req)
	_ = h.authUC.Logout(r.Context(), req.RefreshToken, r.RemoteAddr)
	utils.WriteJSON(w, http.StatusOK, nil, "session revoked")
}

func (h *AuthHandler) JWKS(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	_ = json.NewEncoder(w).Encode(h.keyManager.GetJWKS())
}