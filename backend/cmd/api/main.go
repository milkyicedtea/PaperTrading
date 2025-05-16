package main

import (
	"backend/internal/auth"
	"backend/internal/config"
	"backend/internal/database"
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
)

type RegisterRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type AuthResponse struct {
	AccessToken string `json:"access_token"`
	// refresh token sent in HttpOnly cookie
	User auth.UserInfoForResponse `json:"user"`
}

type UserInfo struct {
	ID    int    `json:"id"`
	Email string `json:"email"`
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// very basic logger
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	log.Println("Logger initialized.")
	if cfg.LogLevel == "debug" {
		log.Println("Service starting with log level: DEBUG")
	}

	if err := database.InitPgxPool(ctx, cfg); err != nil {
		log.Fatalf("Failed to initialize PostgreSQL pool: %v", err)
	}
	// defer closing the pool when the application exits
	defer database.ClosePgxPool()

	// initialize pool and stores
	dbPool := database.GetPool()
	userStore := auth.NewUserStore(dbPool)
	tokenStore := auth.NewTokenStore(dbPool)

	// initialize authService
	authService := auth.NewAuthService(dbPool, userStore, tokenStore, cfg)

	// initialize authHandler
	authHandler := auth.NewHandler(authService, cfg)

	// initialize authMiddleware
	authMiddleware := auth.NewMiddleware(authService)

	r := chi.NewRouter()

	// Middleware
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(60 * time.Second))

	CORSMiddleware := cors.New(cors.Options{
		AllowedOrigins:   []string{"http://localhost:8001"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300, // maximum value not ignored by any major browsers
	})

	r.Use(CORSMiddleware.Handler)

	// public routes
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Welcome to PaperTrading API"))
	})

	// authentication routes
	r.Route("/api/auth", func(ar chi.Router) {
		ar.Post("/register", authHandler.Register)
		ar.Post("/login", authHandler.Login)
		ar.Post("/refresh-token", authHandler.RefreshToken)
		ar.Post("/logout", authHandler.Logout)
	})

	// Protected routes
	r.Group(func(protectedRouter chi.Router) {
		protectedRouter.Use(authMiddleware.Authenticate) // apply the auth middleware

		// get current user's info
		protectedRouter.Get("/api/me", func(w http.ResponseWriter, r *http.Request) {
			claims, ok := auth.GetUserClaims(r.Context())
			if !ok {
				// this should ideally not happen if middleware is working correctly
				// and has already validated, but as a safeguard:
				auth.RespondWithError(w, http.StatusUnauthorized, "Unable to retrieve user claims")
				return
			}
			// respond with the claims:
			auth.RespondWithJSON(w, http.StatusOK, map[string]interface{}{
				"message":   "Current user:",
				"userId":    claims.UserID,
				"email":     claims.Email,
				"expiresAt": claims.ExpiresAt.Time.Format(time.RFC3339),
			})
		})

		// TODO: other future protected routes:
		// protectedRouter.Get("/api/portfolio", portfolioHandler.GetPortfolio)
		// protectedRouter.Post("/api/trades", tradesHandler.CreateTrade)
	})

	server := &http.Server{
		Addr:    ":" + cfg.AppPort,
		Handler: r,
	}

	// graceful shutdown
	go func() {
		quit := make(chan os.Signal, 1)
		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
		<-quit
		log.Println("Shutting down server...")
		cancel() // signal context cancellation

		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer shutdownCancel()

		if err := server.Shutdown(shutdownCtx); err != nil {
			log.Fatalf("Server forced to shutdown: %v", err)
		}
	}()

	log.Printf("Server starting on port %s\n", cfg.AppPort)
	if err := server.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
		log.Fatalf("Failed to start server: %v", err)
	}

	log.Println("Server exited gracefully")
}
