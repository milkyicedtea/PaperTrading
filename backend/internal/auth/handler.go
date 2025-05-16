package auth

import (
	"backend/internal/config"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strings"
	"time"
)

// Handler holds dependencies for authentication HTTP handlers.
type Handler struct {
	service *AuthService
	cfg     *config.Config
}

// NewHandler creates a new authentication handler.
func NewHandler(service *AuthService, cfg *config.Config) *Handler {
	if service == nil {
		log.Fatal("Auth Handler: AuthService cannot be nil")
	}
	if cfg == nil {
		log.Fatal("Auth Handler: Config cannot be nil")
	}
	return &Handler{service: service, cfg: cfg}
}

// --- Request/Response

type RegisterUserRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	// TODO: add more field e.g. first name, surname, ecc.
}

type LoginUserRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// AuthResponse is used for successful authentication responses.
type AuthResponse struct {
	AccessToken string              `json:"accessToken"`
	User        UserInfoForResponse `json:"user"` // using the struct defined in service.go
}

// --- Helper Functions for HTTP responses

func RespondWithError(w http.ResponseWriter, code int, message string) {
	RespondWithJSON(w, code, map[string]string{"error": message})
}

func RespondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	response, err := json.Marshal(payload)
	if err != nil {
		log.Printf("Error marshalling JSON response: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error":"Failed to marshal JSON response"}`)) // Fallback
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(response)
}

// --- HTTP Handlers

// Register handles user registration requests.
// POST /api/auth/register
func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	var req RegisterUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		RespondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	defer r.Body.Close()

	if req.Email == "" || req.Password == "" {
		RespondWithError(w, http.StatusBadRequest, "Email and password are required")
		return
	}

	// map handler to service input
	serviceInput := RegisterUserInput{
		Email:    req.Email,
		Password: req.Password,
	}

	newUser, err := h.service.RegisterUser(r.Context(), serviceInput)
	if err != nil {
		log.Printf("Registration error for email %s: %v", req.Email, err)
		if errors.Is(err, ErrUserAlreadyExists) {
			RespondWithError(w, http.StatusConflict, "User with this email already exists")
		} else if strings.Contains(err.Error(), "password must be at least") {
			RespondWithError(w, http.StatusBadRequest, err.Error())
		} else {
			RespondWithError(w, http.StatusInternalServerError, "Failed to register user")
		}
		return
	}

	// prepare response
	// the newUser from service is *user.User. it needs to be mapped to UserInfoForResponse.
	responseUser := ToUserInfoForResponse(newUser)

	// for registration, don't log the user in immediately or issue tokens.
	// the user is expected to log in separately.
	log.Printf("User registered via handler: %s (ID: %s)", responseUser.Email, responseUser.ID)
	RespondWithJSON(w, http.StatusCreated, responseUser) // return user info, no tokens
}

// Login handles user login requests.
// POST /api/auth/login
func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var req LoginUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		RespondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	defer r.Body.Close()

	if req.Email == "" || req.Password == "" {
		RespondWithError(w, http.StatusBadRequest, "Email and password are required")
		return
	}

	serviceInput := LoginUserInput{
		Email:    req.Email,
		Password: req.Password,
	}

	loginResponse, err := h.service.LoginUser(r.Context(), serviceInput)
	if err != nil {
		log.Printf("Login error for email %s: %v", req.Email, err)
		if errors.Is(err, ErrInvalidCredentials) {
			RespondWithError(w, http.StatusUnauthorized, "Invalid email or password")
		} else {
			RespondWithError(w, http.StatusInternalServerError, "Failed to log in")
		}
		return
	}

	// set refresh token in HttpOnly cookie
	// the cookie's secure attribute should be true if served over HTTPS.
	// for local development on HTTP it may need to be false or the browser will ignore it.
	// checks like h.cfg.AppEnv == "production" (needs to be added) or similar are good here.
	secureCookie := h.cfg.AppEnv == "production" // simple check for "are we in a prod env?"
	// another way would be to have an explicit DOMAIN in config.

	http.SetCookie(w, &http.Cookie{
		Name:     "refreshToken",
		Value:    loginResponse.RefreshToken,
		Path:     "/api/auth", // scope cookie to auth paths for refresh/logout
		Expires:  time.Now().Add(h.service.refreshTokenExpiration),
		HttpOnly: true,
		Secure:   secureCookie, // true in production (HTTPS), false in development (HTTP)
		SameSite: http.SameSiteStrictMode,
		// Domain: h.cfg.CookieDomain, // should be set if api and frontend are on different subdomains
	})

	// prepare response (access token in body, user info)
	apiResponse := AuthResponse{
		AccessToken: loginResponse.AccessToken,
		User:        loginResponse.User,
	}

	log.Printf("User logged in via handler: %s (ID: %s)", apiResponse.User.Email, apiResponse.User.ID)
	RespondWithJSON(w, http.StatusOK, apiResponse)
}

// RefreshToken handles requests to refresh authentication tokens.
// POST /api/auth/refresh-token
func (h *Handler) RefreshToken(w http.ResponseWriter, r *http.Request) {
	// 1. get refresh token from HttpOnly cookie
	cookie, err := r.Cookie("refreshToken")
	if err != nil {
		if errors.Is(err, http.ErrNoCookie) {
			RespondWithError(w, http.StatusUnauthorized, "Refresh token cookie not found")
			return
		}
		log.Printf("Error reading refresh token cookie: %v", err)
		RespondWithError(w, http.StatusBadRequest, "Could not process request") // Or StatusInternalServerError
		return
	}
	oldRefreshTokenString := cookie.Value

	if oldRefreshTokenString == "" {
		RespondWithError(w, http.StatusUnauthorized, "Refresh token is empty")
		return
	}

	// 2. call service to process refresh token and get new tokens
	refreshResponse, err := h.service.ProcessRefreshToken(r.Context(), oldRefreshTokenString)
	if err != nil {
		// ProcessRefreshToken returns ErrInvalidToken for most failures (expired, not found, etc.)
		log.Printf("Failed to refresh token: %v", err)
		if errors.Is(err, ErrInvalidToken) { // generic error from service for bad refresh tokens
			RespondWithError(w, http.StatusUnauthorized, "Invalid or expired refresh token")
		} else {
			RespondWithError(w, http.StatusInternalServerError, "Could not refresh token")
		}
		return
	}

	// 3. set new refresh token in HttpOnly cookie
	secureCookie := h.cfg.AppEnv == "production" // same logic as in Login
	http.SetCookie(w, &http.Cookie{
		Name:     "refreshToken",
		Value:    refreshResponse.RefreshToken, // new refresh token
		Path:     "/api/auth",
		Expires:  time.Now().Add(h.service.refreshTokenExpiration),
		HttpOnly: true,
		Secure:   secureCookie,
		SameSite: http.SameSiteStrictMode,
		// Domain: h.cfg.CookieDomain,
	})

	// 4. send new access token in the response body
	responsePayload := map[string]string{
		"accessToken": refreshResponse.AccessToken,
	}
	RespondWithJSON(w, http.StatusOK, responsePayload)
}

// Logout handles user logout requests
// POST /api/auth/logout
func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	// clear refreshToken cookie by setting MaxAge to -1 or Expires to a past time.
	// the browser will then delete the cookie.

	// check if the cookie exists, though not strictly necessary
	// as setting an expired cookie with the same name will clear it.
	_, err := r.Cookie("refreshToken")
	if err != nil {
		if errors.Is(err, http.ErrNoCookie) {
			// no cookie to clear, user might already be logged out
			// or never had a session.
			RespondWithJSON(w, http.StatusOK, map[string]string{"message": "No active session to logout or already logged out"})
			return
		}
		// some other error reading the cookie, though unlikely to be critical for logout.
		log.Printf("Error reading cookie during logout (non-critical): %v", err)
	}

	secureCookie := h.cfg.DBSslMode != "disable" // Same logic as in Login/Refresh

	// to delete it, either:
	//	 1 - Expires can be set to a past time (like epoch time time.Unix(0, 0))
	//   2 - MaxAge can be set to -1
	// for no reason in particular we'll use MaxAge to -1
	http.SetCookie(w, &http.Cookie{
		Name:     "refreshToken",
		Value:    "", // value can be empty
		Path:     "/api/auth",
		MaxAge:   -1, // tell browser to delete immediately
		HttpOnly: true,
		Secure:   secureCookie,
		SameSite: http.SameSiteStrictMode,
		// Domain: h.cfg.CookieDomain,
	})

	log.Println("User logout: refreshToken cookie cleared.")
	RespondWithJSON(w, http.StatusOK, map[string]string{"message": "Successfully logged out"})
}
