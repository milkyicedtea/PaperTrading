package auth

import (
	"backend/internal/user"
	"context"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"log"
	"time"
)

var (
	// ErrRefreshTokenNotFound is returned when a refresh token is not found in the store.
	ErrRefreshTokenNotFound = errors.New("refresh token not found")
)

type TokenStore struct {
	db *pgxpool.Pool
}

func NewTokenStore(db *pgxpool.Pool) *TokenStore {
	if db == nil {
		log.Fatalf("Error: TokenStore initialized with a nil DB pool.")
	}
	return &TokenStore{db: db}
}

func (s *TokenStore) SaveRefreshToken(ctx context.Context, userID uuid.UUID, tokenHash string, expiresAt time.Time) error {
	query := `
		INSERT INTO refresh_tokens (user_id, token_hash, expires_at)
		VALUES ($1, $2, $3)
	`
	_, err := s.db.Exec(ctx, query, userID, tokenHash, expiresAt)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" { // unique_violation
			log.Printf("Error saving refresh token: unique constraint violation for token_hash. UserID: %s", userID)
			return fmt.Errorf("failed to save refresh token due to conflict: %w", err)
		}
		log.Printf("Error saving refresh token to DB for user %s: %v", userID, err)
		return fmt.Errorf("failed to save refresh token: %w", err)
	}
	return nil
}

// ValidateAndFetchUserByTokenHash finds a refresh token by its hash, checks if it's valid (not expired),
// and returns the associated user's User object.
func (s *TokenStore) ValidateAndFetchUserByTokenHash(ctx context.Context, tokenHash string) (*user.User, error) {
	query := `
		SELECT u.id, u.email, u.password_hash, u.created_at, u.updated_at
		FROM refresh_tokens rt
		JOIN users u ON rt.user_id = u.id
		WHERE rt.token_hash = $1 AND rt.expires_at > NOW()
	`
	var u user.User
	err := s.db.QueryRow(ctx, query, tokenHash).Scan(
		&u.ID,
		&u.Email,
		&u.PasswordHash,
		&u.CreatedAt,
		&u.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			// This means token not found OR found but expired.
			return nil, ErrRefreshTokenNotFound
		}
		log.Printf("Error fetching user by refresh token hash: %v (hash was %s...)", err, tokenHash[:minhashes(len(tokenHash), 10)])
		return nil, fmt.Errorf("error validating refresh token from DB: %w", err)
	}
	return &u, nil
}

// DeleteRefreshTokenByHash deletes a specific refresh token by its hash.
func (s *TokenStore) DeleteRefreshTokenByHash(ctx context.Context, tokenHash string) error {
	query := `DELETE FROM refresh_tokens WHERE token_hash = $1`
	commandTag, err := s.db.Exec(ctx, query, tokenHash)
	if err != nil {
		log.Printf("Error deleting refresh token hash from DB: %v (hash was %s...)", err, tokenHash[:minhashes(len(tokenHash), 10)])
		return fmt.Errorf("failed to delete refresh token from DB: %w", err)
	}
	if commandTag.RowsAffected() == 0 {
		log.Printf("Attempted to delete refresh token hash %s..., but it was not found.", tokenHash[:minhashes(len(tokenHash), 10)])
	}
	return nil
}

// DeleteUserRefreshTokens deletes all refresh tokens associated with a specific user ID.
func (s *TokenStore) DeleteUserRefreshTokens(ctx context.Context, userID uuid.UUID) error {
	query := `DELETE FROM refresh_tokens WHERE user_id = $1`
	commandTag, err := s.db.Exec(ctx, query, userID)
	if err != nil {
		log.Printf("Error deleting refresh tokens for user %s from DB: %v", userID, err)
		return fmt.Errorf("failed to delete user's refresh tokens: %w", err)
	}
	log.Printf("Deleted %d refresh token(s) for user %s", commandTag.RowsAffected(), userID)
	return nil
}

// DeleteExpiredTokens manually deletes all expired refresh tokens from the database.
func (s *TokenStore) DeleteExpiredTokens(ctx context.Context) (int64, error) {
	query := `DELETE FROM refresh_tokens WHERE expires_at <= NOW()`
	commandTag, err := s.db.Exec(ctx, query)
	if err != nil {
		log.Printf("Error deleting expired refresh tokens from DB: %v", err)
		return 0, fmt.Errorf("failed to delete expired refresh tokens: %w", err)
	}
	log.Printf("Successfully deleted %d expired refresh token(s).", commandTag.RowsAffected())
	return commandTag.RowsAffected(), nil
}

// Helper for logging token hashes safely
func minhashes(a, b int) int {
	if a < b {
		return a
	}
	return b
}
