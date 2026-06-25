package user

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/overiss/vectovm-api/internal/model"
	"github.com/overiss/vectovm-api/internal/repository"
)

type Service struct {
	users *repository.UserRepository
}

func NewService(users *repository.UserRepository) *Service {
	return &Service{users: users}
}

func (s *Service) GetByOAuthUserID(ctx context.Context, oauthUserID uuid.UUID) (*model.UserResponse, error) {
	user, err := s.users.GetByOAuthUserID(ctx, oauthUserID)
	if err != nil {
		return nil, fmt.Errorf("get user: %w", err)
	}

	return &model.UserResponse{
		ID:          user.ID.String(),
		OAuthUserID: user.OAuthUserID.String(),
		CreatedAt:   user.CreatedAt,
	}, nil
}
