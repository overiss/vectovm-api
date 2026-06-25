package vault

import (
	"context"
	"encoding/base64"
	"fmt"

	vaultapi "github.com/hashicorp/vault/api"
)

type KeyStore struct {
	client       *vaultapi.Client
	credentialsKeyPath string
}

func NewKeyStore(address, token, credentialsKeyPath string) (*KeyStore, error) {
	cfg := vaultapi.DefaultConfig()
	cfg.Address = address

	client, err := vaultapi.NewClient(cfg)
	if err != nil {
		return nil, fmt.Errorf("create vault client: %w", err)
	}
	client.SetToken(token)

	return &KeyStore{
		client:             client,
		credentialsKeyPath: credentialsKeyPath,
	}, nil
}

func (k *KeyStore) CredentialsMasterKey(ctx context.Context) (string, error) {
	_ = ctx
	data, err := k.readSecret(k.credentialsKeyPath)
	if err != nil {
		return "", err
	}

	masterKey, ok := data["master_key"].(string)
	if !ok || masterKey == "" {
		return "", fmt.Errorf("credentials master_key not found in vault")
	}

	if _, err := base64.StdEncoding.DecodeString(masterKey); err != nil {
		return "", fmt.Errorf("credentials master_key is not valid base64: %w", err)
	}

	return masterKey, nil
}

func (k *KeyStore) readSecret(path string) (map[string]interface{}, error) {
	secret, err := k.client.Logical().Read(path)
	if err != nil {
		return nil, fmt.Errorf("read vault secret %s: %w", path, err)
	}
	if secret == nil || secret.Data == nil {
		return nil, fmt.Errorf("vault secret %s not found", path)
	}

	data, ok := secret.Data["data"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("vault secret %s has unexpected format", path)
	}

	return data, nil
}
