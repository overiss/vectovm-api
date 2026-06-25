package datanode

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	mapiclient "github.com/overiss/vectovm-api/internal/client/mapi"
	"github.com/overiss/vectovm-api/internal/domain"
	"github.com/overiss/vectovm-api/internal/model"
	"github.com/overiss/vectovm-api/internal/repository"
)

type Service struct {
	mapi      *mapiclient.Client
	users     *repository.UserRepository
	datanodes *repository.DatanodeRepository
}

func NewService(
	mapi *mapiclient.Client,
	users *repository.UserRepository,
	datanodes *repository.DatanodeRepository,
) *Service {
	return &Service{
		mapi:      mapi,
		users:     users,
		datanodes: datanodes,
	}
}

func (s *Service) Create(ctx context.Context, oauthUserID uuid.UUID, req model.CreateDatanodeRequest) (*model.JobResponse, error) {
	user, err := s.users.GetByOAuthUserID(ctx, oauthUserID)
	if err != nil {
		return nil, fmt.Errorf("resolve user: %w", err)
	}

	port := req.Port
	if port == 0 {
		port = 22
	}

	job, err := s.mapi.CreateDatanode(ctx, mapiclient.CreateDatanodeRequest{
		UserID:   oauthUserID.String(),
		Name:     req.Name,
		Host:     req.Host,
		Port:     port,
		User:     req.User,
		Password: req.Password,
	})
	if err != nil {
		return nil, fmt.Errorf("create datanode in mapi: %w", err)
	}

	jobID := job.JobID
	_, err = s.datanodes.Create(ctx, &domain.Datanode{
		UserID:    user.ID,
		Name:      req.Name,
		Host:      req.Host,
		Port:      port,
		SSHUser:   req.User,
		LastJobID: &jobID,
	})
	if err != nil {
		return nil, fmt.Errorf("persist datanode: %w", err)
	}

	return &model.JobResponse{
		JobID:   job.JobID,
		Status:  job.Status,
		Message: job.Message,
	}, nil
}

func (s *Service) List(ctx context.Context, oauthUserID uuid.UUID) ([]model.DatanodeResponse, error) {
	user, err := s.users.GetByOAuthUserID(ctx, oauthUserID)
	if err != nil {
		return nil, fmt.Errorf("resolve user: %w", err)
	}

	nodes, err := s.datanodes.ListByUserID(ctx, user.ID)
	if err != nil {
		return nil, fmt.Errorf("list datanodes: %w", err)
	}

	result := make([]model.DatanodeResponse, 0, len(nodes))
	for _, node := range nodes {
		result = append(result, toDatanodeResponse(node))
	}
	return result, nil
}

func (s *Service) DeployVault(ctx context.Context, oauthUserID uuid.UUID, req model.DeployVaultRequest) (*model.JobResponse, error) {
	user, err := s.users.GetByOAuthUserID(ctx, oauthUserID)
	if err != nil {
		return nil, fmt.Errorf("resolve user: %w", err)
	}

	if _, err := s.datanodes.GetByUserAndName(ctx, user.ID, req.DatanodeName); err != nil {
		if errors.Is(err, repository.ErrDatanodeNotFound) {
			return nil, err
		}
		return nil, fmt.Errorf("get datanode: %w", err)
	}

	job, err := s.mapi.DeployVault(ctx, mapiclient.DeployVaultRequest{
		UserID:       oauthUserID.String(),
		DatanodeName: req.DatanodeName,
	})
	if err != nil {
		return nil, fmt.Errorf("deploy vault in mapi: %w", err)
	}

	return &model.JobResponse{
		JobID:   job.JobID,
		Status:  job.Status,
		Message: job.Message,
	}, nil
}

func (s *Service) GetJob(ctx context.Context, oauthUserID uuid.UUID, jobID string) (*model.JobStatusResponse, error) {
	if _, err := s.users.GetByOAuthUserID(ctx, oauthUserID); err != nil {
		return nil, fmt.Errorf("resolve user: %w", err)
	}

	resp, err := s.mapi.GetJob(ctx, jobID)
	if err != nil {
		return nil, fmt.Errorf("get job from mapi: %w", err)
	}

	return mapJobStatus(resp), nil
}

func (s *Service) GetRuntime(ctx context.Context, oauthUserID uuid.UUID, datanodeName string) (*model.RuntimeResponse, error) {
	user, err := s.users.GetByOAuthUserID(ctx, oauthUserID)
	if err != nil {
		return nil, fmt.Errorf("resolve user: %w", err)
	}

	if _, err := s.datanodes.GetByUserAndName(ctx, user.ID, datanodeName); err != nil {
		if errors.Is(err, repository.ErrDatanodeNotFound) {
			return nil, err
		}
		return nil, fmt.Errorf("get datanode: %w", err)
	}

	resp, err := s.mapi.GetRuntime(ctx, oauthUserID.String(), datanodeName)
	if err != nil {
		return nil, fmt.Errorf("get runtime from mapi: %w", err)
	}

	return &model.RuntimeResponse{
		UserID:       resp.UserID,
		DatanodeName: resp.DatanodeName,
		Datanode:     resp.Datanode,
		VaultStatus:  resp.VaultStatus,
		VaultLogs:    resp.VaultLogs,
	}, nil
}

func toDatanodeResponse(node domain.Datanode) model.DatanodeResponse {
	return model.DatanodeResponse{
		ID:        node.ID.String(),
		Name:      node.Name,
		Host:      node.Host,
		Port:      node.Port,
		SSHUser:   node.SSHUser,
		LastJobID: node.LastJobID,
		CreatedAt: node.CreatedAt,
	}
}

func mapJobStatus(resp *mapiclient.JobStatusResponse) *model.JobStatusResponse {
	if resp == nil || resp.Job == nil {
		return &model.JobStatusResponse{}
	}

	return &model.JobStatusResponse{
		Job: &model.Job{
			ID:         resp.Job.ID,
			Type:       resp.Job.Type,
			Status:     resp.Job.Status,
			Error:      resp.Job.Error,
			CreatedAt:  resp.Job.CreatedAt,
			FinishedAt: resp.Job.FinishedAt,
		},
	}
}
