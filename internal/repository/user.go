package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/overiss/vectovm-api/internal/domain"
	"github.com/overiss/vectovm-api/internal/storage/postgres"
)

var ErrUserNotFound = errors.New("user not found")
var ErrUserDEKMissing = errors.New("user dek missing")

type UserRepository struct {
	db *postgres.DB
}

func NewUserRepository(db *postgres.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) Create(ctx context.Context, oauthUserID uuid.UUID, encryptedDEK []byte) (*domain.User, error) {
	row := r.db.Pool().QueryRow(ctx, `
		INSERT INTO users (oauth_user_id, encrypted_dek)
		VALUES ($1, $2)
		ON CONFLICT (oauth_user_id) DO UPDATE
			SET encrypted_dek = COALESCE(users.encrypted_dek, EXCLUDED.encrypted_dek)
		RETURNING id, oauth_user_id, encrypted_dek, created_at
	`, oauthUserID, encryptedDEK)

	user, err := scanUser(row)
	if err != nil {
		return nil, fmt.Errorf("create user: %w", err)
	}
	return user, nil
}

func (r *UserRepository) SetEncryptedDEK(ctx context.Context, userID uuid.UUID, encryptedDEK []byte) error {
	tag, err := r.db.Pool().Exec(ctx, `
		UPDATE users SET encrypted_dek = $2 WHERE id = $1
	`, userID, encryptedDEK)
	if err != nil {
		return fmt.Errorf("set user dek: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrUserNotFound
	}
	return nil
}

func (r *UserRepository) GetByOAuthUserID(ctx context.Context, oauthUserID uuid.UUID) (*domain.User, error) {
	row := r.db.Pool().QueryRow(ctx, `
		SELECT id, oauth_user_id, encrypted_dek, created_at
		FROM users
		WHERE oauth_user_id = $1
	`, oauthUserID)

	user, err := scanUser(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("get user by oauth id: %w", err)
	}
	return user, nil
}

func scanUser(row pgx.Row) (*domain.User, error) {
	var user domain.User
	if err := row.Scan(&user.ID, &user.OAuthUserID, &user.EncryptedDEK, &user.CreatedAt); err != nil {
		return nil, err
	}
	return &user, nil
}
