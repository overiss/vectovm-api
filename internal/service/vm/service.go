package vm

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	cryptosvc "github.com/overiss/vectovm-api/internal/crypto"
	"github.com/overiss/vectovm-api/internal/domain"
	"github.com/overiss/vectovm-api/internal/model"
	"github.com/overiss/vectovm-api/internal/repository"
)

type Service struct {
	vms         *repository.VMRepository
	datanodes   *repository.DatanodeRepository
	users       *repository.UserRepository
	credentials *cryptosvc.CredentialService
}

func NewService(
	vms *repository.VMRepository,
	datanodes *repository.DatanodeRepository,
	users *repository.UserRepository,
	credentials *cryptosvc.CredentialService,
) *Service {
	return &Service{
		vms:         vms,
		datanodes:   datanodes,
		users:       users,
		credentials: credentials,
	}
}

func (s *Service) Create(ctx context.Context, oauthUserID uuid.UUID, req model.CreateVMRequest) (*model.VMResponse, error) {
	user, err := s.users.GetByOAuthUserID(ctx, oauthUserID)
	if err != nil {
		return nil, fmt.Errorf("resolve user: %w", err)
	}

	datanode, err := s.datanodes.GetByUserAndName(ctx, user.ID, req.DatanodeName)
	if err != nil {
		if errors.Is(err, repository.ErrDatanodeNotFound) {
			return nil, ErrDatanodeNotOwned
		}
		return nil, fmt.Errorf("resolve datanode: %w", err)
	}

	user, err = s.credentials.EnsureUserDEK(ctx, user)
	if err != nil {
		return nil, fmt.Errorf("ensure user dek: %w", err)
	}

	port := req.Port
	if port == 0 {
		port = 22
	}

	encryptedCredentials, err := s.credentials.EncryptVMCredentials(ctx, user, req.SSHUser, req.SSHPassword)
	if err != nil {
		return nil, fmt.Errorf("encrypt vm credentials: %w", err)
	}

	created, err := s.vms.Create(ctx, &domain.VM{
		UserID:               user.ID,
		DatanodeID:           datanode.ID,
		DatanodeName:         datanode.Name,
		Name:                 req.Name,
		Host:                 req.Host,
		Port:                 port,
		EncryptedCredentials: encryptedCredentials,
	})
	if err != nil {
		return nil, fmt.Errorf("persist vm: %w", err)
	}

	created.DatanodeName = datanode.Name
	return toVMResponse(created), nil
}

func (s *Service) List(ctx context.Context, oauthUserID uuid.UUID) ([]model.VMResponse, error) {
	user, err := s.users.GetByOAuthUserID(ctx, oauthUserID)
	if err != nil {
		return nil, fmt.Errorf("resolve user: %w", err)
	}

	vms, err := s.vms.ListByUserID(ctx, user.ID)
	if err != nil {
		return nil, fmt.Errorf("list vms: %w", err)
	}

	result := make([]model.VMResponse, 0, len(vms))
	for _, item := range vms {
		result = append(result, *toVMResponse(&item))
	}
	return result, nil
}

func (s *Service) Get(ctx context.Context, oauthUserID uuid.UUID, name string) (*model.VMResponse, error) {
	user, err := s.users.GetByOAuthUserID(ctx, oauthUserID)
	if err != nil {
		return nil, fmt.Errorf("resolve user: %w", err)
	}

	vm, err := s.vms.GetByUserAndName(ctx, user.ID, name)
	if err != nil {
		if errors.Is(err, repository.ErrVMNotFound) {
			return nil, err
		}
		return nil, fmt.Errorf("get vm: %w", err)
	}

	return toVMResponse(vm), nil
}

func toVMResponse(vm *domain.VM) *model.VMResponse {
	return &model.VMResponse{
		ID:           vm.ID.String(),
		Name:         vm.Name,
		DatanodeName: vm.DatanodeName,
		Host:         vm.Host,
		Port:         vm.Port,
		CreatedAt:    vm.CreatedAt,
	}
}
