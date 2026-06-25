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

var ErrDatanodeNotFound = errors.New("datanode not found")

type DatanodeRepository struct {
	db *postgres.DB
}

func NewDatanodeRepository(db *postgres.DB) *DatanodeRepository {
	return &DatanodeRepository{db: db}
}

func (r *DatanodeRepository) Create(ctx context.Context, node *domain.Datanode) (*domain.Datanode, error) {
	row := r.db.Pool().QueryRow(ctx, `
		INSERT INTO datanodes (user_id, name, host, port, ssh_user, last_job_id)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, user_id, name, host, port, ssh_user, last_job_id, created_at
	`, node.UserID, node.Name, node.Host, node.Port, node.SSHUser, node.LastJobID)

	created, err := scanDatanode(row)
	if err != nil {
		return nil, fmt.Errorf("create datanode: %w", err)
	}
	return created, nil
}

func (r *DatanodeRepository) ListByUserID(ctx context.Context, userID uuid.UUID) ([]domain.Datanode, error) {
	rows, err := r.db.Pool().Query(ctx, `
		SELECT id, user_id, name, host, port, ssh_user, last_job_id, created_at
		FROM datanodes
		WHERE user_id = $1
		ORDER BY created_at DESC
	`, userID)
	if err != nil {
		return nil, fmt.Errorf("list datanodes: %w", err)
	}
	defer rows.Close()

	nodes := make([]domain.Datanode, 0)
	for rows.Next() {
		node, err := scanDatanode(rows)
		if err != nil {
			return nil, fmt.Errorf("scan datanode: %w", err)
		}
		nodes = append(nodes, *node)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate datanodes: %w", err)
	}
	return nodes, nil
}

func (r *DatanodeRepository) GetByUserAndName(ctx context.Context, userID uuid.UUID, name string) (*domain.Datanode, error) {
	row := r.db.Pool().QueryRow(ctx, `
		SELECT id, user_id, name, host, port, ssh_user, last_job_id, created_at
		FROM datanodes
		WHERE user_id = $1 AND name = $2
	`, userID, name)

	node, err := scanDatanode(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrDatanodeNotFound
		}
		return nil, fmt.Errorf("get datanode: %w", err)
	}
	return node, nil
}

func (r *DatanodeRepository) UpdateLastJobID(ctx context.Context, id uuid.UUID, jobID string) error {
	tag, err := r.db.Pool().Exec(ctx, `
		UPDATE datanodes SET last_job_id = $2 WHERE id = $1
	`, id, jobID)
	if err != nil {
		return fmt.Errorf("update datanode job id: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrDatanodeNotFound
	}
	return nil
}

func scanDatanode(row pgx.Row) (*domain.Datanode, error) {
	var node domain.Datanode
	if err := row.Scan(
		&node.ID, &node.UserID, &node.Name, &node.Host, &node.Port,
		&node.SSHUser, &node.LastJobID, &node.CreatedAt,
	); err != nil {
		return nil, err
	}
	return &node, nil
}
