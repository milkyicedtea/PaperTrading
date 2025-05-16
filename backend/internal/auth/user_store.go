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
)

// custom errors
var (
	ErrUserNotFound      = errors.New("user not found")
	ErrUserAlreadyExists = errors.New("user already exists")
)

type UserStore struct {
	db *pgxpool.Pool
}

func NewUserStore(db *pgxpool.Pool) *UserStore {
	if db == nil {
		log.Fatalf("Error: UserStore initialized with a nil DB pool.")
	}
	return &UserStore{db: db}
}

func (s *UserStore) CreateUserInDB(ctx context.Context, email string, passwordHash string) (*user.User, error) {
	query := `
		insert into public.users (email, password_hash) 
		values ($1, $2) returning id, email, password_hash, created_at, updated_at
	`
	var u user.User
	err := s.db.QueryRow(ctx, query, email, passwordHash).Scan(
		&u.ID,
		&u.Email,
		&u.PasswordHash,
		&u.CreatedAt,
		&u.UpdatedAt,
	)

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			// unique violation
			if pgErr.Code == "23505" {
				return nil, fmt.Errorf("user with email '%s' already exists: %w", email, ErrUserAlreadyExists)
			}
			log.Printf("Error finding user by email in DB: %v. Email: %s", err, email)
			return nil, fmt.Errorf("could not find user by email: %w", err)
		}
	}

	return &u, nil
}

// FindUserByEmailInDB retrieves a user by their email address
func (s *UserStore) FindUserByEmailInDB(ctx context.Context, email string) (*user.User, error) {
	query := `
		select id, email, password_hash, created_at, updated_at
		from public.users
		where email = $1
	`
	var u user.User
	err := s.db.QueryRow(ctx, query, email).Scan(
		&u.ID,
		&u.Email,
		&u.PasswordHash,
		&u.CreatedAt,
		&u.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		log.Printf("Error finding user by email in DB: %v. Email: %s", err, email)
		return nil, fmt.Errorf("could not find user by email: %w", err)
	}
	return &u, nil
}

// FindUserByIDInDB retrieves a user by their ID
func (s *UserStore) FindUserByIDInDB(ctx context.Context, userID uuid.UUID) (*user.User, error) {
	query := `
		select id, email, password_hash, created_at, updated_at
		from public.users
		where id = $1
	`
	var u user.User
	err := s.db.QueryRow(ctx, query, userID).Scan(
		&u.ID,
		&u.Email,
		&u.PasswordHash,
		&u.CreatedAt,
		&u.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		log.Printf("Error finding user by ID in DB: %v. ID: %s", err, userID)
		return nil, fmt.Errorf("could not find user by ID: %w", err)
	}
	return &u, nil
}
