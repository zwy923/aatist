package handler

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/aatist/backend/internal/user/service"
	"github.com/aatist/backend/pkg/errs"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/idtoken"
)

func (h *AuthHandler) googleOAuthReady() bool {
	return strings.TrimSpace(h.googleOAuth.ClientID) != "" &&
		strings.TrimSpace(h.googleOAuth.ClientSecret) != "" &&
		strings.TrimSpace(h.googleOAuth.RedirectURI) != ""
}

func (h *AuthHandler) googleOAuthConfig() *oauth2.Config {
	return &oauth2.Config{
		ClientID:     h.googleOAuth.ClientID,
		ClientSecret: h.googleOAuth.ClientSecret,
		RedirectURL:  h.googleOAuth.RedirectURI,
		Scopes:       []string{"openid", "email", "profile"},
		Endpoint:     google.Endpoint,
	}
}

// GoogleOAuthStartHandler redirects the browser to Google's consent screen.
func (h *AuthHandler) GoogleOAuthStartHandler(c *gin.Context) {
	if !h.googleOAuthReady() {
		h.respondError(c, http.StatusServiceUnavailable, errs.ErrInvalidInput, "Google sign-in is not configured")
		return
	}
	stateBytes := make([]byte, 24)
	if _, err := rand.Read(stateBytes); err != nil {
		h.logger.Error("Google OAuth: failed to generate state", zap.Error(err))
		h.respondError(c, http.StatusInternalServerError, err, "failed to start Google sign-in")
		return
	}
	state := hex.EncodeToString(stateBytes)
	if err := h.authService.SaveGoogleOAuthState(c.Request.Context(), state); err != nil {
		h.logger.Error("Google OAuth: failed to save state", zap.Error(err))
		h.respondError(c, http.StatusInternalServerError, err, "failed to start Google sign-in")
		return
	}
	url := h.googleOAuthConfig().AuthCodeURL(state, oauth2.AccessTypeOnline)
	c.Redirect(http.StatusFound, url)
}

// GoogleOAuthCallbackHandler receives the redirect from Google, exchanges the code, and redirects to the SPA with tokens in the URL fragment.
func (h *AuthHandler) GoogleOAuthCallbackHandler(c *gin.Context) {
	if !h.googleOAuthReady() {
		h.redirectOAuthError(c, "not_configured")
		return
	}
	if errParam := strings.TrimSpace(c.Query("error")); errParam != "" {
		h.logger.Warn("Google OAuth provider error", zap.String("error", errParam))
		h.redirectOAuthError(c, "denied")
		return
	}
	state := c.Query("state")
	if err := h.authService.ConsumeGoogleOAuthState(c.Request.Context(), state); err != nil {
		h.logger.Warn("Google OAuth invalid state", zap.Error(err))
		h.redirectOAuthError(c, "state_invalid")
		return
	}
	code := c.Query("code")
	if code == "" {
		h.redirectOAuthError(c, "missing_code")
		return
	}

	ctx := c.Request.Context()
	tok, err := h.googleOAuthConfig().Exchange(ctx, code)
	if err != nil {
		h.logger.Warn("Google OAuth token exchange failed", zap.Error(err))
		h.redirectOAuthError(c, "token_exchange_failed")
		return
	}
	rawID, _ := tok.Extra("id_token").(string)
	if strings.TrimSpace(rawID) == "" {
		h.logger.Warn("Google OAuth: id_token missing from token response")
		h.redirectOAuthError(c, "no_id_token")
		return
	}
	payload, err := idtoken.Validate(ctx, rawID, h.googleOAuth.ClientID)
	if err != nil {
		h.logger.Warn("Google OAuth id_token validation failed", zap.Error(err))
		h.redirectOAuthError(c, "invalid_id_token")
		return
	}

	prof := service.GoogleOAuthProfile{
		Subject:       payload.Subject,
		Email:         claimString(payload.Claims, "email"),
		EmailVerified: claimBool(payload.Claims, "email_verified"),
		Name:          claimString(payload.Claims, "name"),
		HostedDomain:  claimString(payload.Claims, "hd"),
	}

	_, tokens, err := h.authService.RegisterOrLoginGoogle(ctx, prof, h.getClientIP(c))
	if err != nil {
		if appErr, ok := err.(*errs.AppError); ok && appErr.StatusCode == 409 {
			h.redirectOAuthError(c, "email_conflict")
			return
		}
		h.logger.Error("Google OAuth register/login failed", zap.Error(err))
		h.redirectOAuthError(c, "server_error")
		return
	}

	h.redirectOAuthSuccess(c, tokens)
}

func claimString(claims map[string]interface{}, key string) string {
	if claims == nil {
		return ""
	}
	v, ok := claims[key]
	if !ok || v == nil {
		return ""
	}
	switch t := v.(type) {
	case string:
		return strings.TrimSpace(t)
	default:
		return strings.TrimSpace(fmt.Sprint(t))
	}
}

func claimBool(claims map[string]interface{}, key string) bool {
	if claims == nil {
		return false
	}
	v, ok := claims[key]
	if !ok || v == nil {
		return false
	}
	if b, ok := v.(bool); ok {
		return b
	}
	return false
}

func (h *AuthHandler) redirectOAuthError(c *gin.Context, code string) {
	base := h.oauthFrontendBase
	if base == "" {
		base = "http://localhost:5173"
	}
	u, err := url.Parse(base + "/auth/oauth/google")
	if err != nil {
		c.String(http.StatusBadRequest, "invalid oauth redirect configuration")
		return
	}
	q := u.Query()
	q.Set("error", code)
	u.RawQuery = q.Encode()
	c.Redirect(http.StatusFound, u.String())
}

func (h *AuthHandler) redirectOAuthSuccess(c *gin.Context, tokens *service.Tokens) {
	base := h.oauthFrontendBase
	if base == "" {
		base = "http://localhost:5173"
	}
	u, err := url.Parse(base + "/auth/oauth/google")
	if err != nil {
		c.String(http.StatusBadRequest, "invalid oauth redirect configuration")
		return
	}
	frag := url.Values{}
	frag.Set("access_token", tokens.AccessToken)
	frag.Set("refresh_token", tokens.RefreshToken)
	u.Fragment = frag.Encode()
	c.Redirect(http.StatusFound, u.String())
}
