package crypto

import (
	"context"
	"fmt"

	"github.com/overiss/vectovm-api/internal/domain"
	"github.com/overiss/vectovm-api/internal/repository"
)

type CredentialService struct {
	envelope *Envelope
	users    *repository.UserRepository
}

func NewCredentialService(envelope *Envelope, users *repository.UserRepository) *CredentialService {
	return &CredentialService{
		envelope: envelope,
		users:    users,
	}
}

func (s *CredentialService) GenerateUserDEK() ([]byte, error) {
	_, encryptedDEK, err := s.envelope.GenerateUserDEK()
	if err != nil {
		return nil, fmt.Errorf("generate user dek: %w", err)
	}
	return encryptedDEK, nil
}

func (s *CredentialService) EncryptVMCredentials(
	ctx context.Context,
	user *domain.User,
	sshUser, sshPassword string,
) ([]byte, error) {
	dek, err := s.userDEK(ctx, user)
	if err != nil {
		return nil, err
	}

	return s.envelope.EncryptCredentials(dek, SSMCredentials{
		User:     sshUser,
		Password: sshPassword,
	})
}

func (s *CredentialService) DecryptVMCredentials(ctx context.Context, user *domain.User, encrypted []byte) (SSMCredentials, error) {
	dek, err := s.userDEK(ctx, user)
	if err != nil {
		return SSMCredentials{}, err
	}
	return s.envelope.DecryptCredentials(dek, encrypted)
}

func (s *CredentialService) EnsureUserDEK(ctx context.Context, user *domain.User) (*domain.User, error) {
	if len(user.EncryptedDEK) > 0 {
		return user, nil
	}

	encryptedDEK, err := s.GenerateUserDEK()
	if err != nil {
		return nil, err
	}

	if err := s.users.SetEncryptedDEK(ctx, user.ID, encryptedDEK); err != nil {
		return nil, fmt.Errorf("persist user dek: %w", err)
	}

	user.EncryptedDEK = encryptedDEK
	return user, nil
}

func (s *CredentialService) userDEK(ctx context.Context, user *domain.User) ([]byte, error) {
	updated, err := s.EnsureUserDEK(ctx, user)
	if err != nil {
		return nil, err
	}

	if len(updated.EncryptedDEK) == 0 {
		return nil, repository.ErrUserDEKMissing
	}

	dek, err := s.envelope.UnwrapDEK(updated.EncryptedDEK)
	if err != nil {
		return nil, fmt.Errorf("unwrap user dek: %w", err)
	}
	return dek, nil
}
