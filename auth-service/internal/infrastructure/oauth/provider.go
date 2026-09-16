package oauth

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

type UserInfo struct {
	Subject string
	Email   string
}

type Provider interface {
	GetAuthURL(state string) string
	ExchangeCode(ctx context.Context, code string) (*UserInfo, error)
}

type googleProvider struct {
	clientID     string
	clientSecret string
	redirectURI  string
}

func NewGoogleProvider() Provider {
	return &googleProvider{
		clientID:     os.Getenv("GOOGLE_CLIENT_ID"),
		clientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
		redirectURI:  os.Getenv("GOOGLE_REDIRECT_URI"),
	}
}

func (g *googleProvider) GetAuthURL(state string) string {
	return fmt.Sprintf("https://accounts.google.com/o/oauth2/v2/auth?client_id=%s&redirect_uri=%s&response_type=code&scope=openid%%20email&state=%s",
		g.clientID, url.QueryEscape(g.redirectURI), state,
	)
}

func (g *googleProvider) ExchangeCode(ctx context.Context, code string) (*UserInfo, error) {
	data := url.Values{
		"code":          {code},
		"client_id":     {g.clientID},
		"client_secret": {g.clientSecret},
		"redirect_uri":  {g.redirectURI},
		"grant_type":    {"authorization_code"},
	}

	req, _ := http.NewRequestWithContext(ctx, "POST", "https://oauth2.googleapis.com/token", strings.NewReader(data.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil || resp.StatusCode != http.StatusOK {
		return nil, errors.New("failed to exchange code with google")
	}
	defer resp.Body.Close()

	var tokenRes struct {
		AccessToken string `json:"access_token"`
	}
	_ = json.NewDecoder(resp.Body).Decode(&tokenRes)

	reqUser, _ := http.NewRequestWithContext(ctx, "GET", "https://www.googleapis.com/oauth2/v2/userinfo", nil)
	reqUser.Header.Set("Authorization", "Bearer "+tokenRes.AccessToken)

	userResp, err := client.Do(reqUser)
	if err != nil || userResp.StatusCode != http.StatusOK {
		return nil, errors.New("failed to fetch userinfo from google")
	}
	defer userResp.Body.Close()

	var userInfo struct {
		ID    string `json:"id"`
		Email string `json:"email"`
	}
	_ = json.NewDecoder(userResp.Body).Decode(&userInfo)

	return &UserInfo{Subject: userInfo.ID, Email: userInfo.Email}, nil
}

type githubProvider struct {
	clientID     string
	clientSecret string
	redirectURI  string
}

func NewGitHubProvider() Provider {
	return &githubProvider{
		clientID:     os.Getenv("GITHUB_CLIENT_ID"),
		clientSecret: os.Getenv("GITHUB_CLIENT_SECRET"),
		redirectURI:  os.Getenv("GITHUB_REDIRECT_URI"),
	}
}

func (gh *githubProvider) GetAuthURL(state string) string {
	return fmt.Sprintf("https://github.com/login/oauth/authorize?client_id=%s&redirect_uri=%s&scope=user:email&state=%s",
		gh.clientID, url.QueryEscape(gh.redirectURI), state,
	)
}

func (gh *githubProvider) ExchangeCode(ctx context.Context, code string) (*UserInfo, error) {
	data := url.Values{
		"code":          {code},
		"client_id":     {gh.clientID},
		"client_secret": {gh.clientSecret},
		"redirect_uri":  {gh.redirectURI},
	}

	req, _ := http.NewRequestWithContext(ctx, "POST", "https://github.com/login/oauth/access_token", strings.NewReader(data.Encode()))
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil || resp.StatusCode != http.StatusOK {
		return nil, errors.New("failed to exchange code with github")
	}
	defer resp.Body.Close()

	var tokenRes struct {
		AccessToken string `json:"access_token"`
	}
	_ = json.NewDecoder(resp.Body).Decode(&tokenRes)

	reqUser, _ := http.NewRequestWithContext(ctx, "GET", "https://api.github.com/user", nil)
	reqUser.Header.Set("Authorization", "Bearer "+tokenRes.AccessToken)

	userResp, err := client.Do(reqUser)
	if err != nil || userResp.StatusCode != http.StatusOK {
		return nil, errors.New("failed to fetch userinfo from github")
	}
	defer userResp.Body.Close()

	var userInfo struct {
		ID    int64  `json:"id"`
		Email string `json:"email"`
	}
	_ = json.NewDecoder(userResp.Body).Decode(&userInfo)

	// In case primary email is private in profile, fallback to synthetic or verified primary email
	if userInfo.Email == "" {
		userInfo.Email = fmt.Sprintf("gh_%d@users.noreply.github.com", userInfo.ID)
	}

	return &UserInfo{Subject: fmt.Sprintf("%d", userInfo.ID), Email: userInfo.Email}, nil
}