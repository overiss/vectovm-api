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

var ErrVMNotFound = errors.New("vm not found")

type VMRepository struct {
	db *postgres.DB
}

func NewVMRepository(db *postgres.DB) *VMRepository {
	return &VMRepository{db: db}
}

func (r *VMRepository) Create(ctx context.Context, vm *domain.VM) (*domain.VM, error) {
	row := r.db.Pool().QueryRow(ctx, `
		INSERT INTO vms (user_id, datanode_id, name, host, port, encrypted_credentials)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, user_id, datanode_id, name, host, port, encrypted_credentials, created_at
	`, vm.UserID, vm.DatanodeID, vm.Name, vm.Host, vm.Port, vm.EncryptedCredentials)

	created, err := scanVM(row)
	if err != nil {
		return nil, fmt.Errorf("create vm: %w", err)
	}
	return created, nil
}

func (r *VMRepository) ListByUserID(ctx context.Context, userID uuid.UUID) ([]domain.VM, error) {
	rows, err := r.db.Pool().Query(ctx, `
		SELECT v.id, v.user_id, v.datanode_id, d.name, v.name, v.host, v.port, v.encrypted_credentials, v.created_at
		FROM vms v
		JOIN datanodes d ON d.id = v.datanode_id
		WHERE v.user_id = $1
		ORDER BY v.created_at DESC
	`, userID)
	if err != nil {
		return nil, fmt.Errorf("list vms: %w", err)
	}
	defer rows.Close()

	vms := make([]domain.VM, 0)
	for rows.Next() {
		vm, err := scanVMWithDatanodeName(rows)
		if err != nil {
			return nil, fmt.Errorf("scan vm: %w", err)
		}
		vms = append(vms, *vm)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate vms: %w", err)
	}
	return vms, nil
}

func (r *VMRepository) GetByUserAndName(ctx context.Context, userID uuid.UUID, name string) (*domain.VM, error) {
	row := r.db.Pool().QueryRow(ctx, `
		SELECT v.id, v.user_id, v.datanode_id, d.name, v.name, v.host, v.port, v.encrypted_credentials, v.created_at
		FROM vms v
		JOIN datanodes d ON d.id = v.datanode_id
		WHERE v.user_id = $1 AND v.name = $2
	`, userID, name)

	vm, err := scanVMWithDatanodeName(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrVMNotFound
		}
		return nil, fmt.Errorf("get vm: %w", err)
	}
	return vm, nil
}

func scanVM(row pgx.Row) (*domain.VM, error) {
	var vm domain.VM
	if err := row.Scan(
		&vm.ID, &vm.UserID, &vm.DatanodeID, &vm.Name, &vm.Host, &vm.Port,
		&vm.EncryptedCredentials, &vm.CreatedAt,
	); err != nil {
		return nil, err
	}
	return &vm, nil
}

func scanVMWithDatanodeName(row pgx.Row) (*domain.VM, error) {
	var vm domain.VM
	if err := row.Scan(
		&vm.ID, &vm.UserID, &vm.DatanodeID, &vm.DatanodeName, &vm.Name, &vm.Host, &vm.Port,
		&vm.EncryptedCredentials, &vm.CreatedAt,
	); err != nil {
		return nil, err
	}
	return &vm, nil
}
