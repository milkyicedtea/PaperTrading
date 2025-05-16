package auth

import (
	"backend/internal/config"
	"backend/internal/user"
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"
	"log"
	"strings"
	"time"
)

// authentication related custom errors
var (
	ErrInvalidCredentials = errors.New("invalid email or password")
	ErrTokenExpired       = errors.New("token has expired")
	ErrTokenNotValidYet   = errors.New("token not valid yet")
	ErrInvalidToken       = errors.New("invalid token")
)

// JWTCustomClaims defines the custom claims for our JWT.
// It embeds jwt.RegisteredClaims and adds our own.
type JWTCustomClaims struct {
	UserID uuid.UUID `json:"uid"`
	Email  string    `json:"email"`
	// TODO: add other claims like roles, permissions, etc.
	jwt.RegisteredClaims
}

// AuthService provides authentication related services.
type AuthService struct {
	db *pgxpool.Pool
	ts *TokenStore
	us *UserStore

	// individual jwt settings
	jwtSecret              string
	jwtExpiration          time.Duration
	refreshTokenExpiration time.Duration
}

func NewAuthService(db *pgxpool.Pool, us *UserStore, ts *TokenStore, cfg *config.Config) *AuthService {
	if cfg == nil {
		log.Fatal("AuthService: config cannot be nil")
	}
	if db == nil {
		log.Fatal("AuthService: database pool cannot be nil")
	}
	return &AuthService{
		db: db,
		us: us,
		ts: ts,

		jwtSecret:              cfg.JWTSecret,
		jwtExpiration:          cfg.JWTExpiration,
		refreshTokenExpiration: cfg.RefreshTokenExpiration,
	}
}

// --- Password Hashing

// HashPassword generates a bcrypt hash of the password.
func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("failed to hash password: %w", err)
	}
	return string(bytes), nil
}

// CheckPasswordHash compares a plain text password with a stored bcrypt hash.
func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// --- Token Generation

// GenerateAccessToken creates a new JWT access token for a user.
func (s *AuthService) GenerateAccessToken(u *user.User) (string, error) {
	if u == nil {
		return "", errors.New("user cannot be nil for token generation")
	}

	claims := &JWTCustomClaims{
		UserID: u.ID,
		Email:  u.Email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(s.jwtExpiration)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "PaperTradingApp", // identifier for our backend
			Subject:   u.ID.String(),     // subject of the token (user that's related to it)
			//ID: // TODO: JWT ID, can be used for tracking/revocation
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString([]byte(s.jwtSecret))
	if err != nil {
		log.Printf("Error signing access token for user %s: %v", u.Email, err)
		return "", fmt.Errorf("could not sign access token: %w", err)
	}
	return signedToken, nil
}

// GenerateRefreshToken creates a new refresh token.
// As of now, this is a simple JWT, though ideally refresh tokens are opaque strings
// stored in the DB and not self-contained JWTs for easy revocation.
// This will be refined later. For simplicity we start with a JWT.
func (s *AuthService) GenerateRefreshToken(u *user.User) (string, error) {
	if u == nil {
		return "", errors.New("user cannot be nil for refresh token generation")
	}

	claims := &JWTCustomClaims{ // refresh token can have simpler claims
		UserID: u.ID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(s.refreshTokenExpiration)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "PaperTradingApp-Refresh",
			Subject:   u.ID.String(),
			// ID: // JWT ID, could be used to link to a stored refresh token record later on
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	// TODO: use different secrets for generating refresh and access
	signedToken, err := token.SignedString([]byte(s.jwtSecret)) // use the same secret for now since it's jwt
	if err != nil {
		log.Printf("Error signing refresh token for user %s: %v", u.Email, err)
		return "", fmt.Errorf("could not sign refresh token: %w", err)
	}
	return signedToken, nil
}

// ValidateToken parses and validates a JWT token string and
// returns the custom claims if the token is valid.
func (s *AuthService) ValidateToken(tokenString string) (*JWTCustomClaims, error) {
	// remove "Bearer " prefix if present
	tokenString = strings.TrimPrefix(tokenString, "Bearer ")

	token, err := jwt.ParseWithClaims(tokenString, &JWTCustomClaims{}, func(token *jwt.Token) (interface{}, error) {
		// validate the used alg is what you expect:
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(s.jwtSecret), nil
	})

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrTokenExpired
		}
		if errors.Is(err, jwt.ErrTokenNotValidYet) {
			return nil, ErrTokenNotValidYet
		}
		// other errors like ErrSignatureInvalid, ErrTokenMalformed etc.
		log.Printf("Error parsing or validating token: %v", err)
		return nil, ErrInvalidToken
	}

	if claims, ok := token.Claims.(*JWTCustomClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, ErrInvalidToken
}

func generateOpaqueTokenString() (string, error) {
	numBytes := 32
	b := make([]byte, numBytes)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("failed to generate random bytes for opaque token: %w", err)
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

func hashToken(token string) string {
	hasher := sha256.New()
	hasher.Write([]byte(token))
	return base64.URLEncoding.EncodeToString(hasher.Sum(nil))
}

// --- Registration

// RegisterUserInput defines the input for user registration.
type RegisterUserInput struct {
	Email    string
	Password string
}

// RegisterUser handles new user registration.
func (s *AuthService) RegisterUser(ctx context.Context, input RegisterUserInput) (*user.User, error) {
	// 1. basic input validation
	if input.Email == "" || input.Password == "" {
		return nil, errors.New("email and password are required")
	}
	// TODO: add robust email validation and password complexity rules
	if len(input.Password) < 8 { // Example: minimum password length
		return nil, errors.New("password must be at least 8 characters long")
	}

	// 2. check if user already exists
	existingUser, err := s.us.FindUserByEmailInDB(ctx, input.Email)
	if err != nil && !errors.Is(err, ErrUserNotFound) {
		log.Printf("Error checking for existing user during registration: %v", err)
		return nil, fmt.Errorf("could not verify user existence: %w", err)
	}
	if existingUser != nil {
		return nil, ErrUserAlreadyExists
	}

	// 3. hash password
	hashedPassword, err := HashPassword(input.Password)
	if err != nil {
		log.Printf("Error hashing password during registration for email %s: %v", input.Email, err)
		return nil, fmt.Errorf("could not process password: %w", err)
	}

	// 4. create user in the database
	newUser, err := s.us.CreateUserInDB(ctx, input.Email, hashedPassword)
	if err != nil {
		return nil, fmt.Errorf("could not register user: %w", err)
	}

	log.Printf("User registered successfully: %s (ID: %s)", newUser.Email, newUser.ID)
	// don't return the password hash in the user object sent back to handler response
	// the user.User struct has `json:"-"` for PasswordHash so it won't error out
	return newUser, nil
}

// --- Login

// LoginUserInput defines the input for user login.
type LoginUserInput struct {
	Email    string
	Password string
}

// LoginUserResponse defines the successful login response.
type LoginUserResponse struct {
	AccessToken  string
	RefreshToken string // this is sent in an HttpOnly cookie from the handler
	User         UserInfoForResponse
}

// LoginUser handles user login.
func (s *AuthService) LoginUser(ctx context.Context, input LoginUserInput) (*LoginUserResponse, error) {
	// 1. validate input
	if input.Email == "" || input.Password == "" {
		return nil, errors.New("email and password are required for login")
	}

	// 2. find user by email
	u, err := s.us.FindUserByEmailInDB(ctx, input.Email)
	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			return nil, ErrInvalidCredentials // Generic error for security
		}
		log.Printf("Error finding user during login for email %s: %v", input.Email, err)
		return nil, fmt.Errorf("login attempt failed: %w", err)
	}

	// 3. check password
	if !CheckPasswordHash(input.Password, u.PasswordHash) {
		return nil, ErrInvalidCredentials // generic error for security
	}

	// 4. generate tokens
	accessToken, err := s.GenerateAccessToken(u)
	if err != nil {
		return nil, fmt.Errorf("could not generate access token: %w", err)
	}

	opaqueRefreshToken, err := generateOpaqueTokenString()
	if err != nil {
		return nil, fmt.Errorf("failed to generate opaque refresh token: %w", err)
	}
	refreshTokenHash := hashToken(opaqueRefreshToken)
	refreshTokenExpiresAt := time.Now().Add(s.refreshTokenExpiration)

	if err := s.ts.SaveRefreshToken(ctx, u.ID, refreshTokenHash, refreshTokenExpiresAt); err != nil {
		return nil, fmt.Errorf("failed to save refresh token for login: %w", err)
	}

	log.Printf("User logged in successfully: %s (ID: %s)", u.Email, u.ID)

	// important: the user.User struct has PasswordHash tagged with `json:"-"`.
	// this means that when this LoginUserResponse is marshalled to json by the handler,
	// the password hash nor a related field will not be included.
	return &LoginUserResponse{
		AccessToken:  accessToken,
		RefreshToken: opaqueRefreshToken,
		User:         ToUserInfoForResponse(u),
	}, nil
}

type UserInfoForResponse struct {
	ID    uuid.UUID `json:"id"`
	Email string    `json:"email"`
}

func ToUserInfoForResponse(u *user.User) UserInfoForResponse {
	if u == nil {
		return UserInfoForResponse{} // Or handle as an error/panic depending on context
	}
	return UserInfoForResponse{
		ID:    u.ID,
		Email: u.Email,
	}
}

// RefreshTokenResponse defines the response for a successful token refresh.
type RefreshTokenResponse struct {
	AccessToken  string
	RefreshToken string
}

// ProcessRefreshToken validates an existing refresh token and issues new tokens.
func (s *AuthService) ProcessRefreshToken(ctx context.Context, oldOpaqueRefreshTokenString string) (*RefreshTokenResponse, error) {
	if oldOpaqueRefreshTokenString == "" {
		return nil, ErrInvalidToken // or a more specific "refresh token missing" error
	}

	// 1. Hash the incoming opaque refresh token string.
	oldTokenHash := hashToken(oldOpaqueRefreshTokenString)

	// 2. validate the old token hash against the DB and fetch the user.
	// ValidateAndFetchUserByTokenHash (from token_store.go) checks expiry too.
	u, err := s.ts.ValidateAndFetchUserByTokenHash(ctx, oldTokenHash)
	if err != nil {
		// err could be ErrRefreshTokenNotFound from token_store.go
		log.Printf("Opaque refresh token validation failed: %v (token was %s...)", err, oldOpaqueRefreshTokenString[:minhashes(len(oldOpaqueRefreshTokenString), 10)])
		return nil, ErrInvalidToken // Return a generic error to the client
	}

	// 3. if valid, delete the old refresh token from DB (strict rotation).
	// this makes the old token unusable immediately.
	if err := s.ts.DeleteRefreshTokenByHash(ctx, oldTokenHash); err != nil {
		// log this error since it could be critical.
		// if deletion fails, the old token might still be valid if the client didn't discard it,
		// though the client *should* replace it with the new one.
		log.Printf("WARNING: Failed to delete old refresh token hash %s after validation for user %s: %v", oldTokenHash[:minhashes(len(oldTokenHash), 10)], u.ID, err)
		// consider proceeding or returning an error. For now, we'll proceed if user was validated.
	}

	// 4. generate a new access token.
	newAccessToken, err := s.GenerateAccessToken(u)
	if err != nil {
		return nil, fmt.Errorf("could not generate new access token during refresh: %w", err)
	}

	// 5. generate a new opaque refresh token.
	newOpaqueRefreshToken, err := generateOpaqueTokenString()
	if err != nil {
		return nil, fmt.Errorf("failed to generate new opaque refresh token: %w", err)
	}
	newRefreshTokenHash := hashToken(newOpaqueRefreshToken)
	newRefreshTokenExpiresAt := time.Now().Add(s.refreshTokenExpiration)

	// 6. save hash of new opaque refresh token to the DB.
	if err := s.ts.SaveRefreshToken(ctx, u.ID, newRefreshTokenHash, newRefreshTokenExpiresAt); err != nil {
		// if saving the new token fails, this is a critical issue.
		// the user might be left in a state where they can't refresh again with the new token.
		log.Printf("CRITICAL: Failed to save new refresh token for user %s during refresh: %v", u.ID, err)
		// forcing re-login by returning an error is safer if the refresh mechanism is broken.
		return nil, fmt.Errorf("failed to save new refresh token: %w", err)
	}

	log.Printf("Tokens refreshed successfully using opaque token for user: %s (ID: %s). New opaque refresh token issued.", u.Email, u.ID)
	return &RefreshTokenResponse{
		AccessToken:  newAccessToken,
		RefreshToken: newOpaqueRefreshToken, // return raw opaque token for the cookie
	}, nil
}
